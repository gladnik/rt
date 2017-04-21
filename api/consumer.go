package api

import (
	"github.com/aerokube/rt/config"
	"github.com/aerokube/rt/service"
	"log"
	"sync"
)

var (
	testCases = &TestCases{}
)

type TestCases struct {
	lock      sync.RWMutex
	testCases map[string]*RunningTestCase // key is testCaseId
}

func (t *TestCases) Get(testCaseId string) (*RunningTestCase, bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if tc, ok := t.testCases[testCaseId]; ok {
		return tc, true
	}
	return nil, false
}

func (t *TestCases) Put(testCaseId string, tc *RunningTestCase) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.testCases[testCaseId] = tc
}

type RunningTestCase struct {
	Cancel     func()
	Finished   <-chan bool
	Terminated chan struct{}
}

func ConsumeLaunches(config *config.Config, exit chan bool) {
	docker, err := service.NewDocker(config)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-exit:
			return
		case launch := <-launchesQueue:
			{
				launchImpl(config, docker, &launch)
			}
		}
	}
}

func launchImpl(config *config.Config, docker *service.Docker, launch *Launch) {
	containerType := launch.Type
	if container, ok := config.GetContainer(containerType); ok {
		log.Printf("[LAUNCHING] [%s] [%s]\n", launch.Id, containerType)
		parallelBuilds := GetParallelBuilds(container, launch)
		for testCaseId, pb := range parallelBuilds {
			go func() {
				cancel, finished, err := docker.StartWithCancel(&pb)
				if err != nil {
					log.Printf("[FAILED_TO_LAUNCH] [%s] [%s] %v\n", launch.Id, containerType, err)
					return
				}
				rtc := &RunningTestCase{
					Cancel:     cancel,
					Finished:   finished,
					Terminated: make(chan struct{}),
				}
				testCases.Put(testCaseId, rtc)
			}()
		}
		return
	}
	log.Printf("[UNSUPPORTED_CONTAINER_TYPE] [%s] [%s]\n", launch.Id, containerType)
}

func ConsumeTerminates(exit chan bool) {
	for {
		select {
		case <-exit:
			return
		case testCaseId := <-terminateQueue:
			{
				terminateImpl(testCaseId)
			}
		}
	}
}

func terminateImpl(testCaseId string) {
	if runningTestCase, ok := testCases.Get(testCaseId); ok {
		log.Printf("[TERMINATING] [%s]\n", testCaseId)
		runningTestCase.Cancel()
		close(runningTestCase.Terminated)
		log.Printf("[TERMINATED] [%s]\n", testCaseId)
	}
}
