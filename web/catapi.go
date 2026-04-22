package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type catRes struct {
	Fact string `json:"fact"`
}

var catApiPath string = "https://catfact.ninja/fact"

func GetCatFact() string {

	c := &http.Client{Timeout: 2 * time.Second}

	req, err := http.NewRequest("GET", catApiPath, nil)
	if err != nil {
		slog.Error("an Error ccourd will making new req :", "err", err)
		return ""
	}
	resp, err := c.Do(req)
	if err != nil {
		slog.Error("An Error occuerd with respond :", "err", err)
		return ""
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	catfact := &catRes{}

	err = decoder.Decode(catfact)
	if err != nil {
		slog.Error("Error decoding respond,:", "err", err)
		return ""
	}

	return catfact.Fact
}
