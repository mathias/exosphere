package docker

import (
	"context"
	"os/exec"

	"github.com/Originate/exosphere/src/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerRunner struct {
	client *client.Client
	logger *util.Logger
}

func NewDockerRunner(logger *util.Logger) (*DockerRunner, error) {
	client, err := client.NewEnvClient()
	return &DockerRunner{client: client}, err
}

func (d *DockerRunner) pullImageIfNecissary(ctx context.Context, imageName string) error {
	imageListFilter := filters.NewArgs()
	imageListFilter.Add("reference", imageName)
	results, err := d.client.ImageList(ctx, types.ImageListOptions{Filters: imageListFilter})
	if err != nil {
		return err
	}
	if len(results) != 0 {
		return nil
	}
	d.logger.LogNewf("Pulling image %s", imageName)
	readCloser, err := d.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	defer readCloser.Close()
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
	case <-readCloserToChan(readCloser):
	}
	return nil
}

func (d *DockerRunner) createContainer(ctx context.Context, imageName string, options ContainerOptions) (string, error) {
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	mounts := []mount.Mount{}
	for containerPort, hostPort := range options.Ports {
		portBindings[nat.Port(containerPort)] = []nat.PortBinding{nat.PortBinding{HostPort: hostPort}}
		exposedPorts[nat.Port(containerPort)] = struct{}{}
	}
	for containerMount, hostMount := range options.Mounts {
		mounts = append(mounts, mount.Mount{Source: hostMount, Target: containerMount, Type: mount.TypeBind})
	}
	response, err := d.client.ContainerCreate(
		ctx,
		&container.Config{
			Tty:          true,
			Image:        imageName,
			Cmd:          options.Cmd,
			Env:          options.Env,
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			AutoRemove:   true,
			PortBindings: portBindings,
			Mounts:       mounts,
		},
		&network.NetworkingConfig{},
		options.Name,
	)
	return response.ID, err
}

func (d *DockerRunner) createAndRunContainer(ctx context.Context, imageName string, options ContainerOptions) error {
	d.logger.LogNewf("Creating container for image %s", imageName)
	containerId, err := d.createContainer(ctx, imageName, options)
	if err != nil {
		return err
	}
	d.logger.LogNewf("Starting container %s", containerId)
	return d.client.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
}

func (d *DockerRunner) buildImage(ctx context.Context, tag string, path string, dockerFileName string) error {
	d.logger.LogNewf("Starting build for %s", tag)
	// Run as exec command since docker go api does not expose a simple way to build an image
	// without reinventing .dockerignore reading/excluding
	cmd := exec.CommandContext(ctx, "docker", "build", "--file", dockerFileName, "--tag", tag, ".")
	cmd.Dir = path
	stdOut, err := cmd.StdoutPipe()
	if err == nil {
		go logReader(stdOut, d.logger)
	}
	stdErr, err := cmd.StderrPipe()
	if err == nil {
		go logReader(stdErr, d.logger)
	}
	err = cmd.Run()
	return err
}

func (d *DockerRunner) RunDockerHubContainer(ctx context.Context, imageName string, options ContainerOptions) error {
	if err := d.pullImageIfNecissary(ctx, imageName); err != nil {
		return err
	}
	return d.createAndRunContainer(ctx, imageName, options)
}

func (d *DockerRunner) RunDirectory(ctx context.Context, path string, dockerFileName string, options ContainerOptions) error {
	imageName := formateImageName(options.Name)
	err := d.buildImage(ctx, imageName, path, dockerFileName)
	if err != nil {
		return err
	}
	return d.createAndRunContainer(ctx, imageName, options)
}
