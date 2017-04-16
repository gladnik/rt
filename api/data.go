package api

// Includes flags, tests in parallel
type Property struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Artifact with tests like Maven or NPM artifact
type Artifact struct {
	Group    string `json:"group"`
	Artifact string `json:"artifact"`
	Version  string `json:"version"`
}

// Main execution unit
type TestCase struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Artifact Artifact `json:"artifact"`
	Tags     []string  `json:"tags"` // Suite name is an automatically added tag
}

// A set of test cases launched in the same request
type Launch struct {
	Id string `json:"id"`
	TestCases  []TestCase `json:"testcases"`
	Properties []Property `json:"properties"`
}

const (
	LAUNCH_STARTED = iota
	LAUNCH_FINISHED
	TEST_CASE_STARTED
	TEST_CASE_PASSED
	TEST_CASE_FAILED
	TEST_CASE_REVOKED
)

type Event struct {
	Type int
	Id   string // Test case or launch ID
}