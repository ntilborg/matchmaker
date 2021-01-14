package matchmaker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	allocation "agones.dev/agones/pkg/apis/allocation/v1"
)

// ServerFinder struct hold required data
type ServerFinder struct {
	mutex      sync.Mutex
	agonesPort int
	agonesHost string
	fleetName  string
	servers    map[uint32]*allocation.GameServerAllocation
}

// AgonesOption struct define engine option configuration
type AgonesOption struct {
	Port      int
	Host      string
	FleetName string
}

// NewFinder function return ServerFinder struct
func NewFinder(opt AgonesOption) *ServerFinder {
	fmt.Println("Agones Host:", opt.Host)
	fmt.Println("Agones Port:", opt.Port)
	return &ServerFinder{
		agonesPort: opt.Port,
		agonesHost: opt.Host,
		fleetName:  opt.FleetName,
		servers:    make(map[uint32]*allocation.GameServerAllocation),
	}
}

// GetServer get game server struct
func (s *ServerFinder) GetServer(poolID uint32, ch chan<- *allocation.GameServerAllocation) {
	if _, ok := s.servers[poolID]; !ok {

		///
		gs := new(allocation.GameServerAllocation)

		jsonErr := s.getJSON(s.agonesHost, s.agonesPort, s.fleetName, gs)
		if jsonErr != nil {
			fmt.Printf("Error with JSON %s\n", jsonErr)
			ch <- nil
			return
		}

		if gs == nil {
			fmt.Printf("Cannot find server")
			ch <- nil
			return
		}

		s.servers[poolID] = gs
	}

	ch <- s.servers[poolID]
}

func (s *ServerFinder) getJSON(host string, port int, fleetname string, alloc *allocation.GameServerAllocation) error {
	type MatchLabels struct {
		AgonesDevFleet string `json:"agones.dev/fleet"`
	}
	type Required struct {
		MatchLabels MatchLabels `json:"matchLabels"`
	}

	type Spec struct {
		Required Required `json:"required"`
	}

	type Payload struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Spec       Spec   `json:"spec"`
	}

	data := Payload{
		// fill struct
		APIVersion: "allocation.agones.dev/v1",
		Kind:       "GameServerAllocation",
		Spec: Spec{
			Required: Required{
				MatchLabels: MatchLabels{
					AgonesDevFleet: fleetname,
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/apis/allocation.agones.dev/v1/namespaces/default/gameserverallocations", host, port), body)
	if err != nil {
		// handle err
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
		return err
	}
	defer resp.Body.Close()

	//IO reading of the body
	newBody, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(newBody))

	//Read JSON
	return json.Unmarshal(newBody, &alloc)
}
