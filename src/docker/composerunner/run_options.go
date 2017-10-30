package composerunner

import (
	"io"

	"github.com/Originate/exosphere/src/types"
)

// RunOptions are the options passed into Run
type RunOptions struct {
	DockerComposeDir         string
	DockerComposeProjectName string
	DockerComposePartial     *types.DockerComposePartial
	Writer                   io.Writer
	AbortOnExit              bool
}
