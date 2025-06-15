package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

var ChampionsCsvFilename = "champions.csv"
var MatchupsCsvFilename = "matchups.csv"
var SynergiesCsvFilename = "synergies.csv"
var MatchupsJsonFilename = "matchups.json"

type Champion struct {
	Name string
	Role string
}

type EvaluatedChampion struct {
	Role     string
	Matchups map[string]int
}

func main() {
	start := time.Now()

	champions := getChampionsFromCsv()

	//createMatchupsCsv(champions, "")
	//createMatchupsJson(champions)

	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func createMatchupsCsv(champions []Champion, filename string) {
	matchups := make([][]string, 0, (len(champions)*(len(champions)-1))/2)
	matchups = append(matchups, []string{"name", "opposition", "evaluation"})

	for i, champion := range champions {
		for _, opposition := range champions[i+1:] {
			evaluation := "3"
			if filename == SynergiesCsvFilename && champion.Role == "Melee" && champion.Role == opposition.Role {
				evaluation = "2"
			}
			matchups = append(matchups, []string{champion.Name, opposition.Name, evaluation})
		}
	}

	saveMatchupsToCsvFile(matchups, filename)
}

func saveMatchupsToCsvFile(matchups [][]string, filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	writer := csv.NewWriter(f)
	err = writer.WriteAll(matchups)
	if err != nil {
		log.Fatal(err)
	}
}

func createMatchupsJson(champions []Champion) {
	matchups := make(map[string]EvaluatedChampion)

	for _, champion := range champions {
		matchups[champion.Name] = EvaluatedChampion{
			Role:     champion.Role,
			Matchups: evaluateChampionMatchups(champions, champion),
		}
	}

	saveMatchupsToJsonFile(matchups)
}

func getChampionsFromCsv() []Champion {
	f, err := os.Open(ChampionsCsvFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var champions []Champion
	for i, row := range records {
		if i == 0 {
			continue
		}

		champions = append(champions, Champion{
			Name: row[0],
			Role: row[1],
		})
	}

	return champions
}

func evaluateChampionMatchups(champions []Champion, yourChampion Champion) map[string]int {
	championMatchups := make(map[string]int)

	for _, champion := range champions {
		if champion.Name == yourChampion.Name {
			championMatchups[champion.Name] = 0
		} else {
			championMatchups[champion.Name] = rand.Intn(11) - 5
		}
	}

	return championMatchups
}

func saveMatchupsToJsonFile(data interface{}) {
	f, err := os.Create(MatchupsJsonFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		log.Fatal(err)
	}
}
