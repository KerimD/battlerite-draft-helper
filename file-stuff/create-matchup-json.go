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
	names := createNamesSlice(champions)

	//createMatchupsCsv(champions, names)
	//createMatchupsJson(champions, names)

	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func createMatchupsCsv(champions []Champion, names []string) {
	matchups := make([][]string, 0, (len(champions)*(len(champions)-1))/2)
	matchups = append(matchups, []string{"name", "opposition", "evaluation"})

	for i, name := range names {
		for _, opposition := range names[i+1:] {
			matchups = append(matchups, []string{name, opposition, "3"})
		}
	}

	saveMatchupsToCsvFile(matchups)
}

func saveMatchupsToCsvFile(matchups [][]string) {
	f, err := os.OpenFile(MatchupsCsvFilename, os.O_WRONLY, 0644)
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

func createMatchupsJson(champions []Champion, names []string) {
	matchups := make(map[string]EvaluatedChampion)

	for _, champion := range champions {
		matchups[champion.Name] = EvaluatedChampion{
			Role:     champion.Role,
			Matchups: evaluateChampionMatchups(names, champion),
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

func createNamesSlice(champions []Champion) []string {
	names := make([]string, 0, len(champions))

	for _, champion := range champions {
		names = append(names, champion.Name)
	}

	return names
}

func evaluateChampionMatchups(names []string, champion Champion) map[string]int {
	championMatchups := make(map[string]int)

	for _, name := range names {
		if name == champion.Name {
			championMatchups[name] = 0
		} else {
			championMatchups[name] = rand.Intn(11) - 5
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
