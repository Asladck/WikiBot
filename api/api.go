package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

type RecentChangesResponse struct {
	Query struct {
		RecentChanges []struct {
			Title     string `json:"title"`
			TimeStamp string `json:"timestamp"`
			User      string `json:"user"`
		} `json:"recentchanges"`
	} `json:"query"`
}
type SearchResponce struct {
	Query struct {
		Search []struct {
			Title     string `json:"title"`
			Snippet   string `json:"snippet"`
			TimeStamp string `json:"timestamp"`
			User      string `json:"user"`
		} `json:"search"`
	} `json:"query"`
}

func TakeQuery(url string) string {
	var data RecentChangesResponse
	resp, err := http.Get(url)
	if err != nil {
		log.Panic("TakeQuery method error", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Panic("TakeQuery method error 2:", err)
	}
	var cleanData string
	change := data.Query.RecentChanges[0]
	cleanData += "Title:" + change.Title + ",\nTimeStamp: " + change.TimeStamp + ",\nUser: " + change.User + "\n"
	return cleanData
}
func Search(url, query string) string {
	var data SearchResponce
	resp, err := http.Get(url + query)
	if err != nil {
		log.Panic("Search method error", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Panic("Search method error 2:", err)
	}
	var cleanData string
	change := data.Query.Search[0]
	re := regexp.MustCompile(`<[^>]*>`)
	cleanText := re.ReplaceAllString(change.Snippet, "")
	cleanData += "Title:" + change.Title + "\n" + "Information " + cleanText + ",\nTimeStamp: " + change.TimeStamp + ",\nUser: " + change.User + "\n"
	return cleanData
}
