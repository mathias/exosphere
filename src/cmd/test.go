package cmd

import (
	"log"
	"os"

	"github.com/Originate/exosphere/src/application"
	"github.com/Originate/exosphere/src/docker/composewriter"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs tests for the application",
	Long:  "Runs tests for the application",
	Run: func(cmd *cobra.Command, args []string) {
		if printHelpIfNecessary(cmd, args) {
			return
		}

		context, err := GetContext()
		if err != nil {
			log.Fatal(err)
		}
		writer := os.Stdout
		buildMode := composewriter.BuildMode{
			Type:        composewriter.BuildModeTypeLocal,
			Mount:       true,
			Environment: composewriter.BuildModeEnvironmentTest,
		}
		if noMountFlag {
			buildMode.Mount = false
		}

		var testsPassed bool
		if context.HasServiceContext {
			testsPassed, err = application.TestService(context.ServiceContext, writer, buildMode)
		} else {
			testsPassed, err = application.TestApp(context.AppContext, writer, buildMode)
		}
		if err != nil {
			panic(err)
		}
		if !testsPassed {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().BoolVarP(&noMountFlag, "no-mount", "", false, "Run tests without mounting")
}
