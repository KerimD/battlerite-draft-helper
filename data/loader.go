package data

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ChampionsCsvFilename          = "champions.csv"
	ChampionsShortListCsvFilename = "champions-short-list.csv"
	MatchupsCsvFilename           = "matchups.csv"
	SynergiesCsvFilename          = "synergies.csv"
	MatchupsJsonFilename          = "matchups.json"
)

type Champion struct {
	Name string
	Role string
	Id   string
}

func main() {
	start := time.Now()

	//champions := GetChampionsFromCsv(ChampionsCsvFilename)

	//createMatchupsCsv(champions, "")
	//createMatchupsJson(champions)
	migrateMatchupsCsvToJson()

	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func GetChampionsFromCsv(filename string) []Champion {
	data := getCsvData(filename)

	var champions []Champion
	for i, row := range data {
		if i == 0 {
			continue
		}

		champions = append(champions, Champion{
			Name: row[0],
			Role: row[1],
			Id:   row[2],
		})
	}

	return champions
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

func migrateMatchupsCsvToJson() {
	data := getCsvData(MatchupsCsvFilename)

	matchups := make(map[string]map[string]int)
	for i, row := range data {
		if i == 0 {
			continue
		}
		populateMatchups(row, matchups)
	}

	sortedMatchups := sortMatchups(matchups)
	SaveDataToJsonFile(sortedMatchups, MatchupsJsonFilename)
}

func populateMatchups(row []string, matchups map[string]map[string]int) {
	name := row[0]
	opposition := row[1]
	evaluation, err := strconv.Atoi(row[2])
	if err != nil {
		log.Fatal(err)
	}

	populateMatchup(matchups, name, opposition, evaluation)
	populateMatchup(matchups, opposition, name, 6-evaluation)
}

func populateMatchup(matchups map[string]map[string]int, name string, opposition string, evaluation int) {
	_, exists := matchups[name]
	if !exists {
		matchups[name] = make(map[string]int)
		matchups[name][name] = 3
	}
	matchups[name][opposition] = evaluation
}

// no clue how this function works
func sortMatchups(matchups map[string]map[string]int) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage)

	for champion, opponents := range matchups {
		type pair struct {
			key   string
			value int
		}

		var pairs []pair
		for opponent, evaluation := range opponents {
			pairs = append(pairs, pair{opponent, evaluation})
		}

		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].value > pairs[j].value
		})

		var jsonParts []string
		for _, p := range pairs {
			jsonParts = append(jsonParts, fmt.Sprintf(`"%s":%d`, p.key, p.value))
		}

		jsonStr := "{" + strings.Join(jsonParts, ",") + "}"
		result[champion] = json.RawMessage(jsonStr)
	}

	return result
}

func SaveDataToJsonFile(data any, filename string) {
	f, err := os.Create(filename)
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

func getCsvData(filename string) [][]string {
	f, err := os.Open(filename)
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
	return records
}
