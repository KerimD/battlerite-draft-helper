package main

import (
	"battlerite-draft-helper/data"
	"fmt"
	"log"
)

type Role string

const (
	MeleeRole   Role = "Melee"
	RangedRole  Role = "Ranged"
	SupportRole Role = "Support"
)

type ScoredTrieNode struct {
	championName      string
	AverageEvaluation int
	CompletedStates   []string
	Children          map[string]*ScoredTrieNode
}

type TeamSelectableChampions struct {
	PickableChampions map[Role]map[string]bool
	BannableChampions map[Role]map[string]bool
}

const DataDir = "data/"

var DraftOrder = []string{"T1GB1", "T2GB1", "T2B2", "T1B2", "T1P1", "T2P1", "T2P2", "T1P2", "T1B3", "T2B3", "T1P3", "T2P3"}

func main() {
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	ChampionToID, IdToChampion := initializeChampionMaps(champions)

	championSet := make(map[string]data.Champion, len(champions))
	for _, champion := range champions {
		championSet[champion.Id] = champion
	}

	vroomVroom(championSet, ChampionToID, IdToChampion)
}

func initializeChampionMaps(champions []data.Champion) (map[string]string, map[string]data.Champion) {
	ChampionNameToId := make(map[string]string, len(champions))
	IdToChampion := make(map[string]data.Champion, len(champions))

	for _, champion := range champions {
		ChampionNameToId[champion.Name] = champion.Id
		IdToChampion[champion.Id] = champion
	}

	return ChampionNameToId, IdToChampion
}

func vroomVroom(champions map[string]data.Champion, ChampionNameToId map[string]string, IdToChampion map[string]data.Champion) {
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
			ChampionNameToId,
			IdToChampion,
			"",
			championId,
			0,
			0,
			0,
			false,
			false,
			createTeamSelectableChampions(champions),
			createTeamSelectableChampions(champions),
		)
		fmt.Println("Processed: ", champion, " with evaluation: ", evaluation, " and num completed states: ", len(completedStates))

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)
	}

	node.AverageEvaluation = evaluationSum / len(champions)
}

func createTeamSelectableChampions(champions map[string]data.Champion) TeamSelectableChampions {
	teamSelectableChampions := TeamSelectableChampions{
		PickableChampions: make(map[Role]map[string]bool),
		BannableChampions: make(map[Role]map[string]bool),
	}

	teamSelectableChampions.PickableChampions[MeleeRole] = make(map[string]bool)
	teamSelectableChampions.PickableChampions[RangedRole] = make(map[string]bool)
	teamSelectableChampions.PickableChampions[SupportRole] = make(map[string]bool)

	teamSelectableChampions.BannableChampions[MeleeRole] = make(map[string]bool)
	teamSelectableChampions.BannableChampions[RangedRole] = make(map[string]bool)
	teamSelectableChampions.BannableChampions[SupportRole] = make(map[string]bool)

	for championId, champion := range champions {
		role := Role(champion.Role)
		teamSelectableChampions.PickableChampions[role][championId] = true
		teamSelectableChampions.BannableChampions[role][championId] = true
	}

	return teamSelectableChampions
}

func process(
	numChampions int,
	ChampionNameToId map[string]string,
	IdToChampion map[string]data.Champion,
	previousState string,
	chosenChampionId string,
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1HasSupport bool,
	t2HasSupport bool,
	t1SelectableChampions TeamSelectableChampions,
	t2SelectableChampions TeamSelectableChampions,
) (int, []string) {
	currentState := previousState + chosenChampionId

	// Base case
	if draftStepIdx >= len(DraftOrder) {
		evaluation := 9
		// TODO: Evaluate completed state.
		return evaluation, []string{currentState}
	}

	node := ScoredTrieNode{
		championName:      IdToChampion[chosenChampionId].Name,
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*ScoredTrieNode),
	}

	selectableChampions := getSelectableChampions(
		t1SelectableChampions,
		t2SelectableChampions,
		draftStepIdx,
	)

	evaluationSum := 0
	for championId := range selectableChampions {

		deepCopyT1SelectableChampions := deepCopyTeamSelectableChampions(t1SelectableChampions)
		deepCopyT2SelectableChampions := deepCopyTeamSelectableChampions(t2SelectableChampions)

		numT1Picks, numT2Picks = updateSelectableChampionsInPlace(
			IdToChampion,
			championId,
			draftStepIdx,
			numT1Picks,
			numT2Picks,
			deepCopyT1SelectableChampions,
			deepCopyT2SelectableChampions,
		)

		evaluation, completedStates := process(
			numChampions,
			ChampionNameToId,
			IdToChampion,
			currentState,
			championId,
			draftStepIdx+1,
			numT1Picks,
			numT2Picks,
			t1HasSupport || (DraftOrder[draftStepIdx][1] == '1' && Role(IdToChampion[championId].Role) == SupportRole),
			t2HasSupport || (DraftOrder[draftStepIdx][1] == '2' && Role(IdToChampion[championId].Role) == SupportRole),
			deepCopyT1SelectableChampions,
			deepCopyT2SelectableChampions,
		)

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)
	}

	node.AverageEvaluation = evaluationSum / numChampions
	return node.AverageEvaluation, node.CompletedStates
}

