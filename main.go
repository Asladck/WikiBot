package main

import (
	"log"
	//"github.com/Asladck/NewTestTask/api"
	"github.com/Asladck/WikiBot/bot"
	"github.com/Asladck/WikiBot/sql"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("safestuff/token.env", "safestuff/wiki_url.env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файлов:", err)
	}
	mainBot := bot.NewBot(sql.Connect())
	mainBot.Start()
}
