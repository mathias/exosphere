package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/Originate/exosphere/src/types"
	"github.com/Originate/exosphere/src/util"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
	"github.com/pkg/errors"
)

// CatFileInDockerImage reads the file fileName inside the given image
func CatFileInDockerImage(c *client.Client, imageName, fileName string) ([]byte, error) {
	if err := PullImage(c, imageName); err != nil {
		return []byte(""), err
	}
	command := fmt.Sprintf("cat %s", fileName)
	output, err := util.Run("", fmt.Sprintf("docker run --rm %s %s", imageName, command))
	return []byte(output), err
}

// GetDockerCompose reads docker-compose.yml at the given path and
// returns the dockerCompose object
func GetDockerCompose(dockerComposePath string) (result types.DockerCompose, err error) {
	yamlFile, err := ioutil.ReadFile(dockerComposePath)
	if err != nil {
		return result, err
	}
	err = yaml.Unmarshal(yamlFile, &result)
	if err != nil {
		return result, errors.Wrap(err, "Failed to unmarshal docker-compose.yml")
	}
	return result, nil
}

// GetExitCode returns the exit code for the given container
func GetExitCode(containerName string) (int, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return 1, err
	}
	containerJSON, err := c.ContainerInspect(context.Background(), containerName)
	if err != nil {
		return 1, err
	}
	if containerJSON.State.Status != "exited" {
		return 1, fmt.Errorf("%s has not exited", containerName)
	}
	return containerJSON.State.ExitCode, nil
}

// ListRunningContainers returns the names of running containers
// and an error (if any)
func ListRunningContainers(c *client.Client) ([]string, error) {
	containerNames := []string{}
	ctx := context.Background()
	containers, err := c.ContainerList(ctx, dockerTypes.ContainerListOptions{})
	if err != nil {
		return containerNames, err
	}
	for _, container := range containers {
		containerNames = append(containerNames, strings.Replace(container.Names[0], "/", "", -1))
	}
	return containerNames, nil
}

// ListImages returns the names of all images and an error (if any)
func ListImages(c *client.Client) ([]string, error) {
	imageNames := []string{}
	ctx := context.Background()
	imageSummaries, err := c.ImageList(ctx, dockerTypes.ImageListOptions{
		All: true,
	})
	if err != nil {
		return imageNames, err
	}
	for _, imageSummary := range imageSummaries {
		if len(imageSummary.RepoTags) > 0 {
			repoTag := imageSummary.RepoTags[0]
			imageNames = append(imageNames, strings.Split(repoTag, ":")[0])
		}
	}
	return imageNames, nil
}

// PullImage pulls the given image from DockerHub, returns an error if any
func PullImage(c *client.Client, image string) error {
	ctx := context.Background()
	stream, err := c.ImagePull(ctx, image, dockerTypes.ImagePullOptions{})
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(stream)
	return err
}

// RunInDockerContainer runs the given command in a new writeable container layer
// over the given image, removes the container when the command exits, and returns
// the an error if any
func RunInDockerContainer(runConfig RunConfig) error {
	cmd := []string{"docker", "run", "--rm"}
	for _, volume := range runConfig.Volumes {
		cmd = append(cmd, "-v", volume)
	}
	if runConfig.Interactive {
		cmd = append(cmd, "-it")
	}
	if runConfig.WorkingDir != "" {
		cmd = append(cmd, "-w", runConfig.WorkingDir)
	}
	cmd = append(cmd, runConfig.ImageName)
	cmd = append(cmd, runConfig.Command...)
	return util.RunAndPipe("", []string{}, runConfig.Writer, cmd...)
}

// TagImage tags a docker image srcImage as targetImage
func TagImage(srcImage, targetImage string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	return dockerClient.ImageTag(ctx, srcImage, targetImage)
}

// PushImage pushes image with imageName to the registry given an encoded auth object
func PushImage(c *client.Client, writer io.Writer, imageName, encodedAuth string) error {
	util.PrintSectionHeader(writer, fmt.Sprintf("docker push %s", imageName))
	ctx := context.Background()
	stream, err := c.ImagePush(ctx, imageName, dockerTypes.ImagePushOptions{
		RegistryAuth: encodedAuth,
	})
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		err = printPushProgress(writer, scanner.Text())
		if err != nil {
			return fmt.Errorf("Cannot push image '%s': %s", imageName, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "error reading ImagePush output")
	}
	fmt.Println()
	return stream.Close()
}

func printPushProgress(writer io.Writer, output string) error {
	outputObject := struct {
		Status      string      `json:"status"`
		ID          string      `json:"id"`
		ErrorDetail interface{} `json:"errorDetail"`
		Error       string      `json:"error"`
	}{}
	err := json.Unmarshal([]byte(output), &outputObject)
	if err != nil {
		return err
	}
	if outputObject.Error != "" {
		return errors.New(outputObject.Error)
	}
	_, err = fmt.Fprintf(writer, "%s: %s\n", outputObject.Status, outputObject.ID)
	return err
}
