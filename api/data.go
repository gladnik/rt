package api

type Pack struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	TestCases  []TestCase `json:"testcases"`
	Properties []Property `json:"properties"`
}

// Includes flags, tests in parallel
type Property struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TestCase struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Artifact *Artifact `json:"artifact"`
	Tags     []string  `json:"tags"` // Suite name is an automatically added tag
}

// Artifact with tests like Maven or NPM artifact
type Artifact struct {
	Group    string `json:"group"`
	Artifact string `json:"artifact"`
	Version  string `json:"version"`
}

type Launch struct {
	Id        string           `json:"id"`
	Pack      *Pack            `json:"pack,omitempty"`
	State     int              `json:"state"`
	TestCases []TestCaseLaunch `json:"testcases"`
}

const (
	LAUNCH_QUEUED = iota
	LAUNCH_RUNNING
	LAUNCH_FINISHED
	LAUNCH_REVOKED
	TEST_CASE_QUEUED
	TEST_CASE_RUNNING
	TEST_CASE_PASSED
	TEST_CASE_FAILED
	TEST_CASE_REVOKED
)

type TestCaseLaunch struct {
	Id       string    `json:"id"`
	TestCase *TestCase `json:"testcase"`
	State    int       `json:"state"`
}

type Event struct {
	Type int
	Id   string
}

type LaunchEvent struct {
	Event
}

type TestCaseEvent struct {
	Event
}
