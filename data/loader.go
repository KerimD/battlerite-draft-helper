package data

import (
	"battlerite-draft-helper/c"
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
	DataDir                       = "data/"
	PlayerDataDir                 = "player-data/"
)

func main() {
	start := time.Now()

	//champions := GetChampionsFromCsv(ChampionsCsvFilename)

	//createMatchupsCsv(champions, "")
	//createMatchupsJson(champions)
	migrateMatchupsCsvToJson()

	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func GetChampionsFromCsv(filename string) []c.Champion {
	data := getCsvData(filename)

	var champions []c.Champion
	for i, row := range data {
		if i == 0 {
			continue
		}

		champions = append(champions, c.Champion{
			Name: row[0],
			Role: c.Role(row[1]),
			Id:   byte(i - 1),
		})
	}

	return champions
}

func FormatCsvData(championNameToId map[string]byte, filename string, isSynergy bool) map[byte]map[byte]int8 {
	data := getCsvData(filename)
	stringData := parseCsvData(data, isSynergy)

	byteData := make(map[byte]map[byte]int8)
	for name, nameId := range championNameToId {
		for opposition, oppositionId := range championNameToId {
			if byteData[nameId] == nil {
				byteData[nameId] = make(map[byte]int8)
			}
			byteData[nameId][oppositionId] = stringData[name][opposition]
		}
	}
	return byteData
}

func GetPlayerChampions(
	championNameToId map[string]byte,
	p1Name string,
	p2Name string,
	p3Name string,
) (c.Player, c.Player, c.Player) {
	playersChampionPoolData := [][][]string{
		getCsvData(DataDir + PlayerDataDir + p1Name + ".csv"),
		getCsvData(DataDir + PlayerDataDir + p2Name + ".csv"),
		getCsvData(DataDir + PlayerDataDir + p3Name + ".csv"),
	}

	player1 := c.Player{Name: p1Name, ChampionPool: make(map[byte]int8, len(playersChampionPoolData[0])-1)}
	player2 := c.Player{Name: p2Name, ChampionPool: make(map[byte]int8, len(playersChampionPoolData[1])-1)}
	player3 := c.Player{Name: p3Name, ChampionPool: make(map[byte]int8, len(playersChampionPoolData[2])-1)}

	for i, player := range []c.Player{player1, player2, player3} {
		for j, row := range playersChampionPoolData[i] {
			if j == 0 {
				continue
			}
			// TODO: Remove when using full champions list
			if id, exists := championNameToId[row[0]]; exists {
				evaluation, err := strconv.ParseInt(row[1], 10, 8)
				if err != nil {
					log.Fatal(err)
				}
				player.ChampionPool[id] = int8(evaluation)
			}
		}
	}

	return player1, player2, player3
}

func createMatchupsCsv(champions []c.Champion, filename string) {
	matchups := make([][]string, 0, (len(champions)*(len(champions)-1))/2)
	matchups = append(matchups, []string{"name", "opposition", "evaluation"})

	for i, champion := range champions {
		for _, opposition := range champions[i+1:] {
			evaluation := "0"
			if filename == SynergiesCsvFilename && champion.Role == "Melee" && champion.Role == opposition.Role {
				evaluation = "-1"
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
	matchups := parseCsvData(data, false)
	sortedMatchups := sortMatchups(matchups)
	SaveDataToJsonFile(sortedMatchups, MatchupsJsonFilename)
}

func parseCsvData(data [][]string, isSynergy bool) map[string]map[string]int8 {
	matchups := make(map[string]map[string]int8)
	for i, row := range data {
		if i == 0 {
			continue
		}
		populateMatchups(row, matchups, isSynergy)
	}
	return matchups
}

func populateMatchups(row []string, matchups map[string]map[string]int8, isSynergy bool) {
	name := row[0]
	opposition := row[1]
	intEvaluation, err := strconv.Atoi(row[2])
	if err != nil {
		log.Fatal(err)
	}
	evaluation := int8(intEvaluation)

	populateMatchup(matchups, name, opposition, evaluation)
	if isSynergy {
		populateMatchup(matchups, opposition, name, evaluation)
	} else {
		populateMatchup(matchups, opposition, name, (-1)*evaluation)
	}
}

func populateMatchup(matchups map[string]map[string]int8, name string, opposition string, evaluation int8) {
	_, exists := matchups[name]
	if !exists {
		matchups[name] = make(map[string]int8)
		matchups[name][name] = 0
	}
	matchups[name][opposition] = evaluation
}

// no clue how this function works
func sortMatchups(matchups map[string]map[string]int8) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage)

	for champion, opponents := range matchups {
		type pair struct {
			key   string
			value int8
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
