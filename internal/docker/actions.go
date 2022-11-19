package docker

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

// Build is used to build the image.
func Build(cli *client.Client, ctx context.Context) (err error) {
	dir, err := ioutil.TempDir("", "cody")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	// Dockerfile
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	dockerFileContent, err := Files.ReadFile("files/Dockerfile")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(dockerfilePath, dockerFileContent, fs.ModePerm)
	if err != nil {
		return
	}

	// start script
	startPath := filepath.Join(dir, "start.sh")
	startContent, err := Files.ReadFile("files/start.sh")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(startPath, startContent, fs.ModePerm)
	if err != nil {
		return
	}

	tar, err := archive.TarWithOptions(dir, &archive.TarOptions{})
	if err != nil {
		return
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"cody:latest"},
		Remove:     true,
	}
	res, err := cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		return
	}
	defer res.Body.Close()
	_, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}

	return nil
}

// Run is used to start a container.
func Run(cli *client.Client, ctx context.Context) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	target := filepath.Join("/home/workspace", filepath.Base(cwd))

	name, err := containerName(target)
	if err != nil {
		return
	}

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: "cody",
			Cmd:   []string{"3000", "mytoken"}, // TODO : variabilize
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
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: cwd,
					Target: target,
				},
			},
		}, nil, nil, name)
	if err != nil {
		return
	}

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	return nil
}

// Stop is used to stop a container
func Stop(cli *client.Client, ctx context.Context, instance string) (deleted bool, err error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == fmt.Sprintf("/%s", instance) {
				_ = cli.ContainerStop(ctx, container.ID, nil)
				_ = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
				return true, nil
			}
		}
	}

	return
}

func containerName(path string) (string, error) {
	base := filepath.Base(path)
	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(base, "_"), nil
}
