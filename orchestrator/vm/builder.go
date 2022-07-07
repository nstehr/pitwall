package vm

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func buildFilesystemFromImage(ctx context.Context, image string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)
	imgInspect, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		panic(err)
	}

	fmt.Println(imgInspect.Config.Entrypoint)
	fmt.Println(imgInspect.Config.Cmd)
	fmt.Println(imgInspect.RootFS.Type)

	rc, err := cli.ImageSave(ctx, []string{image})
	if err != nil {
		panic(err)
	}
	defer rc.Close()
	extractLayers(rc, "test")
}

func extractLayers(reader io.Reader, target string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		info := header.FileInfo()

		if !info.IsDir() {
			if strings.Contains(header.Name, "layer.tar") {
				untar(tarReader, target)
			}
		}
	}

	return nil
}

// https://golangdocs.com/tar-gzip-in-golang
func untar(reader io.Reader, target string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
