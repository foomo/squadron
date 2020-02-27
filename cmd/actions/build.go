package actions

import (
	"context"
	"log"

	"github.com/foomo/configurd"
)

func Build(dir, service, tag string) {
	cnf, err := configurd.New(dir)
	if err != nil {
		log.Fatal(err)
	}

	svc, err := cnf.Service(service)
	if err != nil {
		log.Fatalf("service not found: %v", err)
	}
	output, err := svc.Build(context.Background(), tag)
	if err != nil {
		log.Fatalf("could not build: %v  output:\n%v", output, err)
	}
	log.Print(output)
}
