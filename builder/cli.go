package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func buildCli(ctx context.Context, client *dagger.Client, project *dagger.Directory, platform dagger.Platform) error {

	oses := []string{"linux", "darwin"}
	arches := []string{"amd64", "arm64"}
	b := client.Container(dagger.ContainerOpts{Platform: platform}).From("golang:latest")
	b = b.WithMountedDirectory("/src", project).WithWorkdir("/src")
	path := "dist"

	for _, goos := range oses {
		for _, goarch := range arches {
			builder := b.WithEnvVariable("GOOS", goos)
			builder = builder.WithEnvVariable("GOARCH", goarch)

			builder = builder.WithExec([]string{"go", "build", "-o", fmt.Sprintf("dist/%s/%s/pwctl", goos, goarch)})

			output := builder.Directory(path)

			_, err := output.Export(ctx, path)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
