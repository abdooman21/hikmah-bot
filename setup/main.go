package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	godotenv.Load()
	// when you run os.open;  in go it looks for the path from the path u ran the program in  "go run setup/main.go"  from the view of home dir
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
	fmt.Println("this is host", os.Getenv("DB_HOST"))
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("failled to connect: %v", err)
	}
	defer dbConn.Close()

	if err = dbConn.Ping(); err != nil {
		log.Fatalf("cant ping %v", err)
	}

	if err = insertData(dbConn, db); err != nil {
		log.Fatalf("insert data: %v", err)
	}

	fmt.Println("Cnnected and data inserted O.K.")

}

func insertData(dbConn *sql.DB, db Db) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, cat := range db.MainCat {
		// fmt.Print(cat.EnglishName)
		var catgoryID int
		err := tx.QueryRow(
			"INSERT INTO MainCatagories (arabicName, englishName, description, icon_path) VALUES ($1,$2,$3,$4) RETURNING id",
			cat.ArabicName,
			cat.EnglishName,
			cat.Description,
			cat.Icon_path,
		).Scan(&catgoryID)

		if err != nil {
			return fmt.Errorf("failed at insert, %v ", err)
		}

		for _, topic := range cat.Topics {
			var topicid int
			err = tx.QueryRow(
				"INSERT INTO Topics (category_id, name, slug) VALUES ($1,$2,$3) RETURNING id",
				catgoryID,
				topic.Name,
				topic.Slug,
			).Scan(&topicid)
			if err != nil {
				return fmt.Errorf("failed at topic, %v ", err)
			}

			for level, questions := range topic.LevelsData {
				var levelNumber int
				fmt.Sscanf(level, "$d", &levelNumber)
				for _, q := range questions {
					// var qID int
					ans, err := json.Marshal(q.Answers)
					if err != nil {
						return fmt.Errorf("failed at marshaling %v", err)
					}
					_, err = tx.Exec(
						"INSERT INTO Questions (topic_id, level_number, q_text, answers, link) VALUES ($1,$2,$3,$4,$5)",
						topicid,
						levelNumber,
						q.Q,
						ans,
						q.Link)
					if err != nil {
						return err
					}
				}
				fmt.Printf("    Inserted %d qs for level %s\n", len(questions), level)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (t Topic) LevelsSorted() []string {
	levels := make([]string, 0, len(t.LevelsData))
	for k := range t.LevelsData {
		levels = append(levels, k)
	}
	sort.Strings(levels)
	return levels
}
