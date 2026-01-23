package web

import (
	"encoding/json"
	"log"
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
		log.Println("an Error ccourd will making new req :", err)
		return ""
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Println("An Error occuerd with respond :", err)
		return ""
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	catfact := &catRes{}

	err = decoder.Decode(catfact)
	if err != nil {
		log.Println("Error decoding respond,: ", err)
		return ""
	}

	return catfact.Fact
}
