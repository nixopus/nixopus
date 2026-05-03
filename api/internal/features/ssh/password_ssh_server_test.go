package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	cryptossh "golang.org/x/crypto/ssh"
)

func mustRSASigner(t *testing.T) cryptossh.Signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := cryptossh.NewSignerFromKey(key)
	require.NoError(t, err)
	return signer
}

func startPasswordEchoSSHServer(t *testing.T, user, password string) (addr string, shutdown func()) {
	t.Helper()
	hostSigner := mustRSASigner(t)
	cfg := &cryptossh.ServerConfig{
		PasswordCallback: func(conn cryptossh.ConnMetadata, pass []byte) (*cryptossh.Permissions, error) {
			if conn.User() == user && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("denied")
		},
	}
	cfg.AddHostKey(hostSigner)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleTestSSHConn(conn, cfg)
		}
	}()

	return ln.Addr().String(), func() { _ = ln.Close() }
}

func handleTestSSHConn(c net.Conn, cfg *cryptossh.ServerConfig) {
	defer c.Close()
	serverConn, chans, reqs, err := cryptossh.NewServerConn(c, cfg)
	if err != nil {
		return
	}
	defer func() { _ = serverConn.Close() }()
	go cryptossh.DiscardRequests(reqs)
	handleSessionChannels(chans)
}

func handleSessionChannels(chans <-chan cryptossh.NewChannel) {
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(cryptossh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go func(ch cryptossh.Channel, reqs <-chan *cryptossh.Request) {
			defer ch.Close()
			for req := range reqs {
				switch req.Type {
				case "exec":
					_ = req.Reply(true, nil)
					cmd := ""
					p := req.Payload
					if len(p) >= 4 {
						n := binary.BigEndian.Uint32(p[:4])
						if len(p) >= int(4+n) {
							cmd = string(p[4 : 4+n])
						}
					}
					_, _ = ch.Write([]byte(cmd))
					_, _ = ch.SendRequest("exit-status", false, cryptossh.Marshal(struct{ Status uint32 }{0}))
					return
				default:
					if req.WantReply {
						_ = req.Reply(false, nil)
					}
				}
			}
		}(channel, requests)
	}
}

func dialPasswordGoph(t *testing.T, addr, user, password string) (*goph.Client, error) {
	t.Helper()
	host, ps, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(ps)
	require.NoError(t, err)
	s := &SSH{
		User:     user,
		Host:     host,
		Port:     uint(port),
		Password: password,
	}
	return s.ConnectWithPassword()
}

func testSSHOrgContext(parent context.Context, orgID uuid.UUID) context.Context {
	return context.WithValue(context.WithValue(parent, types.OrganizationIDKey, orgID.String()), types.ServerIDKey, "")
}
