package vm

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func buildFilesystemFromImage(ctx context.Context, image string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	io.Copy(os.Stdout, reader)
	imgInspect, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return "", err
	}

	log.Println(imgInspect.Config.Entrypoint)
	log.Println(imgInspect.Config.Cmd)
	log.Println(imgInspect.RootFS.Type)

	rc, err := cli.ImageSave(ctx, []string{image})
	if err != nil {
		return "", err
	}
	defer rc.Close()

	// right now, will do everything in the pwd of the pitwall executable,
	// creating a directory with the image name to hold the contents of the image
	// and a corresponding ext4 filesystem that is named <image>.ext4
	path := image
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	// create an empty ext4 filesystem
	err = createFilesystem(path)
	if err != nil {
		return "", err
	}
	// mount the filesytem against the empty directory
	err = mount(path)
	if err != nil {
		return "", err
	}
	// extract the layers from the image into the directory (and now the filesystem, since its mounted)
	err = extractLayers(rc, path)
	if err != nil {
		return "", err
	}
	// unmount the filesystem
	err = unmount(path)
	if err != nil {
		return "", err
	}
	// set ownership, probably not necessary but useful for debugging
	err = setOwnership(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.ext4", image), nil
}

func setOwnership(path string) error {
	// TODO: set real permissions
	err := os.Chmod(path, 0777)
	if err != nil {
		return err
	}
	return err
}

func mount(name string) error {
	cmd := exec.Command("mount", fmt.Sprintf("%s.ext4", name), name)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("mounting...")
	err := cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		return err
	}
	return nil
}

func unmount(path string) error {
	cmd := exec.Command("umount", path)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("unmount....")
	err := cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		return err
	}
	return nil
}

func createFilesystem(name string) error {
	arguments := []string{}
	arguments = append(arguments, "if=/dev/zero")
	arguments = append(arguments, fmt.Sprintf("of=%s.ext4", name))
	arguments = append(arguments, "bs=1M")
	arguments = append(arguments, "count=500")
	cmd := exec.Command("dd", arguments...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("creating file using dd")
	err := cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		return err
	}
	cmd = exec.Command("mkfs.ext4", fmt.Sprintf("%s.ext4", name))
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("creating empty filesystem")
	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		return err
	}

	return nil
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
