package api

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"math"
	"net/http"
)

/*

GET /ping
POST /launch -> {"id": "<uuid>", "test-cases": {"test-case-1": "id1", "test-case-2": "id2", ...}}
WS /events
PUT /terminate
GET /status

*/

const (
	pingPath      = "/ping"
	launchPath    = "/launch"
	terminatePath = "/terminate"
	eventsPath    = "/events"
	messageType   = 19
)

func Mux(exit chan bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(pingPath, ping)
	mux.HandleFunc(launchPath, launch)
	mux.HandleFunc(terminatePath, terminate)
	mux.HandleFunc(eventsPath, events(exit))
	return mux
}

var (
	launchesQueue  = make(chan Launch, math.MaxUint32)
	terminateQueue = make(chan string, math.MaxUint32)
	eventsQueue    = make(chan Event, math.MaxUint32)
	upgrader       = websocket.Upgrader{}
)

func ping(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Ok\n"))
}

func launch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var launch Launch
	err := json.NewDecoder(r.Body).Decode(&launch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("A launch object is expected"))
		return
	}
	launchesQueue <- launch
}

func terminate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var uuids []string
	err := json.NewDecoder(r.Body).Decode(&uuids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("An array or test case IDs is expected"))
		return
	}
	for _, uuid := range uuids {
		terminateQueue <- uuid
	}
}

func events(exit chan bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Websocket upgrade error: %v\n", err)
			return
		}
		defer c.Close()
		for {
			select {
			case <-exit:
				return
			case event := <-eventsQueue:
				{
					data, err := json.Marshal(event)
					if err != nil {
						log.Printf("Event serialization error: %v\n", err)
						break
					}
					err = c.WriteMessage(messageType, data)
					if err != nil {
						log.Printf("Websocket output error: %v\n", err)
						break
					}
				}
			}
		}
	}
}
