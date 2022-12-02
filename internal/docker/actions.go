package docker

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	internaltypes "github.com/tbourrely/cody/internal/types"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

var imageName = "cody"

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
		Tags:       []string{fmt.Sprintf("%s:latest", imageName)},
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
func Run(cli *client.Client, ctx context.Context, name string, port int, authToken string) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	target := filepath.Join("/home/workspace", filepath.Base(cwd))

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: "cody",
			Cmd:   []string{"3000", authToken},
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
						HostPort: fmt.Sprint(port),
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
	container, err := findContainerByName(cli, ctx, instance)
	if err != nil {
		return
	}

	err = cli.ContainerStop(ctx, container.ID, nil)
	if err != nil {
		return
	}

	return true, nil
}

func Url(cli *client.Client, ctx context.Context, instance string) (url string, err error) {
	container, err := findContainerByName(cli, ctx, instance)
	if err != nil {
		return
	}

	options := types.ContainerLogsOptions{ShowStdout: true}
	var cLogs io.ReadCloser
	var content []byte

	cLogs, err = cli.ContainerLogs(ctx, container.ID, options)
	if err != nil {
		return
	}

	content, err = io.ReadAll(cLogs)
	if err != nil {
		return
	}

	url = replacePort(findUrl(string(content)), container)
	if url == "" {
		err = errors.New("could not determine instance URL")
	}

	return
}

func GetInstances(cli *client.Client, ctx context.Context) (instances []internaltypes.Instance) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, container := range containers {
		if container.Image != imageName {
			continue
		}

		instances = append(instances, internaltypes.Instance{Name: container.Names[0][1:]})
	}

	return
}

func GenerateName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return safeContainerName(filepath.Base(cwd))
}

func GenerateToken() (string, error) {
	n := 16
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func InstallExtension(cli *client.Client, ctx context.Context, instance string, extensionId string) error {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/home/.openvscode-server/bin/openvscode-server", "--install-extension", extensionId},
	}

	cresp, err := cli.ContainerExecCreate(ctx, instance, execConfig)
	if err != nil {
		return err
	}

	// TODO : see if possible ot use events to detect exec end
	// instead of waiting output for nothing
	aresp, err := cli.ContainerExecAttach(ctx, cresp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer aresp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, aresp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return err
		}
		break

	case <-ctx.Done():
		return ctx.Err()
	}

	iresp, err := cli.ContainerExecInspect(ctx, cresp.ID)
	if err != nil {
		return err
	}

	if iresp.ExitCode != 0 {
		return errors.New(fmt.Sprintf("Process failed with code %d", iresp.ExitCode))
	}

	return nil
}

func SetEditorSettings(cli *client.Client, ctx context.Context, container string, settings string) error {
	dstPath := "/home/workspace/.openvscode-server/data/Machine/settings.json"
	dstInfo := archive.CopyInfo{Path: dstPath, Exists: false, IsDir: false}

	// Wait for 10 seconds for folder to be created
	var pathExists bool
	for i := 1; i <= 10; i++ {
		time.Sleep(1 * time.Second)
		_, err := cli.ContainerStatPath(ctx, container, filepath.Dir(dstPath))
		// no error == path exists
		if err == nil {
			pathExists = true
			break
		}
	}

	if pathExists == false {
		return errors.New("Settings path does not exists")
	}

	// Create tmp file with settings content
	file, err := ioutil.TempFile("", "settings")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	file.Write([]byte(settings))

	srcInfo, err := archive.CopyInfoSourcePath(file.Name(), true)
	if err != nil {
		return err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return err
	}
	defer srcArchive.Close()

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return err
	}
	defer preparedArchive.Close()

	return cli.CopyToContainer(ctx,
		container,
		dstDir,
		preparedArchive,
		types.CopyToContainerOptions{})
}

func findContainerByName(cli *client.Client, ctx context.Context, name string) (types.Container, error) {
	var result types.Container

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return result, err
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if containerName == fmt.Sprintf("/%s", name) {
				return container, nil
			}
		}
	}

	return result, errors.New("container not found")
}

func safeContainerName(name string) (string, error) {
	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(name, "_"), nil
}

func replacePort(url string, container types.Container) string {
	bindings := container.Ports[0]
	return strings.ReplaceAll(url, fmt.Sprint(bindings.PrivatePort), fmt.Sprint(bindings.PublicPort))
}

func findUrl(containerLogs string) string {
	r, _ := regexp.Compile("http://(.*):(.*)/?tkn=(.*)")
	result := r.FindString(containerLogs)
	return result
}
