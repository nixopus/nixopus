package types

import (
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Domain struct {
	bun.BaseModel `bun:"table:domains,alias:do" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	UserID        uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`
	// ServerID      uuid.UUID  `json:"server_id" bun:"server_id,notnull,type:uuid"` // enable this when we have multiple server architecture, to keep things simple removing this
	Name          string     `json:"name" bun:"name,notnull"`
}

type Server struct {
	bun.BaseModel `bun:"table:servers,alias:s" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	UserID        uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`
	Name          string     `json:"name" bun:"name,notnull"`
	IP            string     `json:"ip" bun:"ip,notnull"`
	Hostname      string     `json:"hostname" bun:"hostname,notnull"`
}

func GetDefaultServer() Server {
	ip := getHostIP()

	hostname, _ := os.Hostname()

	return Server{
		ID:        uuid.UUID{},
		UserID:    uuid.UUID{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
		Name:      "Default",
		IP:        ip,
		Hostname:  hostname,
	}
}


// this logic has to be rechecked for when app is running inside a docker container
func getHostIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String()
		}
	}

	return ""
}
