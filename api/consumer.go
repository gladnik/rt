package api

func ConsumeLaunches(exit chan bool) {
	for {
		select {
		case <-exit:
			return
		case launch := <-launchesQueue:
			{
				launchImpl(&launch)
			}
		}
	}
}

func launchImpl(launch *Launch) {
	//TODO: should launch a set of test case each in separate goroutine
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
	//TODO: to be implemented!
}
