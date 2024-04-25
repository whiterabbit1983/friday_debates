package main

import (
	"context"
	s "friday_debates/internal/service"
	"log"

	"github.com/joho/godotenv"
	"github.com/objectbox/objectbox-go/objectbox"
	"github.com/sethvargo/go-envconfig"
)

func initObjectBox() (*objectbox.ObjectBox, error) {
	objectBox, err := objectbox.NewBuilder().Model(s.ObjectBoxModel()).Build()

	if err != nil {
		return nil, err
	}

	return objectBox, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("error loading .env file")
	}

	ob, err := initObjectBox()

	if err != nil {
		log.Fatalln("error loading objects", err)
	}

	defer ob.Close()

	userInfoBox := s.BoxForUserInfo(ob)
	flowStepBox := s.BoxForFlowStep(ob)

	var env s.Env

	if err := envconfig.Process(context.Background(), &env); err != nil {
		log.Fatalln(err)
	}

	api := s.NewAPI(env.JulepBaseUrl, env.JulepApiKey, "application/json")
	svc := s.New(env, userInfoBox, flowStepBox, api)
	svc.Run()
}
