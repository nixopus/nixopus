package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"strings"
	"context"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	"github.com/raghavyuva/nixopus-api/internal/types"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/uptrace/bun"
)

const (
	maxMessageSize = 10 * 1024 * 1024
	pingInterval   = 10 * time.Second
	pingTimeout    = 60 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SocketServer struct {
	conns            *sync.Map 
	shutdown         chan struct{}
	deployController *deploy.DeployController
	db               *bun.DB
	ctx              context.Context
}

func NewSocketServer(deployController *deploy.DeployController, db *bun.DB, ctx context.Context) (*SocketServer, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	server := &SocketServer{
		conns:            &sync.Map{},
		shutdown:         make(chan struct{}),
		deployController: deployController,
		db:               db,
		ctx:              ctx,
	}

	return server, nil
}


func (s *SocketServer) verifyToken(tokenString string) (*types.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return types.JWTSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token expired")
			}
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}
		
		userStorage := user_storage.UserStorage{
			DB:  s.db,
			Ctx: s.ctx,
		}
		
		user, err := userStorage.FindUserByEmail(email)
		if err != nil {
			return nil, err
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *SocketServer) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	
	fmt.Printf("New connection from client: %s\n", conn.RemoteAddr())
	conn.SetReadLimit(maxMessageSize)
	
	if token != "" {
		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}
		
		user, err := s.verifyToken(token)
		if err != nil {
			log.Printf("Auth error: %v", err)
			s.sendError(conn, "Invalid authorization token")
			conn.Close()
			return
		}
		
		log.Printf("User authenticated via URL token. ID: %s, Email: %s", user.ID, user.Email)
		s.conns.Store(conn, user.ID)
		defer s.handleDisconnect(conn)
		s.readLoop(conn, user)
		return
	}
	
	s.conns.Store(conn, "")
	defer s.handleDisconnect(conn)
	
	authTimer := time.NewTimer(30 * time.Second)
	authChan := make(chan *types.User)
	
	go func() {
		select {
		case user := <-authChan:
			if user != nil {
				s.readLoop(conn, user)
			}
		case <-authTimer.C:
			log.Printf("Authentication timeout for client: %s", conn.RemoteAddr())
			s.sendError(conn, "Authentication timeout")
			conn.Close()
		}
	}()
	
	s.waitForAuth(conn, authChan)
}

func (s *SocketServer) waitForAuth(conn *websocket.Conn, authChan chan<- *types.User) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("Error reading auth message: %v\n", err)
		s.sendError(conn, "Failed to read authentication message")
		authChan <- nil
		return
	}

	var msg types.Payload
	if err := json.Unmarshal(message, &msg); err != nil {
		fmt.Printf("Error unmarshaling auth message: %v\n", err)
		s.sendError(conn, "Invalid authentication message format")
		authChan <- nil
		return
	}

	if msg.Action != "authenticate" {
		s.sendError(conn, "First message must be authentication")
		authChan <- nil
		return
	}

	token, ok := msg.Data.(string)
	if !ok {
		s.sendError(conn, "Invalid authentication token format")
		authChan <- nil
		return
	}

	user, err := s.verifyToken(token)
	if err != nil {
		log.Printf("Auth error: %v", err)
		s.sendError(conn, "Invalid authorization token")
		authChan <- nil
		return
	}

	s.conns.Store(conn, user.ID)
	log.Printf("User authenticated. ID: %s, Email: %s", user.ID, user.Email)
	
	conn.WriteJSON(types.Payload{
		Action: "authenticated",
		Data:   user.ID,
	})
	authChan <- user
}

func (s *SocketServer) readLoop(conn *websocket.Conn, user *types.User) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			s.sendError(conn, "Failed to read message")
			return
		}

		var msg types.Payload
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			s.sendError(conn, "Invalid message format")
			continue
		}

		switch msg.Action {
		case "ping":
			s.handlePing(conn)
		case "deploy":
			if user == nil {
				s.sendError(conn, "Authentication required")
				continue
			}
			s.deployController.HandleDeploy(conn, msg.Data,user)
		case "authenticate":
			token, ok := msg.Data.(string)
			if !ok {
				s.sendError(conn, "Invalid authentication token format")
				continue
			}
			
			newUser, err := s.verifyToken(token)
			if err != nil {
				s.sendError(conn, "Invalid authorization token")
				continue
			}
			
			user = newUser
			s.conns.Store(conn, user.ID)
			
			conn.WriteJSON(types.Payload{
				Action: "authenticated",
				Data:   user.ID,
			})
			
			log.Printf("User re-authenticated. ID: %s, Email: %s", user.ID, user.Email)
		default:
			s.sendError(conn, "Unknown message action")
		}
	}
}

func (s *SocketServer) handleDisconnect(conn *websocket.Conn) {
	userID, _ := s.conns.Load(conn)
	s.conns.Delete(conn)
	conn.Close()
	fmt.Printf("Client disconnected: %s (User ID: %v)\n", conn.RemoteAddr(), userID)
}

func (s *SocketServer) Shutdown() {
	close(s.shutdown)
	s.conns.Range(func(conn, _ interface{}) bool {
		conn.(*websocket.Conn).Close()
		return true
	})
}

func (s *SocketServer) handlePing(conn *websocket.Conn) {
	userID, _ := s.conns.Load(conn)
	fmt.Printf("Received ping from %s (User ID: %v)\n", conn.RemoteAddr(), userID)
}

func (s *SocketServer) sendError(conn *websocket.Conn, message string) {
	conn.WriteJSON(types.Payload{
		Action: "error",
		Data:   message,
	})
}