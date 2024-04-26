package main

import (
	"context"
	s "friday_debates/internal/service"
	"log"

	"github.com/chaisql/chai"
	"github.com/joho/godotenv"

	"github.com/sethvargo/go-envconfig"
)

func initDB(db *chai.DB) {
	err := db.Exec(`
        CREATE TABLE IF NOT EXISTS user_info (
			chat_id INT,
            name TEXT,
			about TEXT,
			user_id TEXT,
			session_id TEXT
        )
    `)

	if err != nil {
		log.Fatalln(err)
	}

	err = db.Exec(`
        CREATE TABLE IF NOT EXISTS flow_steps (
			chat_id INT,
            step TEXT
        )
    `)

	if err != nil {
		log.Fatalln(err)
	}

	err = db.Exec(`
        CREATE TABLE IF NOT EXISTS session_info (
			chat_id INT,
            session_id TEXT
        )
    `)

	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("error loading .env file")
	}

	db, err := chai.Open("bot_db")

	if err != nil {
		log.Fatalln(err)
	}

	var env s.Env

	if err := envconfig.Process(context.Background(), &env); err != nil {
		log.Fatalln(err)
	}

	initDB(db)

	api := s.NewAPI(env.JulepBaseUrl, env.JulepApiKey, "application/json")
	svc := s.New(env, db, api)
	svc.Run()
}
