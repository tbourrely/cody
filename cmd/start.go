package cmd

import (
	"context"

	"github.com/cody/internal/docker"
	"github.com/cody/internal/networking"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a cody instance",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return
		}
		defer cli.Close()

		err = docker.Build(cli, ctx)
		if err != nil {
			panic(err)
		}

		availablePort, err := networking.FindRandomPort()
		if err != nil {
			panic(err)
		}

		err = docker.Run(cli, ctx, availablePort)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
