package cmd

import (
	goContext "context"
	"fmt"
	"log"
	"os"

	"github.com/Originate/exosphere/src/application"
	"github.com/docker/docker/api/types/filters"
	"github.com/moby/moby/client"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes dangling Docker images and volumes",
	Run: func(cmd *cobra.Command, args []string) {
		if printHelpIfNecessary(cmd, args) {
			return
		}
		writer := os.Stdout
		userContext, err := GetUserContext()
		if err != nil {
			log.Fatal(err)
		}
		err = application.GenerateComposeFiles(userContext.AppContext)
		if err != nil {
			log.Fatal(err)
		}
		c, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(writer, "Removing dangling images")
		imagesPruneReport, err := c.ImagesPrune(goContext.Background(), filters.NewArgs())
		for _, imageDeleted := range imagesPruneReport.ImagesDeleted {
			fmt.Fprintln(writer, imageDeleted.Deleted)
		}
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(writer, "Removing dangling volumes")
		volumesPruneReport, err := c.VolumesPrune(goContext.Background(), filters.NewArgs())
		for _, volumeDeleted := range volumesPruneReport.VolumesDeleted {
			fmt.Fprintln(writer, volumeDeleted)
		}
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(writer, "Removing application and test containers")
		err = application.CleanContainers(userContext.AppContext, writer)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
}
