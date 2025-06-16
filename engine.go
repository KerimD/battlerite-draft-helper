package main

import (
	"battlerite-draft-helper/data"
	"fmt"
)

type ScoredTrieNode struct {
	Champion          data.Champion
	AverageEvaluation int
	CompleteStates    []string
	Children          map[string]*ScoredTrieNode
}

var DraftOrder = []string{"T1GB", "T2GB", "T2B", "T1B", "T1P", "T2P", "T2P", "T1P", "T1B", "T2B", "T1P", "T2P"}

func main() {
	champions := data.GetChampionsFromCsv()
	ChampionToID, IDToChampion := initializeChampionMaps(champions)

	vroomVroom(champions, ChampionToID, IDToChampion)
}

func initializeChampionMaps(champions []data.Champion) (map[string]string, map[string]string) {
	ChampionToID := make(map[string]string, len(champions))
	IDToChampion := make(map[string]string, len(champions))

	for i, champion := range champions {
		var id string
		if i < 10 {
			id = fmt.Sprintf("%d", i)
		} else {
			id = string(rune('A' + i - 10))
		}

		ChampionToID[champion.Name] = id
		IDToChampion[id] = champion.Name
	}

	return ChampionToID, IDToChampion
}

func vroomVroom(champions []data.Champion, ChampionToID map[string]string, IDToChampion map[string]string) {
	node := ScoredTrieNode{
		Champion:          data.Champion{},
		AverageEvaluation: 0,
		CompleteStates:    []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}

	evaluationSum := 0
	for _, champion := range champions {
		// kick off a process for each champion getting first chosen
		evaluation, completedStates := process(champions, ChampionToID, IDToChampion, champion, 0)

		evaluationSum += evaluation
		node.CompleteStates = append(node.CompleteStates, completedStates...)
	}

	node.AverageEvaluation = evaluationSum / len(champions)
}

func process(
	champions []data.Champion,
	ChampionToID map[string]string,
	IDToChampion map[string]string,
	chosenChampion data.Champion,
	currentDraftStep int,
) (int, []string) {
	// hit base case (draft step > draft order len)
	// return empty array

	node := ScoredTrieNode{
		Champion:          chosenChampion,
		AverageEvaluation: 0,
		CompleteStates:    []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}
	/*
	 *
	 *
	 *
	 *
	 *
	 *
	 *
	 *
	 *
	 *
	 */
	return node.AverageEvaluation, node.CompleteStates
}
