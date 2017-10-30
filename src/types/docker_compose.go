package types

// DockerCompose represents the docker compose object
type DockerCompose struct {
	Version  string
	Services DockerConfigs
	Volumes  map[string]interface{}
}

// NewDockerCompose joins the given docker compose partials into one
func NewDockerCompose(partial *DockerComposePartial) DockerCompose {
	result := DockerCompose{Version: "3", Services: partial.Services}
	for _, volumeName := range partial.VolumeNames {
		result.Volumes[volumeName] = nil
	}
	return result
}