func getSelectableChampions(
	team1SelectableChampions TeamSelectableChampions,
	team2SelectableChampions TeamSelectableChampions,
	draftStepIdx int,
) map[string]bool {
	/*
	 * TODO: Prune child nodes that don't have at least 1 support in the comp
	 * if T1 is picking
	 * and T1 has already made 2 picks
	 * and T1 has no sups
	 * return only supports
	 *
	 */

	switch DraftOrder[draftStepIdx] {
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

func deepCopyTeamSelectableChampions(teamSelectableChampions TeamSelectableChampions) TeamSelectableChampions {
	deepCopy := TeamSelectableChampions{
		PickableChampions: make(map[Role]map[string]bool),
		BannableChampions: make(map[Role]map[string]bool),
	}

	deepCopy.PickableChampions[MeleeRole] = copyMap(teamSelectableChampions.PickableChampions[MeleeRole])
	deepCopy.PickableChampions[RangedRole] = copyMap(teamSelectableChampions.PickableChampions[MeleeRole])
	deepCopy.PickableChampions[SupportRole] = copyMap(teamSelectableChampions.PickableChampions[MeleeRole])

	deepCopy.BannableChampions[MeleeRole] = copyMap(teamSelectableChampions.PickableChampions[MeleeRole])
	deepCopy.BannableChampions[RangedRole] = copyMap(teamSelectableChampions.PickableChampions[MeleeRole])
	deepCopy.BannableChampions[SupportRole] = copyMap(teamSelectableChampions.PickableChampions[MeleeRole])

	return deepCopy
}

func updateSelectableChampionsInPlace(
	IdToChampion map[string]data.Champion,
	championId string,
	currDraftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1SelectableChampions TeamSelectableChampions,
	t2SelectableChampions TeamSelectableChampions,
) (int, int) {
	champion := IdToChampion[championId]
	role := Role(champion.Role)

	switch DraftOrder[currDraftStepIdx] {
	case "T1GB":
		delete(t1SelectableChampions.PickableChampions[role], championId)
		delete(t2SelectableChampions.BannableChampions[role], championId)
		fallthrough
	case "T1B":
		delete(t1SelectableChampions.BannableChampions[role], championId)
		delete(t2SelectableChampions.PickableChampions[role], championId)
		return numT1Picks, numT2Picks
	case "T1P":
		delete(t1SelectableChampions.PickableChampions[role], championId)
		delete(t2SelectableChampions.BannableChampions[role], championId)
		return numT1Picks + 1, numT2Picks
	case "T2GB":
		delete(t1SelectableChampions.BannableChampions[role], championId)
		delete(t2SelectableChampions.PickableChampions[role], championId)
		fallthrough
	case "T2B":
		delete(t1SelectableChampions.PickableChampions[role], championId)
		delete(t2SelectableChampions.BannableChampions[role], championId)
		return numT1Picks, numT2Picks
	case "T2P":
		delete(t2SelectableChampions.PickableChampions[role], championId)
		delete(t1SelectableChampions.BannableChampions[role], championId)
		return numT1Picks, numT2Picks + 1
	default:
		log.Fatal("Invalid draft step.")
		return numT1Picks, numT2Picks
	}
}
