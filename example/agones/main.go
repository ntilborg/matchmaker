package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ntilborg/matchmaker"

	allocation "agones.dev/agones/pkg/apis/allocation/v1"
)

//Configuration from the config.json configuration file
type Configuration struct {
	ListeningPort int    `json:"MatchmakingPort"`
	Port          int    `json:"AllocatorPort"`
	Host          string `json:"AllocatorHost"`
	MaxPlayers    int    `json:"MaxPlayers"`
	WaitTime      int    `json:"WaitTime"`
	FleetName     string `json:"FleetName"`
}

//MatchResponse structure
type MatchResponse struct {
	MatchID        uint32 `json:"MatchmakingID"`
	IsFull         bool   `json:"IsFull"`
	CurrentPlayers uint32 `json:"CurrentPlayers"`
	MaxPlayers     uint32 `json:"MaxPlayers"`
	Port           int32  `json:"ServerPort"`
	Host           string `json:"ServerHost"`
}

var (
	m    *matchmaker.MatchMaker
	s    *matchmaker.ServerFinder
	conf Configuration

	alloc *allocation.GameServerAllocation
)

func main() {
	fmt.Println("Starting matchmaker")
	conf = readConfig("config.json")

	m = matchmaker.New(matchmaker.Option{
		MaxPlayers: conf.MaxPlayers,
		WaitTime:   time.Duration(conf.WaitTime) * time.Second,
	})

	s = matchmaker.NewFinder(matchmaker.AgonesOption{
		Port:      conf.Port,
		Host:      conf.Host,
		FleetName: conf.FleetName,
	})

	//Handler to register a new player. Returns unique player ID
	http.HandleFunc("/register", handleRegisterPlayer)

	//Handler to let a new player search and join a new match. Returns match MatchResponse
	http.HandleFunc("/join", handleJoinMatch)

	//Handler to poll if match has been found. Returns MatchResponse
	http.HandleFunc("/match", handleMatchStatus)

	// Run the HTTP server using the bound certificate and key for TLS
	fmt.Println(fmt.Sprintf("Start listening on port: %d", conf.ListeningPort))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.ListeningPort), nil); err != nil {
		fmt.Println("HTTPS server failed to run")
	}
}

// Register player
func handleRegisterPlayer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	uID := uuid.New().ID()
	_, err := io.WriteString(w, fmt.Sprintf("%d", uID))
	if err != nil {
		fmt.Println("Error registering player", uID)
	}
}

// Join with certain match ID. join?id=XXX
func handleJoinMatch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Println("Error joining match")
		return
	}

	fmt.Println("New player joining:", id)

	//Join the matchmaker with the unique client ID
	pool := m.Join(uint32(id))

	if pool == nil {
		fmt.Println("Error joining match")
		return
	}

	replyPool(w, r, pool)
}

// Join with certain uid
func handleMatchStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)
		_, err = io.WriteString(w, "ERROR")
		fmt.Println("Error getting match info")
		return
	}

	poolResp := m.GetPool(uint32(id))

	if poolResp == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)
		_, err = io.WriteString(w, "ERROR")

		fmt.Println("Match not found or empty")
		return
	}

	replyPool(w, r, poolResp)
}

func replyPool(w http.ResponseWriter, r *http.Request, pool *matchmaker.PoolResp) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var matchresponse []byte
	var err error

	if pool.IsFull {
		if pool.Gs == nil {

			ch := make(chan *allocation.GameServerAllocation)
			go s.GetServer(pool.PoolID, ch)
			pool.Gs = <-ch

			var playerString string = "Pool is full. Players joining: "

			for _, player := range pool.Players {
				playerString += fmt.Sprintf("%d ", player)
			}

			fmt.Println(playerString)
		}

		//Reply to the client server
		if pool.Gs == nil {
			println("Error finding server")
			matchresponse, err = json.Marshal(MatchResponse{MatchID: pool.PoolID, IsFull: pool.IsFull, CurrentPlayers: uint32(len(pool.Players)), MaxPlayers: uint32(conf.MaxPlayers)})
		} else {
			matchresponse, err = json.Marshal(MatchResponse{MatchID: pool.PoolID, IsFull: pool.IsFull, CurrentPlayers: uint32(len(pool.Players)), MaxPlayers: uint32(conf.MaxPlayers), Port: pool.Gs.Status.Ports[0].Port, Host: pool.Gs.Status.Address})
		}
	} else {
		//Pool not full
		matchresponse, err = json.Marshal(MatchResponse{MatchID: pool.PoolID, IsFull: pool.IsFull, CurrentPlayers: uint32(len(pool.Players)), MaxPlayers: uint32(conf.MaxPlayers)})
	}

	_, err = io.WriteString(w, fmt.Sprintf(string(matchresponse)))
	if err != nil {
		fmt.Println("Error joining match")
		return
	}
}

//Read the config file
func readConfig(filename string) Configuration {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened JSONfile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// Initialize Configuration struct var
	var conf Configuration

	// Create the struct
	json.Unmarshal(byteValue, &conf)

	return conf
}
