package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

type Db struct {
	Description string    `json:"description"`
	MainCat     []catgory `json:"mainCategories"`
}
type catgory struct {
	Id          int
	ArabicName  string  `json:"arabicName"`
	EnglishName string  `json:"englishName"`
	Description string  `json:"description"`
	Icon_path   string  `json:"icons"`
	Topics      []Topic `json:"topics"`
}
type Topic struct {
	Name       string                `json:"name"`
	Slug       string                `json:"slug"`
	LevelsData map[string][]Question `json:"levelsData"`
}

type Question struct {
	ID      int      `json:"id"`
	Q       string   `json:"q"`
	Link    string   `json:"link"`
	Answers []Answer `json:"answers"`
}

type Answer struct {
	Answer string `json:"answer"`
	T      int    `json:"t"`
}

func main() {

	// when you run os.open;  in go it looks for the path from the dir u ran the program in  "go run setup/main.go"  from the view of home dir
	path := "setup/database.json"

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("file not found path: %s, err %s ", path, err.Error())
	}
	defer f.Close()

	var db Db

	dec := json.NewDecoder(f)
	if err := dec.Decode(&db); err != nil {
		log.Fatal(err)
	}

	for _, cat := range db.MainCat {
		for _, topic := range cat.Topics {
			for level, questions := range topic.LevelsData {
				fmt.Println("Level:", level)
				for _, q := range questions {
					fmt.Println("Q:", q.Q)
				}
			}
		}
	}

	// for dec.More() {

	// }

}

func (t Topic) LevelsSorted() []string {
	levels := make([]string, 0, len(t.LevelsData))
	for k := range t.LevelsData {
		levels = append(levels, k)
	}
	sort.Strings(levels)
	return levels
}
