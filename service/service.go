package service

type Starter interface {
	StartWithCancel() (func(), error)
}

// Build settings
type BuildSettings struct {
	Image        string
	Command      string
	Tmpfs        map[string]string
	DataDir      string //Data directory inside container
	TemplateFile string
	BuildData    map[string]string
	BuildFile    string
}
