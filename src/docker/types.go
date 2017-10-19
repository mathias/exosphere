package docker

type ContainerOptions struct {
	Name    string
	Cmd     []string
	Env     []string
	Volumes map[string]struct{}
	Ports   map[string]string
	Mounts  map[string]string
}
