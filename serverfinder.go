package matchmaker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ServerFinder struct hold required data
type ServerFinder struct {
	mutex      sync.Mutex
	agonesPort int
	agonesHost string
	client     *http.Client
	servers    map[uint32]*AgonesGameServer
}

// AgonesOption struct define engine option configuration
type AgonesOption struct {
	Port int
	Host string
}

// AgonesGameServer struct define game server result
type AgonesGameServer struct {
	Port int
	Host string
}

// NewFinder function return ServerFinder struct
func NewFinder(opt AgonesOption) *ServerFinder {
	fmt.Println("Agones Host:", opt.Host)
	fmt.Println("Agones Port:", opt.Port)
	return &ServerFinder{
		agonesPort: opt.Port,
		agonesHost: opt.Host,
		client:     &http.Client{Timeout: 10 * time.Second},
		servers:    make(map[uint32]*AgonesGameServer),
	}

}

// GetServer get game server struct
func (s *ServerFinder) GetServer(poolID uint32) *AgonesGameServer {
	if _, ok := s.servers[poolID]; !ok {

		gs := new(AgonesGameServer)

		s.getJSON(fmt.Sprintf(s.agonesHost, "/address"), gs)
		println("Found", gs.Host, "port", gs.Port)

		s.servers[poolID] = gs
		return gs
	} else {
		return s.servers[poolID]
	}

}

func (s *ServerFinder) getJSON(url string, target interface{}) error {
	r, err := s.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
