package cmd

import (
	"os"

	"github.com/Originate/exosphere/src/application"
	"github.com/Originate/exosphere/src/docker/composewriter"
	"github.com/Originate/exosphere/src/types"
	"github.com/Originate/exosphere/src/util"
	"github.com/spf13/cobra"
)

var productionFlag bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs an Exosphere application",
	Long:  "Runs an Exosphere application",
	Run: func(cmd *cobra.Command, args []string) {
		if printHelpIfNecessary(cmd, args) {
			return
		}
		appDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		homeDir, err := util.GetHomeDirectory()
		if err != nil {
			panic(err)
		}
		appConfig, err := types.NewAppConfig(appDir)
		if err != nil {
			panic(err)
		}
		dockerComposeProjectName := composewriter.GetDockerComposeProjectName(appDir)
		writer := os.Stdout
		buildMode := composewriter.BuildMode{
			Type:        composewriter.BuildModeTypeLocal,
			Mount:       true,
			Environment: composewriter.BuildModeEnvironmentDevelopment,
		}
		if productionFlag {
			buildMode.Environment = composewriter.BuildModeEnvironmentProduction
		} else if noMountFlag {
			buildMode.Mount = false
		}
		runner, err := application.NewRunner(appConfig, writer, appDir, homeDir, dockerComposeProjectName, buildMode)
		if err != nil {
			panic(err)
		}
		if err := runner.Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().BoolVarP(&noMountFlag, "no-mount", "", false, "Run without mounting")
	runCmd.PersistentFlags().BoolVarP(&productionFlag, "production", "", false, "Run in production mode")
}
