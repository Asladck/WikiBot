package safestuff

import (
	"log"
	"os"
)

func TokenTaker() string {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Panic("Ошибка: DISCORD_TOKEN не найден в .env")
	}
	return token
}
func UrlTaker(lang string) string {
	var url string
	switch lang {
	case "rus":
		url = os.Getenv("RU_URL_WIKI")
	case "eng":
		url = os.Getenv("EN_URL_WIKI")
	}
	if url == "" {
		log.Fatal("URL not Found")
	}
	return url
}
func SearchTaker(lang string) string {
	var url string
	switch lang {
	case "rus":
		url = os.Getenv("RU_SEARCH_WIKI")
	case "eng":
		url = os.Getenv("EN_SEARCH_WIKI")
	}
	if url == "" {
		log.Fatal("URL not Found")
	}
	return url
}
