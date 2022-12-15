package vm

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func buildFilesystemFromImage(ctx context.Context, image string, publicKey string) (string, error) {
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
	log.Println(len(imgInspect.RootFS.Layers))

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
	// TODO: let's explore using initrfs to contain the init binary and the public key
	// this would allow us to have reuse amongst the Docker image based filesystem
	err = injectInit(path)
	if err != nil {
		return "", err
	}
	err = injectPublicKey(path, publicKey)
	log.Println("DONE")
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

func injectInit(path string) error {
	_, err := os.Stat("powerunit")
	if err != nil {
		return err
	}

	powerUnitDir := fmt.Sprintf("%s/etc/powerunit", path)
	os.Mkdir(powerUnitDir, 0755)
	bytesRead, err := ioutil.ReadFile("powerunit")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/powerunit", powerUnitDir), bytesRead, 0755)

	if err != nil {
		return err
	}

	if _, err := os.Lstat(fmt.Sprintf("%s/sbin/init", path)); err == nil {
		// init already exists in some form
		log.Println("init already exists, renaming to init.old")

		err = os.Rename(fmt.Sprintf("%s/sbin/init", path), fmt.Sprintf("%s/sbin/init.old", path))
		if err != nil {
			return err
		}
	}
	err = os.Symlink("/etc/powerunit/powerunit", fmt.Sprintf("%s/sbin/init", path))
	return err
}

func injectPublicKey(path string, publicKey string) error {
	if publicKey == "" {
		log.Println("No public key specified, ssh to VM will not work")
		return nil
	}

	// for now write key to the powerunit directory
	powerUnitDir := fmt.Sprintf("%s/etc/powerunit", path)
	if _, err := os.Stat(powerUnitDir); os.IsNotExist(err) {
		os.Mkdir(powerUnitDir, 0755)
	}

	f, err := os.Create(fmt.Sprintf("%s/key", powerUnitDir))

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(publicKey)

	if err != nil {
		return err
	}
	return nil
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
	// TODO: switch to syscall.Mount
	cmd := exec.Command("mount", fmt.Sprintf("%s.ext4", name), name)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("mounting...")
	err := cmd.Run()
	if err != nil {
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
	log.Println("unmount....")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func createFilesystem(name string) error {
	arguments := []string{}
	arguments = append(arguments, "if=/dev/zero")
	arguments = append(arguments, fmt.Sprintf("of=%s.ext4", name))
	arguments = append(arguments, "bs=5M")
	arguments = append(arguments, "count=1000")
	cmd := exec.Command("dd", arguments...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("creating file using dd")
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("mkfs.ext4", fmt.Sprintf("%s.ext4", name))
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	log.Println("creating empty filesystem")
	err = cmd.Run()
	if err != nil {
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
				err = untar(tarReader, target)
				if err != nil {
					return err
				}
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
		// symlink navigation from inspired from here: https://github.com/hashicorp/go-getter/blob/85c3ba950122547165a31bf8f6080c6e71c49ce0/decompress_tar.go
		if header.Typeflag == tar.TypeSymlink {
			// If the type is a symlink we re-write it and
			// continue instead of attempting to copy the contents

			// TODO: do I need to add some more safety here (and in other spots?) for any tar expanding attacks
			if _, err := os.Lstat(path); err == nil {
				// link exists
				continue
			}

			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("failed writing symbolic link: %s", err)
			}
			continue
		}

		// process one file at to make sure we open, read, and close it as to not go over any ulimit settings
		// https://stackoverflow.com/a/66684218
		err = writeFile(tarReader, path, info.Mode())
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(tarReader io.Reader, filePath string, fileMode os.FileMode) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileMode)
	if err != nil {
		return err
	}
	defer file.Close() // close error discarded
	_, err = io.Copy(file, tarReader)
	return err
}
