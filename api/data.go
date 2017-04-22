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