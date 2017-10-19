package application

import (
	"context"

	"github.com/Originate/exosphere/src/docker"
	"github.com/Originate/exosphere/src/util"
)

func RunService(logger *util.Logger) error {
	d, err := docker.NewDockerRunner(logger)
	if err != nil {
		return err
	}
	ctx := context.Background()

	err = d.RunDirectory(ctx, "/Users/alexdavid/Development/src/github.com/Originate/test", "Dockerfile", docker.ContainerOptions{
		Name:   "Mygocontainer",
		Ports:  map[string]string{"8080": "1234", "9999": "8888"},
		Mounts: map[string]string{"/mnt": "/Users/alexdavid/foobar"},
	})
	if err != nil {
		return err
	}
	return nil
}
