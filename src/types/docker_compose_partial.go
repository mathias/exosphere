package types

// DockerComposePartial represents the a part of docker compose file
type DockerComposePartial struct {
	Services    DockerConfigs
	VolumeNames []string
}

// NewDockerComposePartial returns a DockerComposePartial
func NewDockerComposePartial() *DockerComposePartial {
	return &DockerComposePartial{
		Services:    DockerConfigs{},
		VolumeNames: []string{},
	}
}

// Merge returns the a partial combining this and the given partials
func (d *DockerComposePartial) Merge(partials ...*DockerComposePartial) *DockerComposePartial {
	result := NewDockerComposePartial()
	for _, partial := range append(partials, d) {
		for key, val := range partial.Services {
			result.Services[key] = val
		}
		result.VolumeNames = append(result.VolumeNames, partial.VolumeNames...)
	}
	return result
}
