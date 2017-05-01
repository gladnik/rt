package api

import (
	. "github.com/aerokube/rt/common"
	"github.com/aerokube/rt/config"
	"github.com/aerokube/rt/event"
	"github.com/aerokube/rt/service"
	"log"
	"sync"
	"time"
)

var (
	launches = &Launches{launches: make(map[string] *Launch)}
	testCases = &TestCases{testCases: make(map[string] *RunningTestCase)}
)

type Launches struct {
	lock sync.RWMutex
	launches map[string] *Launch
}

func (l *Launches) Get(launchId string) (*Launch, bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l, ok := l.launches[launchId]; ok {
		return l, true
	}
	return nil, false
}

func (l *Launches) PutIfAbsent(launchId string, launch *Launch) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	_, isPresent := l.launches[launchId]
	if !isPresent {
		l.launches[launchId] = launch
	}
	return isPresent
}

func (l *Launches) Delete(launchId string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.launches, launchId)
}

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

func (t *TestCases) ForEach(fn func(*RunningTestCase)) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for _, tc := range t.testCases {
		fn(tc)
	}
}

func (t *TestCases) Put(testCaseId string, tc *RunningTestCase) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.testCases[testCaseId] = tc
}

func (t *TestCases) Delete(testCaseId string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.testCases, testCaseId)
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
			waitForTestCasesToFinish(config)
			return
		case launchRequest := <-launchesQueue:
			{
				requestId := launchRequest.RequestId
				launchId := launchRequest.Id
				if launch, ok := launches.Get(launchId); ok {
					go launchImpl(requestId, config, docker, launch)
				} else {
					log.Printf("[%d] [MISSING_LAUNCH] [%s]\n", requestId, launchId)
				}
			}
		}
	}
}

func waitForTestCasesToFinish(config *config.Config) {
	log.Printf("[SHUTTING_DOWN] [%s] [%d]\n", config.ShutdownTimeout, len(testCases.testCases))
	testCases.ForEach(func(tc *RunningTestCase) {
		go func() {
			select {
			case <-tc.Terminated:
				return
			case <-time.After(config.ShutdownTimeout):
				close(tc.Terminated)
			}
		}()
	})
}

func launchImpl(requestId RequestId, config *config.Config, docker *service.Docker, launch *Launch) {
	containerType := launch.Type
	launchId := launch.Id
	eventBus.Fire(event.LaunchStarted, launchId)
	log.Printf("[%d] [LAUNCH_STARTED] [%s] [%s]\n", requestId, launchId, containerType)
	if container, ok := config.GetContainer(containerType); ok {
		parallelBuilds := GetParallelBuilds(container, launch)
		wg := sync.WaitGroup{}
		wg.Add(len(parallelBuilds))
		for testCaseId, pb := range parallelBuilds {
			go func() {
				_, testCaseIsAlreadyRunning := testCases.Get(testCaseId)
				if testCaseIsAlreadyRunning {
					log.Printf("[%d] [TEST_CASE_ALREADY_RUNNING] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
					wg.Done()
					return
				}
				start := time.Now()
				log.Printf("[%d] [LAUNCHING] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
				cancel, finished, err := docker.StartWithCancel(&pb)
				if err != nil {
					eventBus.Fire(event.TestCaseNotStarted, testCaseId)
					log.Printf("[%d] [FAILED_TO_LAUNCH] [%s] [%s] [%s] %v\n", requestId, launchId, containerType, testCaseId, err)
					wg.Done()
					return
				}
				rtc := &RunningTestCase{
					Cancel:     cancel,
					Finished:   finished,
					Terminated: make(chan struct{}),
				}
				testCases.Put(testCaseId, rtc)
				duration := float64(time.Now().Sub(start).Seconds())
				eventBus.Fire(event.TestCaseStarted, testCaseId)
				log.Printf("[%d] [LAUNCHED] [%s] [%s] [%s] [%.2fs]\n", requestId, launchId, containerType, testCaseId, duration)
				select {
				case success := <-rtc.Finished:
					{
						if success {
							eventBus.Fire(event.TestCasePassed, testCaseId)
							log.Printf("[%d] [PASSED] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
						} else {
							eventBus.Fire(event.TestCaseFailed, testCaseId)
							log.Printf("[%d] [FAILED] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
						}
					}

				case <-rtc.Terminated:
					{
						eventBus.Fire(event.TestCaseRevoked, testCaseId)
						log.Printf("[%d] [TERMINATED] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
					}
				case <-time.After(config.Timeout):
					{
						log.Printf("[%d] [TIMED_OUT] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
						terminateImpl(requestId, testCaseId)
						eventBus.Fire(event.TestCaseTimedOut, testCaseId)
						log.Printf("[%d] [TERMINATED] [%s] [%s] [%s]\n", requestId, launchId, containerType, testCaseId)
					}
				}
				testCases.Delete(testCaseId)
				wg.Done()
			}()
		}
		wg.Wait()
		launches.Delete(launchId)
		eventBus.Fire(event.LaunchFinished, launchId)
		log.Printf("[%d] [LAUNCH_FINISHED] [%s] [%s]\n", requestId, launchId, containerType)
	} else {
		log.Printf("[%d] [UNSUPPORTED_CONTAINER_TYPE] [%s] [%s]\n", requestId, launchId, containerType)
	}
}

func ConsumeTerminates(exit chan bool) {
	for {
		select {
		case <-exit:
			return
		case terminateRequest := <-terminateQueue:
			{
				requestId := terminateRequest.RequestId
				testCaseId := terminateRequest.Id
				go terminateImpl(requestId, testCaseId)
			}
		}
	}
}

func terminateImpl(requestId RequestId, testCaseId string) {
	if runningTestCase, ok := testCases.Get(testCaseId); ok {
		log.Printf("[%d] [TERMINATING] [%s]\n", requestId, testCaseId)
		runningTestCase.Cancel()
		close(runningTestCase.Terminated)
	}
}
