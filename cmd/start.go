package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	internalconfig "github.com/tbourrely/cody/internal/configuration"
	"github.com/tbourrely/cody/internal/docker"
	"github.com/tbourrely/cody/internal/networking"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <instance name>",
	Short: "Start a cody instance",
	Long:  "Start a cody instance with the given name (default to folder name if not specified).",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := internalconfig.Load(os.DirFS("/"))
		if err != nil {
			panic(err)
		}

		var port int
		if config.IsRangeValid() {
			port, err = networking.FindRandomPortInRange(config.Ports.Start, config.Ports.End)
		} else {
			port, err = networking.FindRandomPort()
		}

		var authToken string
		if config.AuthToken != "" {
			authToken = config.AuthToken // TODO : validate token before using
		} else {
			authToken, err = docker.GenerateToken()
		}

		if err != nil {
			panic(err)
		}

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

		var instanceName string

		if len(args) == 1 {
			instanceName = args[0]
		} else {
			instanceName, err = docker.GenerateName()
			if err != nil {
				panic(err)
			}
		}

		err = docker.Run(cli, ctx, instanceName, port, authToken)
		if err != nil {
			panic(err)
		}

		var url string
		for i := 1; i <= 10; i++ {
			url, err = docker.Url(cli, ctx, instanceName)
			time.Sleep(1 * time.Second)

			if err == nil {
				break
			}
		}

		if err != nil {
			panic(err)
		}

		fmt.Println(url)
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
