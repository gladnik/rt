package api

import (
	. "github.com/aerokube/rt/common"
	"github.com/aerokube/rt/config"
	"github.com/aerokube/rt/service"
	"github.com/emicklei/go-restful/log"
)

const (
	Maven = "maven"
)

var (
	supportedTools = map[string]Tool{
		Maven: &MavenTool{},
	}
)

type Command []string

type Tool interface {
	GetSettings(testCase TestCase, properties []Property) Command
}

// Converts launch object to a set of build settings for each separate container
func GetParallelBuilds(container *config.Container, launch *Launch) map[string]service.BuildSettings {
	ret := make(map[string]service.BuildSettings)
	tool, ok := supportedTools[launch.Type]
	if ok {
		//TODO: could do this in parallel with goroutines...
		for _, testCase := range launch.TestCases {
			bs := service.BuildSettings{
				Image:     container.Image,
				Command:   tool.GetSettings(testCase, launch.Properties),
				Tmpfs:     container.Tmpfs,
				DataDir:   container.DataDir,
				Templates: container.Templates,
				BuildData: StandaloneTestCase{
					TestCase:   testCase,
					Properties: launch.Properties,
				},
			}
			ret[testCase.Id] = bs
		}
	} else {
		log.Printf("Trying to use unsupported tool: %s. This is probably a bug.\n", launch.Type)
	}
	return ret
}

func IsToolSupported(toolType string) bool {
	_, ok := supportedTools[toolType]
	return ok
}
