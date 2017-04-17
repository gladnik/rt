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
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Artifact Artifact `json:"artifact"`
	Tags     []string `json:"tags"` // Suite name is an automatically added tag
}

// A set of test cases launched in the same request
type Launch struct {
	Id         string     `json:"id"`
	Type       string     `json:"type"` //I.e. which technology is being used
	TestCases  []TestCase `json:"testcases"`
	Properties []Property `json:"properties"`
}

// Config file struct
type ConfigFile map[string]Config

type Config struct {
	Image   string            `json:"image"`
	DataDir string            `json:"dataDir"`
	Tmpfs   map[string]string `json:"tmpfs"`
}

// Events
const (
	LaunchStarted   = "launch_started"
	LaunchFinished  = "launch_finished"
	TestCaseStarted = "test_case_started"
	TestCasePassed  = "test_case_finished"
	TestCaseFailed  = "test_case_failed"
	TestCaseRevoked = "test_case_revoked"
)

type Event struct {
	Type string
	Id   string // Test case ID or launch ID
}
