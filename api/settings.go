package api

import (
	"github.com/aerokube/rt/config"
	"github.com/aerokube/rt/service"
)

// Converts launch object to a set of build settings for each separate container
func GetParallelBuilds(container *config.Container, launch *Launch) map[string]service.BuildSettings {
	//TODO: to be implemented!
	return nil
}
