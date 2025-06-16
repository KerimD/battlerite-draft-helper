package main

import (
	"battlerite-draft-helper/data"
	"fmt"
	"log"
)

type ScoredTrieNode struct {
	Champion          string
	AverageEvaluation int
	CompletedStates   []string
	Children          map[string]*ScoredTrieNode
}

const DataDir = "data/"

var DraftOrder = []string{"T1GB", "T2GB", "T2B", "T1B", "T1P", "T2P", "T2P", "T1P", "T1B", "T2B", "T1P", "T2P"}

func main() {
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	championNamesSet := make(map[string]bool, len(champions))
	for _, champion := range champions {
		championNamesSet[champion.Name] = true
	}
	ChampionToID, IDToChampion := initializeChampionMaps(championNamesSet)

	vroomVroom(championNamesSet, ChampionToID, IDToChampion)
}

func initializeChampionMaps(champions map[string]bool) (map[string]string, map[string]string) {
	ChampionToID := make(map[string]string, len(champions))
	IDToChampion := make(map[string]string, len(champions))

	i := 0
	for champion := range champions {
		var id string
		if i < 10 {
			id = fmt.Sprintf("%d", i)
		} else {
			id = string(rune('A' + i - 10))
		}

		ChampionToID[champion] = id
		IDToChampion[id] = champion
		i++
	}

	return ChampionToID, IDToChampion
}

func vroomVroom(champions map[string]bool, ChampionToID map[string]string, IDToChampion map[string]string) {
	node := ScoredTrieNode{
		Champion:          "",
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}

	evaluationSum := 0
	for champion := range champions {
		// kick off a process for each champion getting first chosen
		evaluation, completedStates := process(
			champions,
			ChampionToID,
			IDToChampion,
			"",
			champion,
			0,
			copyMap(champions),
			copyMap(champions),
			copyMap(champions),
			copyMap(champions),
		)
		fmt.Println("Processed: ", champion, " with evaluation: ", evaluation, " and num completed states: ", len(completedStates))

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)
	}

	node.AverageEvaluation = evaluationSum / len(champions)
}

func process(
	champions map[string]bool,
	ChampionToID map[string]string,
	IDToChampion map[string]string,
	previousState string,
	chosenChampion string,
	currentDraftStep int,
	T1PickableChampions map[string]bool,
	T1BannableChampions map[string]bool,
	T2PickableChampions map[string]bool,
	T2BannableChampions map[string]bool,
) (int, []string) {
	currentState := previousState + ChampionToID[chosenChampion]

	// Base case
	if currentDraftStep >= len(DraftOrder) {
		evaluation := 9
		// TODO: Evaluate completed state.
		// TODO: Prune if no support.
		return evaluation, []string{currentState}
	}

	node := ScoredTrieNode{
		Champion:          chosenChampion,
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}

	selectableChampions := getSelectableChampions(
		currentDraftStep,
		T1PickableChampions,
		T1BannableChampions,
		T2PickableChampions,
		T2BannableChampions,
	)

	evaluationSum := 0
	for champion := range selectableChampions {
		copyT1PickableChampions := copyMap(T1PickableChampions)
		copyT1BannableChampions := copyMap(T1BannableChampions)
		copyT2PickableChampions := copyMap(T2PickableChampions)
		copyT2BannableChampions := copyMap(T2BannableChampions)

		updateChampionSlicesInPlace(
			currentDraftStep,
			copyT1PickableChampions,
			copyT1BannableChampions,
			copyT2PickableChampions,
			copyT2BannableChampions,
			champion,
		)

		evaluation, completedStates := process(
			champions,
			ChampionToID,
			IDToChampion,
			currentState,
			champion,
			currentDraftStep+1,
			copyT1PickableChampions,
			copyT1BannableChampions,
			copyT2PickableChampions,
			copyT2BannableChampions,
		)

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)
	}

	node.AverageEvaluation = evaluationSum / len(champions)
	return node.AverageEvaluation, node.CompletedStates
}

func getSelectableChampions(
	currentDraftStep int,
	T1PickableChampions map[string]bool,
	T1BannableChampions map[string]bool,
	T2PickableChampions map[string]bool,
	T2BannableChampions map[string]bool,
) map[string]bool {
	switch DraftOrder[currentDraftStep] {
	case "T1P":
		return T1PickableChampions
	case "T1GB", "T1B":
		return T1BannableChampions
	case "T2P":
		return T2PickableChampions
	case "T2GB", "T2B":
		return T2BannableChampions
	default:
		log.Fatal("Invalid draft step.")
		return nil
	}
}

func updateChampionSlicesInPlace(
	currentDraftStep int,
	T1PickableChampions map[string]bool,
	T1BannableChampions map[string]bool,
	T2PickableChampions map[string]bool,
	T2BannableChampions map[string]bool,
	champion string,
) {
	switch DraftOrder[currentDraftStep] {
	case "T1GB":
		delete(T1PickableChampions, champion)
		delete(T1BannableChampions, champion)
		delete(T2PickableChampions, champion)
		delete(T2BannableChampions, champion)
		break
	case "T1B":
		delete(T1BannableChampions, champion)
		delete(T2PickableChampions, champion)
		break
	case "T1P":
		delete(T1PickableChampions, champion)
		delete(T2BannableChampions, champion)
		break
	case "T2GB":
		delete(T1PickableChampions, champion)
		delete(T1BannableChampions, champion)
		delete(T2PickableChampions, champion)
		delete(T2BannableChampions, champion)
		break
	case "T2B":
		delete(T1PickableChampions, champion)
		delete(T2BannableChampions, champion)
		break
	case "T2P":
		delete(T2PickableChampions, champion)
		delete(T1BannableChampions, champion)
		break
	default:
		log.Fatal("Invalid draft step.")
	}
}
