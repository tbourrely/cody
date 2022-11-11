package cmd

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
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
			panic(err)
		}
		defer cli.Close()

		tar, err := archive.TarWithOptions("docker", &archive.TarOptions{})
		if err != nil {
			panic(err)
		}

		opts := types.ImageBuildOptions{
			Dockerfile: "Dockerfile",
			Tags:       []string{"cody:latest"},
			Remove:     true,
		}
		res, err := cli.ImageBuild(ctx, tar, opts)
		if err != nil {
			panic(err)
		}

		defer res.Body.Close()

		resp, err := cli.ContainerCreate(ctx,
			&container.Config{
				Image: "cody",
				Cmd:   []string{"3000", "mytoken"},
				Tty:   false,
				ExposedPorts: nat.PortSet{
					"3000/tcp": struct{}{},
				},
			},
			&container.HostConfig{
				AutoRemove: true,
				PortBindings: nat.PortMap{
					"3000/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "3000",
						},
					},
				},
			}, nil, nil, "cody")
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
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
