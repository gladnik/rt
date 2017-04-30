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
	l.lock.RUnlock()
	defer l.lock.RUnlock()
	if l, ok := l.launches[launchId]; ok {
		return l, true
	}
	return nil, false
}

func (l *Launches) Put(launchId string, launch *Launch) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.launches[launchId] = launch
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
		case launchId := <-launchesQueue:
			{
				if launch, ok := launches.Get(launchId); ok {
					launchImpl(config, docker, launch)
				} else {
					log.Printf("[MISSING_LAUNCH] [%s]\n", launchId)
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

func launchImpl(config *config.Config, docker *service.Docker, launch *Launch) {
	containerType := launch.Type
	launchId := launch.Id
	wg := sync.WaitGroup{}
	log.Printf("[LAUNCH_STARTED] [%s] [%s]\n", launchId, containerType)
	eventBus.Fire(event.LaunchStarted, launchId)
	if container, ok := config.GetContainer(containerType); ok {
		parallelBuilds := GetParallelBuilds(container, launch)
		for testCaseId, pb := range parallelBuilds {
			go func() {
				start := time.Now()
				log.Printf("[LAUNCHING] [%s] [%s] [%s]\n", launchId, containerType, testCaseId)
				cancel, finished, err := docker.StartWithCancel(&pb)
				if err != nil {
					log.Printf("[FAILED_TO_LAUNCH] [%s] [%s] [%s] %v\n", launchId, containerType, testCaseId, err)
					return
				}
				rtc := &RunningTestCase{
					Cancel:     cancel,
					Finished:   finished,
					Terminated: make(chan struct{}),
				}
				testCases.Put(testCaseId, rtc)
				duration := float64(time.Now().Sub(start).Seconds())
				wg.Add(1)
				eventBus.Fire(event.TestCaseStarted, testCaseId)
				log.Printf("[LAUNCHED] [%s] [%s] [%s] [%.2fs]\n", launchId, containerType, testCaseId, duration)
				select {
				case success := <-rtc.Finished:
					{
						if success {
							eventBus.Fire(event.TestCasePassed, testCaseId)
							log.Printf("[PASSED] [%s] [%s] [%s]\n", launchId, containerType, testCaseId)
						} else {
							eventBus.Fire(event.TestCaseFailed, testCaseId)
							log.Printf("[FAILED] [%s] [%s] [%s]\n", launchId, containerType, testCaseId)
						}
					}

				case <-rtc.Terminated:
					{
						eventBus.Fire(event.TestCaseRevoked, testCaseId)
						log.Printf("[TERMINATED] [%s] [%s] [%s]\n", launchId, containerType, testCaseId)
					}
				case <-time.After(config.Timeout):
					{
						log.Printf("[TIMED_OUT] [%s] [%s] [%s]\n", launchId, containerType, testCaseId)
						terminateImpl(testCaseId)
						eventBus.Fire(event.TestCaseTimedOut, testCaseId)
						log.Printf("[TERMINATED] [%s] [%s] [%s]\n", launchId, containerType, testCaseId)
					}
				}
				testCases.Delete(testCaseId)
				wg.Done()
			}()
		}
		go func() {
			wg.Wait()
			eventBus.Fire(event.LaunchFinished, launchId)
			log.Printf("[LAUNCH_FINISHED] [%s] [%s]\n", launchId, containerType)
		}()
		return
	}
	log.Printf("[UNSUPPORTED_CONTAINER_TYPE] [%s] [%s]\n", launchId, containerType)
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
	}
}
