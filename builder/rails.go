package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"log"
	"os"
	"strings"

	"dagger.io/dagger"
	dockerClient "github.com/docker/docker/client"
)

func buildRailsApp(ctx context.Context, client *dagger.Client, project *dagger.Directory, platform dagger.Platform) error {
	// could read from .dockerignore to generate this, but there is an open issue I was hitting: https://github.com/dagger/dagger/issues/3791
	// when building non-local some of these wouldn't even be there, since they aren't checked into git
	railsExclusions := []string{"Dockerfile.web", "Dockerfile.worker", ".env", ".gitignore", ".dockerignore", ".git", ".DS_Store", ".gitattributes", "db/*.sqlite3"}

	_, present := os.LookupEnv("SECRET_KEY_BASE")
	if !present {
		return fmt.Errorf("SECRET_KEY_BASE not set on host")
	}
	secret := client.Host().EnvVariable("SECRET_KEY_BASE").Secret()
	// Build our app

	base := client.Container(dagger.ContainerOpts{Platform: platform}).
		From("ruby:3.1.0-alpine").
		WithExec([]string{"apk", "add", "--update", "postgresql-dev", "tzdata", "gcompat", "nodejs"})

		// because I am using the rootFS from base, I'll need to manually set the env variables of the intermediate and
		// prod images.  I should just be able to get them from the base using `base.EnvVariables`, but there is an issue: https://github.com/dagger/dagger/issues/3860
		// preventing that
		// I could also look to refactor how I build this prod image, but for now just copying what I had in the dockerfile, and what I followed from
		//https://lipanski.com/posts/dockerfile-ruby-best-practices
	deps := client.Container(dagger.ContainerOpts{Platform: platform}).
		WithRootfs(base.Rootfs()).
		WithEnvVariable("GEM_HOME", "/usr/local/bundle").
		WithEnvVariable("BUNDLE_APP_CONFIG", "/usr/local/bundle").
		WithEnvVariable("PATH", "/usr/local/bundle/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin").
		WithExec([]string{"apk", "add", "--update", "build-base"}).
		WithFile("Gemfile", project.File("Gemfile")).
		WithFile("Gemfile.lock", project.File("Gemfile.lock")).
		WithExec([]string{"gem", "install", "--platform", "ruby", "google-protobuf", "-v", "3.21.9", "-N"}).
		WithExec([]string{"bundle", "config", "set", "without", "development", "test"}).
		WithExec([]string{"bundle", "install", "--jobs=5", "--retry=3"})

	web := client.Container(dagger.ContainerOpts{Platform: platform}).
		WithRootfs(base.Rootfs()).
		WithEnvVariable("GEM_HOME", "/usr/local/bundle").
		WithEnvVariable("BUNDLE_APP_CONFIG", "/usr/local/bundle").
		WithEnvVariable("PATH", "/usr/local/bundle/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin").
		WithExec([]string{"adduser", "-D", "app"}).
		WithWorkdir("/home/app").
		WithDirectory("/usr/local/bundle/", deps.Directory("/usr/local/bundle/")).
		WithDirectory("./", project.Directory("."), dagger.ContainerWithDirectoryOpts{Exclude: railsExclusions}).
		WithExec([]string{"chown", "-R", "app", "./"}).
		WithUser("app").
		WithSecretVariable("SECRET_KEY_BASE", secret).
		WithEnvVariable("RAILS_ENV", "production").
		WithExec([]string{"bundle", "exec", "rake", "assets:precompile"}).
		WithExec([]string{"mkdir", "-p", "tmp/pids"}).
		WithEnvVariable("RAILS_LOG_TO_STDOUT", "true").
		WithEnvVariable("RAILS_SERVE_STATIC_FILES", "true").
		WithEntrypoint([]string{"bundle", "exec", "rackup", "--host", "0.0.0.0", "-E", "production"})

	worker := client.Container(dagger.ContainerOpts{Platform: platform}).
		WithRootfs(base.Rootfs()).
		WithEnvVariable("GEM_HOME", "/usr/local/bundle").
		WithEnvVariable("BUNDLE_APP_CONFIG", "/usr/local/bundle").
		WithEnvVariable("PATH", "/usr/local/bundle/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin").
		WithExec([]string{"adduser", "-D", "app"}).
		WithWorkdir("/home/app").
		WithDirectory("/usr/local/bundle/", deps.Directory("/usr/local/bundle/")).
		WithDirectory("./", project.Directory("."), dagger.ContainerWithDirectoryOpts{Exclude: railsExclusions}).
		WithExec([]string{"chown", "-R", "app", "./"}).
		WithUser("app").
		WithEnvVariable("RAILS_LOG_TO_STDOUT", "true").
		WithEnvVariable("RAILS_ENV", "production").
		WithEnvVariable("WORKERS", "OrchestratorVmStatusWorker,OrchestratorHealthWorker").
		WithEntrypoint([]string{"bundle", "exec", "rake", "sneakers:run"})

	_, err := web.Export(ctx, "./web")
	if err != nil {
		return err
	}

	_, err = worker.Export(ctx, "./worker")
	if err != nil {
		return err
	}

	err = publishLocal(ctx, "./web", "pitwall-web")
	if err != nil {
		return err
	}
	err = publishLocal(ctx, "./worker", "pitwall-worker")
	if err != nil {
		return err
	}
	return nil
}

func publishLocal(ctx context.Context, tar string, tag string) error {
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return err
	}

	workerFile, err := os.Open(tar)
	if err != nil {
		return err
	}
	defer workerFile.Close()
	resp, err := cli.ImageLoad(ctx, workerFile, false)
	if err != nil {
		return err
	}

	var data map[string]string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	stream := data["stream"]
	log.Println(stream)
	image := strings.Split(stream, "Loaded image ID:")[1]
	image = strings.Trim(image, " ")
	image = strings.Trim(image, "\n")
	err = cli.ImageTag(ctx, image, tag)
	if err != nil {
		return err
	}
	log.Printf("Pushed %s to local repository\n", tag)
	return nil
}
