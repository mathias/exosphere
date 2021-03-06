package deployer

import (
	"fmt"

	"github.com/Originate/exosphere/src/aws"
	"github.com/Originate/exosphere/src/terraform"
	"github.com/Originate/exosphere/src/types"
	"github.com/Originate/exosphere/src/types/deploy"
)

// StartDeploy starts the deployment process
// nolint gocyclo
func StartDeploy(deployConfig deploy.Config) error {
	err := validateConfigs(deployConfig)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(deployConfig.Writer, "Setting up AWS account...")
	if err != nil {
		return err
	}
	err = aws.InitAccount(deployConfig.AwsConfig)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(deployConfig.Writer, "Pushing Docker images to ECR...")
	if err != nil {
		return err
	}
	imagesMap, err := PushApplicationImages(deployConfig)
	if err != nil {
		return err
	}
	prevTerraformFileContents, err := terraform.ReadTerraformFile(deployConfig)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(deployConfig.Writer, "Generating Terraform files...")
	if err != nil {
		return err
	}
	err = terraform.GenerateFile(deployConfig)
	if err != nil {
		return err
	}
	err = terraform.CheckTerraformFile(deployConfig, prevTerraformFileContents)
	if err != nil {
		return err
	}
	fmt.Fprintln(deployConfig.Writer, "Retrieving secrets...")
	secrets, err := aws.ReadSecrets(deployConfig.AwsConfig)
	if err != nil {
		return err
	}

	return deployApplication(deployConfig, imagesMap, secrets)
}

func validateConfigs(deployConfig deploy.Config) error {
	fmt.Fprintln(deployConfig.Writer, "Validating application configuration...")
	err := deployConfig.AppContext.Config.Remote.ValidateFields()
	if err != nil {
		return err
	}

	fmt.Fprintln(deployConfig.Writer, "Validating service configurations...")
	for _, serviceContext := range deployConfig.AppContext.ServiceContexts {
		err = serviceContext.Config.ValidateDeployFields(serviceContext.Source.Location, serviceContext.Config.Type)
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(deployConfig.Writer, "Validating application dependencies...")
	validatedDependencies := map[string]string{}
	for _, dependency := range deployConfig.AppContext.Config.Remote.Dependencies {
		err = dependency.ValidateFields()
		if err != nil {
			return err
		}
		validatedDependencies[dependency.Name] = dependency.Version
	}

	fmt.Fprintln(deployConfig.Writer, "Validating service dependencies...")
	for _, serviceContext := range deployConfig.AppContext.ServiceContexts {
		for _, dependency := range serviceContext.Config.Remote.Dependencies {
			if validatedDependencies[dependency.Name] == "" {
				err = dependency.ValidateFields()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func deployApplication(deployConfig deploy.Config, imagesMap map[string]string, secrets types.Secrets) error {
	fmt.Fprintln(deployConfig.Writer, "Retrieving remote state...")
	err := terraform.RunInit(deployConfig)
	if err != nil {
		return err
	}

	fmt.Fprintln(deployConfig.Writer, "Applying changes...")
	return terraform.RunApply(deployConfig, secrets, imagesMap, deployConfig.AutoApprove)
}
