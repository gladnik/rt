package service

import . "github.com/aerokube/rt/common"

type Starter interface {
	StartWithCancel(bs *BuildSettings) (func(), <-chan bool, error)
}

// Build settings
type BuildSettings struct {
	RequestId RequestId
	Image     string
	Command   []string
	Tmpfs     map[string]string
	DataDir   string //Data directory inside container
	Templates map[string]string
	Volumes   []string
	BuildData StandaloneTestCase
}
