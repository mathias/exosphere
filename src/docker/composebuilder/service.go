package composebuilder

import (
	"path"
	"regexp"
	"strings"

	"github.com/Originate/exosphere/src/types"
)

// GetServicePartial returns a partial for a service and its dependencies
func GetServicePartial(appConfig types.AppConfig, serviceConfig types.ServiceConfig, serviceData types.ServiceData, role string, appDir string, homeDir string, mode BuildMode) (*types.DockerComposePartial, error) {
	if mode.Type == BuildModeTypeDeploy {
		return NewProductionDockerComposeBuilder(appConfig, serviceConfig, serviceData, role, appDir).getPartial()
	}
	return NewDevelopmentDockerComposeBuilder(appConfig, serviceConfig, serviceData, role, appDir, homeDir, mode).getPartial()
}

// GetDockerComposeProjectName creates a docker compose project name the same way docker-compose mutates the COMPOSE_PROJECT_NAME env var
func GetDockerComposeProjectName(appDir string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]")
	replacedStr := reg.ReplaceAllString(path.Base(appDir), "")
	return strings.ToLower(replacedStr)
}
