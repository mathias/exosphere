package composebuilder

import (
	"github.com/Originate/exosphere/src/config"
	"github.com/Originate/exosphere/src/types"
)

// GetApplicationPartial returns a partial for a application
func GetApplicationPartial(options ApplicationOptions) (*types.DockerComposePartial, error) {
	dependenciesPartial, err := GetDependenciesPartial(options)
	if err != nil {
		return nil, err
	}
	servicesPartial, err := GetServicesPartial(options)
	if err != nil {
		return nil, err
	}
	return dependenciesPartial.Merge(servicesPartial), nil
}

// GetDependenciesPartial returns a partial for all the application dependencies
func GetDependenciesPartial(options ApplicationOptions) (*types.DockerComposePartial, error) {
	result := types.NewDockerComposePartial()
	if options.BuildMode.Type == BuildModeTypeDeploy {
		appDependencies := config.GetBuiltAppProductionDependencies(options.AppConfig, options.AppDir)
		for _, builtDependency := range appDependencies {
			if builtDependency.HasDockerConfig() {
				dockerConfig, err := builtDependency.GetDockerConfig()
				if err != nil {
					return result, err
				}
				result.Services[builtDependency.GetServiceName()] = dockerConfig
			}
		}
	} else {
		appDependencies := config.GetBuiltAppDevelopmentDependencies(options.AppConfig, options.AppDir, options.HomeDir)
		for _, builtDependency := range appDependencies {
			dockerConfig, err := builtDependency.GetDockerConfig()
			if err != nil {
				return result, err
			}
			result.Services[builtDependency.GetContainerName()] = dockerConfig
		}
	}
	return result, nil
}

// GetServicesPartial returns a partial for all the application services
func GetServicesPartial(options ApplicationOptions) (*types.DockerComposePartial, error) {
	result := types.NewDockerComposePartial()
	serviceConfigs, err := config.GetServiceConfigs(options.AppDir, options.AppConfig)
	if err != nil {
		return result, err
	}
	serviceData := options.AppConfig.GetServiceData()
	for serviceRole, serviceConfig := range serviceConfigs {
		servicePartial, err := GetServicePartial(options.AppConfig, serviceConfig, serviceData[serviceRole], serviceRole, options.AppDir, options.HomeDir, options.BuildMode)
		if err != nil {
			return result, err
		}
		result = result.Merge(servicePartial)
	}
	return result, nil
}
