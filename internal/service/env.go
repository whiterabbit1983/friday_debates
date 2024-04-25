package service

type Env struct {
	BotToken    string `env:"BOT_TOKEN"`
	JulepApiKey string `env:"JULEP_API_KEY"`
	JulepBaseUrl string `env:"JULEP_BASE_URL"`
}
