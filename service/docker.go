package service

import (
	"context"
	"fmt"
	. "github.com/aerokube/rt/common"
	"github.com/aerokube/rt/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"log"
	"path"
	"time"
	"encoding/json"
)

type Docker struct {
	dataDir   string //Data directory on host machine
	client    *client.Client
	logConfig *container.LogConfig
}

func NewDocker(config *config.Config) (*Docker, error) {
	cl, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v\n", err)
	}
	return &Docker{
		dataDir:   config.DataDir,
		client:    cl,
		logConfig: config.LogConfig,
	}, nil
}

func (docker *Docker) StartWithCancel(bs *BuildSettings) (func(), <-chan bool, error) {
	ctx := context.Background()
	rawTemplates, err := marshalData(bs.Templates)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal templates info: %v\n", err)
	}
	rawBuildData, err := marshalData(bs.BuildData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal build data: %v\n", err)
	}
	env := []string{
		fmt.Sprintf("TZ=%s", time.Local),
		fmt.Sprintf("%s=%s", DataDir, bs.DataDir),
		fmt.Sprintf("%s=%s", Templates, rawTemplates),
		fmt.Sprintf("%s=%s", BuildData, rawBuildData),
	}
	volumes := []string{fmt.Sprintf("%s:%s", path.Join(docker.dataDir, bs.BuildData.TestCase.Id), bs.DataDir)}
	volumes = append(volumes, bs.Volumes...)
	resp, err := docker.client.ContainerCreate(ctx,
		&container.Config{
			Hostname: "localhost",
			Image:    bs.Image,
			Env:      env,
			Cmd:      bs.Command,
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds:      volumes,
			LogConfig:  *docker.logConfig,
			Tmpfs:      bs.Tmpfs,
			ShmSize:    268435456,
			Privileged: true,
		},
		&network.NetworkingConfig{}, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create container: %v", err)
	}
	containerId := resp.ID
	requestId := bs.RequestId
	testCaseId := bs.BuildData.TestCase.Id
	image := bs.Image
	containerStartTime := time.Now()
	log.Printf("[%d] [STARTING_CONTAINER] [%s] [%s]\n", requestId, testCaseId, image) 
	err = docker.client.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
	finished := make(chan bool)
	go docker.waitFor(ctx, containerId, finished)
	if err != nil {
		docker.removeContainer(ctx, containerId, bs)
		return nil, nil, fmt.Errorf("failed to start container: %v", err)
	}
	log.Printf("[%d] [CONTAINER_STARTED] [%s] [%s] [%s] [%v]\n", requestId, testCaseId, image, containerId, time.Since(containerStartTime))
	return func() { docker.removeContainer(ctx, containerId, bs) }, finished, nil
}

func marshalData(m interface{}) (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %v\n", err)
	}
	return string(data), nil
}

func (docker *Docker) waitFor(ctx context.Context, containerId string, finished chan bool) {
	//TODO: does this automatically exit on container removal?
	statusCode, err := docker.client.ContainerWait(ctx, containerId)
	success := err != nil && statusCode == 0
	finished <- success
}

func (docker *Docker) removeContainer(ctx context.Context, containerId string, bs *BuildSettings) {
	requestId := bs.RequestId
	testCaseId := bs.BuildData.TestCase.Id
	image := bs.Image
	log.Printf("[%d] [REMOVING_CONTAINER] [%s] [%s] [%s]\n", requestId, testCaseId, image, containerId)
	err := docker.client.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		log.Println("error: unable to remove container", containerId, err)
		return
	}
	log.Printf("[%d] [CONTAINER_REMOVED] [%s] [%s] [%s]\n", requestId, testCaseId, image, containerId)
}
