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
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"time"
)

type Docker struct {
	DataDir   string //Data directory on host machine
	Client    *client.Client
	LogConfig *container.LogConfig
}

func NewDocker(config *config.Config) (*Docker, error) {
	cl, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v\n", err)
	}
	return &Docker{
		DataDir:   config.DataDir,
		Client:    cl,
		LogConfig: config.LogConfig,
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
	volumes := []string{fmt.Sprintf("%s:%s", docker.DataDir, bs.DataDir)}
	resp, err := docker.Client.ContainerCreate(ctx,
		&container.Config{
			Hostname: "localhost",
			Image:    bs.Image,
			Env:      env,
			Cmd:      bs.Command,
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds:      volumes,
			LogConfig:  *docker.LogConfig,
			Tmpfs:      bs.Tmpfs,
			ShmSize:    268435456,
			Privileged: true,
		},
		&network.NetworkingConfig{}, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create container: %v", err)
	}
	containerStartTime := time.Now()
	log.Println("[STARTING_CONTAINER]")
	err = docker.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	finished := make(chan bool)
	go docker.waitFor(ctx, resp.ID, finished)
	if err != nil {
		docker.removeContainer(ctx, resp.ID)
		return nil, nil, fmt.Errorf("failed to start container: %v", err)
	}
	log.Printf("[CONTAINER_STARTED] [%s] [%v]\n", resp.ID, time.Since(containerStartTime))
	return func() { docker.removeContainer(ctx, resp.ID) }, finished, nil
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
	statusCode, err := docker.Client.ContainerWait(ctx, containerId)
	success := err != nil && statusCode == 0
	finished <- success
}

func (docker *Docker) removeContainer(ctx context.Context, containerId string) {
	log.Printf("[REMOVING_CONTAINER] [%s]\n", containerId)
	err := docker.Client.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		log.Println("error: unable to remove container", containerId, err)
		return
	}
	log.Printf("[CONTAINER_REMOVED] [%s]\n", containerId)
}
