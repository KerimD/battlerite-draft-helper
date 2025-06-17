package main

import (
	"battlerite-draft-helper/data"
	"fmt"
	"log"
)

type ScoredTrieNode struct {
	championName      string
	AverageEvaluation int
	CompletedStates   []string
	Children          map[string]*ScoredTrieNode
}

const SupportRole = "Support"

const DataDir = "data/"

var DraftOrder = []string{"T1GB1", "T2GB1", "T2B2", "T1B2", "T1P1", "T2P1", "T2P2", "T1P2", "T1B3", "T2B3", "T1P3", "T2P3"}

func main() {
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	ChampionToID, IDToChampion := initializeChampionMaps(champions)

	championSet := make(map[string]data.Champion, len(champions))
	for _, champion := range champions {
		championSet[champion.ID] = champion
	}

	vroomVroom(championSet, ChampionToID, IDToChampion)
}

func initializeChampionMaps(champions []data.Champion) (map[string]string, map[string]string) {
	ChampionNameToID := make(map[string]string, len(champions))
	IDToChampionName := make(map[string]string, len(champions))

	for _, champion := range champions {
		ChampionNameToID[champion.Name] = champion.ID
		IDToChampionName[champion.ID] = champion.Name
	}

	return ChampionNameToID, IDToChampionName
}

func vroomVroom(champions map[string]data.Champion, ChampionNameToID map[string]string, IDToChampionName map[string]string) {
	node := ScoredTrieNode{
		championName:      "",
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}

	evaluationSum := 0
	for championId, champion := range champions {
		// kick off a process for each champion getting first chosen
		evaluation, completedStates := process(
			len(champions),
			ChampionNameToID,
			IDToChampionName,
			"",
			championId,
			0,
			false,
			false,
			0,
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
	numChampions int,
	ChampionNameToID map[string]string,
	IDToChampionName map[string]string,
	previousState string,
	chosenChampionId string,
	currDraftStepIdx int,
	T1HasSupport bool,
	T2HasSupport bool,
	numT1Picks int,
	numT2Picks int,
	T1PickableChampions map[string]bool,
	T1BannableChampions map[string]bool,
	T2PickableChampions map[string]bool,
	T2BannableChampions map[string]bool,
) (int, []string) {
	currentState := previousState + chosenChampionId

	// Base case
	if currDraftStepIdx >= len(DraftOrder) {
		evaluation := 9
		// TODO: Evaluate completed state.
		return evaluation, []string{currentState}
	}

	node := ScoredTrieNode{
		championName:      IDToChampionName[chosenChampionId],
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}

	selectableChampions := getSelectableChampions(
		currDraftStepIdx,
		T1HasSupport,
		T2HasSupport,
		numT1Picks,
		numT2Picks,
		T1PickableChampions,
		T1BannableChampions,
		T2PickableChampions,
		T2BannableChampions,
	)

	evaluationSum := 0
	for championId, champion := range selectableChampions {

		copyT1PickableChampions := copyMap(T1PickableChampions)
		copyT1BannableChampions := copyMap(T1BannableChampions)
		copyT2PickableChampions := copyMap(T2PickableChampions)
		copyT2BannableChampions := copyMap(T2BannableChampions)

		numT1Picks, numT2Picks = updateSelectableChampionsInPlace(
			championId,
			currDraftStepIdx,
			numT1Picks,
			numT2Picks,
			copyT1PickableChampions,
			copyT1BannableChampions,
			copyT2PickableChampions,
			copyT2BannableChampions,
		)

		evaluation, completedStates := process(
			numChampions,
			ChampionNameToID,
			IDToChampionName,
			currentState,
			championId,
			currDraftStepIdx+1,
			T1HasSupport || DraftOrder[currDraftStepIdx][1] == '1' || champion.Role == SupportRole,
			T2HasSupport || DraftOrder[currDraftStepIdx][1] == '2' || champion.Role == SupportRole,
			numT1Picks,
			numT2Picks,
			copyT1PickableChampions,
			copyT1BannableChampions,
			copyT2PickableChampions,
			copyT2BannableChampions,
		)

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)
	}

	node.AverageEvaluation = evaluationSum / numChampions
	return node.AverageEvaluation, node.CompletedStates
}

func getSelectableChampions(
	currDraftStepIdx int,
	T1HasSupport bool,
	T2HasSupport bool,
	numT1Picks int,
	numT2Picks int,
	T1PickableChampions map[string]bool,
	T1BannableChampions map[string]bool,
	T2PickableChampions map[string]bool,
	T2BannableChampions map[string]bool,
) map[string]data.Champion {
	/*
	 * TODO: Prune child nodes that don't have at least 1 support in the comp
	 * if T1 is picking
	 * and T1 has already made 2 picks
	 * and T1 has no sups
	 * return only supports
	 *
	 */

	switch DraftOrder[currDraftStepIdx] {
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

func updateSelectableChampionsInPlace(
	champion string,
	currDraftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	T1PickableChampions map[string]bool,
	T1BannableChampions map[string]bool,
	T2PickableChampions map[string]bool,
	T2BannableChampions map[string]bool,
) (int, int) {
	switch DraftOrder[currDraftStepIdx] {
	case "T1GB":
		delete(T1PickableChampions, champion)
		delete(T1BannableChampions, champion)
		delete(T2PickableChampions, champion)
		delete(T2BannableChampions, champion)
		return numT1Picks, numT2Picks
	case "T1B":
		delete(T1BannableChampions, champion)
		delete(T2PickableChampions, champion)
		return numT1Picks, numT2Picks
	case "T1P":
		delete(T1PickableChampions, champion)
		delete(T2BannableChampions, champion)
		return numT1Picks + 1, numT2Picks
	case "T2GB":
		delete(T1PickableChampions, champion)
		delete(T1BannableChampions, champion)
		delete(T2PickableChampions, champion)
		delete(T2BannableChampions, champion)
		return numT1Picks, numT2Picks
	case "T2B":
		delete(T1PickableChampions, champion)
		delete(T2BannableChampions, champion)
		return numT1Picks, numT2Picks
	case "T2P":
		delete(T2PickableChampions, champion)
		delete(T1BannableChampions, champion)
		return numT1Picks, numT2Picks + 1
	default:
		log.Fatal("Invalid draft step.")
		return numT1Picks, numT2Picks
	}
}
