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
	"sync"
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
	launchesQueue  = make(chan IdentifiedRequest)
	terminateQueue = make(chan IdentifiedRequest)
	eventBus       = event.NewEventBus()
	upgrader       = websocket.Upgrader{}
	startTime      = time.Now()
	
	num      RequestId
	numLock  sync.Mutex
)

type IdentifiedRequest struct {
	RequestId RequestId
	Id string
}

func Mux(exit chan bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(pingPath, ping)
	mux.HandleFunc(launchPath, launch)
	mux.HandleFunc(terminatePath, terminate)
	mux.HandleFunc(eventsPath, events(exit))
	return mux
}

func ping(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(struct {
		Uptime         string `json:"uptime"`
	}{time.Since(startTime).String()})
}

func launch(w http.ResponseWriter, r *http.Request) {
	requestId := serial()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("[%d] [UNSUPPORTED_LAUNCH_METHOD] [%s]\n", requestId, r.Method)
		return
	}
	var launch Launch
	err := json.NewDecoder(r.Body).Decode(&launch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("A launch object is expected"))
		log.Printf("[%d] [INVALID_LAUNCH_DATA] [%s]\n", requestId, r.Method)
		return
	}

	launchType := launch.Type
	launchId := launch.Id
	if !IsToolSupported(launchType) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Unsupported launch type: %s\n", launchType)))
		log.Printf("[%d] [UNSUPPORTED_LAUNCH_TYPE] [%s]\n", requestId, launchType)
		return
	}
	launchIsAlreadyRunning := launches.PutIfAbsent(launchId, &launch)
	if launchIsAlreadyRunning {
		log.Printf("[%d] [LAUNCH_ALREADY_RUNNING] [%s]\n", requestId, launchId)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Launch %s is already running", launchId)))
		return
	}
	launchesQueue <- IdentifiedRequest{RequestId: requestId, Id: launchId}
	log.Printf("[%d] [LAUNCH_REQUESTED] [%s]\n", requestId, launchId)
}

func terminate(w http.ResponseWriter, r *http.Request) {
	requestId := serial()
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("[%d] [UNSUPPORTED_TERMINATE_METHOD] [%s]\n", requestId, r.Method)
		return
	}
	var uuids []string
	err := json.NewDecoder(r.Body).Decode(&uuids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("An array of test case IDs is expected"))
		log.Printf("[%d] [INVALID_TERMINATE_DATA]\n", requestId)
		return
	}
	for _, uuid := range uuids {
		log.Printf("[%d] [TERMINATE_REQUESTED] [%s]\n", requestId, uuid)
		terminateQueue <- IdentifiedRequest{RequestId: requestId, Id: uuid}
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

func serial() RequestId {
	numLock.Lock()
	defer numLock.Unlock()
	id := num
	num++
	return id
}
