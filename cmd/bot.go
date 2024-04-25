package main

import (
	"context"
	s "friday_debates/internal/service"
	"log"

	"github.com/sethvargo/go-envconfig"
)

func main() {
	var env s.Env

	if err := envconfig.Process(context.Background(), &env); err != nil {
		log.Fatalln(err)
	}

	svc := s.New(env)
	svc.Run()
}
