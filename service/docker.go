package service

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"log"
	"time"
)

type Docker struct {
	DataDir       string //Data directory on host machine
	BuildSettings *BuildSettings
	Client        *client.Client
	LogConfig     *container.LogConfig
}

func (docker *Docker) StartWithCancel() (func(), error) {
	ctx := context.Background()
	env := []string{
		fmt.Sprintf("TZ=%s", time.Local),
	}
	volumes := make(map[string]string)
	volumes[docker.BuildSettings.DataDir] = docker.DataDir
	resp, err := docker.Client.ContainerCreate(ctx,
		&container.Config{
			Hostname: "localhost",
			Image:    docker.BuildSettings.Image,
			Volumes:  volumes,
			Env:      env,
			Cmd:      docker.BuildSettings.Command,
		},
		&container.HostConfig{
			AutoRemove: true,
			LogConfig:  *docker.LogConfig,
			Tmpfs:      docker.BuildSettings.Tmpfs,
			ShmSize:    268435456,
			Privileged: true,
		},
		&network.NetworkingConfig{}, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to create container: %v", err)
	}
	containerStartTime := time.Now()
	log.Println("[STARTING_CONTAINER]")
	err = docker.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		removeContainer(ctx, docker.Client, resp.ID)
		return nil, nil, fmt.Errorf("Failed to start container: %v", err)
	}
	log.Printf("[CONTAINER_STARTED] [%s] [%v]\n", resp.ID, time.Since(containerStartTime))
	return func() { removeContainer(ctx, docker.Client, resp.ID) }, nil
}

func removeContainer(ctx context.Context, cli *client.Client, id string) {
	log.Printf("[REMOVING_CONTAINER] [%s]\n", id)
	err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		log.Println("error: unable to remove container", id, err)
		return
	}
	log.Printf("[CONTAINER_REMOVED] [%s]\n", id)
}
