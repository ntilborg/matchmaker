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
)

//Configuration from the config.json configuration file
type Configuration struct {
	Port       int    `json:"AllocatorPort"`
	Host       string `json:"AllocatorHost"`
	MaxPlayers int    `json:"MaxPlayers"`
	WaitTime   int    `json:"WaitTime"`
}

const (
	numberOfPlayerInRoom = 2
)

var (
	m *matchmaker.MatchMaker
)

func main() {
	fmt.Println("Starting matchmaker example")
	conf := readConfig("config.json")

	fmt.Println("Host:", conf.Host)
	fmt.Println("Port:", conf.Port)

	m = matchmaker.New(matchmaker.Option{
		MaxPlayers: conf.MaxPlayers,
		WaitTime:   time.Duration(conf.WaitTime) * time.Second,
	})

	//Handler to register a new player. Returns unique player ID
	http.HandleFunc("/register", handleRegisterPlayer)

	//Handler to let a new player search and join a new match. Returns match ID.
	http.HandleFunc("/join", handleJoinMatch)

	//Handler to poll if match has been found. Returns game configuration (IP and Port)
	http.HandleFunc("/match", handleMatchStatus)

	// Run the HTTP server using the bound certificate and key for TLS
	if err := http.ListenAndServe(":8001", nil); err != nil {
		fmt.Println("HTTPS server failed to run")
	} else {
		fmt.Println("HTTPS server is running on port 8001")
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
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("New player joining")

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Println("Error joining match")
		return
	}

	w.WriteHeader(http.StatusOK)

	//Join the matchmaker with the unique client ID
	pool := m.Join(int32(id))

	_, wrErr := io.WriteString(w, pool.PoolID)
	if wrErr != nil {

		fmt.Println("Error joining match")
		return
	}
}

// Join with certain uid
func handleMatchStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "TODO")
	if err != nil {
		fmt.Println("Error match status")
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
