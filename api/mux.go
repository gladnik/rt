package api

import (
	"encoding/json"
	"fmt"
	. "github.com/aerokube/rt/common"
	"github.com/aerokube/rt/event"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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

var (
	launchesQueue  = make(chan string)
	terminateQueue = make(chan string)
	eventBus       = event.NewEventBus()
	upgrader       = websocket.Upgrader{}
	startTime      = time.Now()
)

func Mux(exit chan bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(pingPath, ping)
	mux.HandleFunc(launchPath, launch)
	mux.HandleFunc(terminatePath, terminate)
	mux.HandleFunc(eventsPath, events(exit))
	return mux
}

func ping(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(fmt.Sprintf("{\"uptime\": \"%s\"}\n", time.Since(startTime))))
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

	launchType := launch.Type
	launchId := launch.Id
	if !IsToolSupported(launchType) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Unsupported launch type: %s\n", launchType)))
		log.Printf("[UNSUPPORTED_LAUNCH_TYPE] [%s]\n", launchType)
		return
	}
	launchIsAlreadyRunning := launches.PutIfAbsent(launchId, &launch)
	if launchIsAlreadyRunning {
		log.Printf("[LAUNCH_ALREADY_RUNNING] [%s]\n", launchId)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Launch %s is already running", launchId)))
	}
	launchesQueue <- launchId
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
			case evt := <-eventBus.Events():
				{
					data, err := json.Marshal(evt)
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
