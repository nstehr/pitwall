package main

import (
	"log"
	"os"

	"github.com/nstehr/pitwall/terminator/ziti"
	zitiSdk "github.com/openziti/sdk-golang/ziti"
)

func main() {

	zitiController, present := os.LookupEnv("ZITI_CONTROLLER")
	if !present {
		log.Fatal("Must specify ZITI_CONTROLLER environment variable")
	}
	zitiUser, present := os.LookupEnv("ZITI_USER")
	if !present {
		log.Fatal("Must specify ZITI_USER environment variable")
	}
	zitiPass, present := os.LookupEnv("ZITI_PASS")
	if !present {
		log.Fatal("Must specify ZITI_PASS environment variable")
	}
	client, err := ziti.NewClient(zitiController, zitiUser, zitiPass)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Login()
	if err != nil {
		log.Fatal(err)
	}
	id, err := client.CreateIdentity(ziti.Service, "foo", false, []string{})
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := client.EnrollIdentity(id)
	if err != nil {
		log.Fatal(err)
	}

	serviceId, err := client.CreateService("bar", true, []string{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(serviceId)

	ztx := zitiSdk.NewContextWithConfig(cfg)
	err = ztx.Authenticate()
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}
}
