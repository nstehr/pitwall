package main

import (
	"context"
	"flag"
	"log"
	"os"

	"dagger.io/dagger"
)

var (
	workDir = flag.String("workDir", "", "unique name for the orchestrator, defaults to hostname")
)

func main() {
	flag.Parse()
	if *workDir == "" {
		log.Println("workDir must be specified")
		os.Exit(1)
	}
	// Create dagger client
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		log.Println("Error connecting to dagger: ", err)
		os.Exit(1)
	}

	defer client.Close()

	project := client.Host().Directory(*workDir)
	railsDir := project.Directory("pitwall")
	platform := dagger.Platform("linux/amd64")
	err = buildRailsApp(ctx, client, railsDir, platform)
	if err != nil {
		log.Println(err)
	}
	orchestratorDir := project.Directory("orchestrator")
	err = buildOrchestrator(ctx, client, orchestratorDir, platform)
	if err != nil {
		log.Println(err)
	}

	powerunitDir := project.Directory("powerunit")
	err = buildPowerunit(ctx, client, powerunitDir, platform)
	if err != nil {
		log.Println(err)
	}

	terminatorDir := project.Directory("terminator")
	err = buildTerminator(ctx, client, terminatorDir, platform)
	if err != nil {
		log.Println(err)
	}

	cliDir := project.Directory("cli")
	err = buildCli(ctx, client, cliDir, platform)
	if err != nil {
		log.Println(err)
	}

}
