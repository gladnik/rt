package api

import (
	"testing"
	. "github.com/aandryashin/matchers"
	. "github.com/aerokube/rt/common"
	"github.com/aerokube/rt/config"
	"github.com/aerokube/rt/service"
)
var (
	testContainer = config.Container{
		Image: "test-image",
		DataDir: "test-dir",
		Tmpfs: map[string]string{},
		Templates: map[string]string{},
		Volumes: []string{},
	}
	
	testArtifact = Artifact{
		Id: "test-id",
		GroupId: "test-group-id",
		Version: "test-version",
	}
	
	testCase1 = TestCase{
		Id: "test-case-1",
		Name: "com.aerokube.rt.TestSuite#testCase1",
		Artifact: testArtifact,
		Tags: []string{},
	}
	
	testCase2 = TestCase{
		Id: "test-case-2",
		Name: "com.aerokube.rt.TestSuite#testCase2",
		Artifact: testArtifact,
		Tags: []string{},
	}
	
	testProperties = []Property{
		{Key:"key1", Value: "value1"},
		{Key:"key2", Value: "value2"},
	}
	
	testLaunch = Launch{
		Id: "test-launch-id",
		Type: "maven",
		TestCases: []TestCase{testCase1, testCase2},
		Properties: testProperties,
	}
	
	testCommand = []string{"test-command"}
	
)

func init() {
	supportedTools[Maven] = &MockTool{testCommand}
}

type MockTool struct {
	Command Command
}

func (m *MockTool) GetCommand(container *config.Container, testCase TestCase, properties []Property) Command {
	return m.Command
}

func TestGetParallelBuilds(t *testing.T) {
	parallelBuilds := GetParallelBuilds(&testContainer, &testLaunch)
	correctBuilds := map[string] service.BuildSettings{
		"test-case-1": {
			Image: "test-image",
			Command: testCommand,
			Tmpfs: map[string] string{},
			DataDir: "test-dir",
			Templates: map[string] string{},
			Volumes: []string{},
			BuildData: StandaloneTestCase{
				TestCase: testCase1,
				Properties: testProperties,
			},
		},
		"test-case-2": {
			Image: "test-image",
			Command: testCommand,
			Tmpfs: map[string] string{},
			DataDir: "test-dir",
			Templates: map[string] string{},
			Volumes: []string{},
			BuildData: StandaloneTestCase{
				TestCase: testCase2,
				Properties: testProperties,
			},
		},
	}
	AssertThat(t, parallelBuilds, EqualTo{correctBuilds})
}