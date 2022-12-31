package main

import (
	"context"

	"dagger.io/dagger"
)

func buildPowerunit(ctx context.Context, client *dagger.Client, project *dagger.Directory, platform dagger.Platform) error {

	builder := client.Container(dagger.ContainerOpts{Platform: platform}).From("golang:latest")
	builder = builder.WithMountedDirectory("/src", project).WithWorkdir("/src")
	path := "dist"
	builder = builder.WithExec([]string{"go", "build", "-o", "dist/powerunit"})

	output := builder.Directory(path)

	_, err := output.Export(ctx, path)
	if err != nil {
		return err
	}
	return nil
}
