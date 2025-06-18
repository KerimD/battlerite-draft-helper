package main

import (
	"battlerite-draft-helper/c"
	"battlerite-draft-helper/data"
	"fmt"
	"log"
)

const DataDir = "data/"

func main() {
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	ChampionToID, IdToChampion := initializeChampionMaps(champions)

	championSet := make(map[string]c.Champion, len(champions))
	for _, champion := range champions {
		championSet[champion.Id] = champion
	}

	vroomVroom(championSet, ChampionToID, IdToChampion)
}

func initializeChampionMaps(champions []c.Champion) (map[string]string, map[string]c.Champion) {
	ChampionNameToId := make(map[string]string, len(champions))
	IdToChampion := make(map[string]c.Champion, len(champions))

	for _, champion := range champions {
		ChampionNameToId[champion.Name] = champion.Id
		IdToChampion[champion.Id] = champion
	}

	return ChampionNameToId, IdToChampion
}

func vroomVroom(champions map[string]c.Champion, ChampionNameToId map[string]string, IdToChampion map[string]c.Champion) {
	node := c.ScoredTrieNode{
		ChampionName:      "",
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*c.ScoredTrieNode),
	}

	evaluationSum := 0
	for championId, champion := range champions {
		t1SelectableChampions := c.CreateTeamSelectableChampions(champions)
		t2SelectableChampions := c.CreateTeamSelectableChampions(champions)

		numT1Picks, numT2Picks := updateSelectableChampionsInPlace(
			championId,
			0,
			0,
			0,
			t1SelectableChampions,
			t2SelectableChampions,
		)

		// kick off a process for each champion getting first chosen
		evaluation, completedStates := process(
			len(champions),
			ChampionNameToId,
			IdToChampion,
			"",
			championId,
			0,
			numT1Picks,
			numT2Picks,
			false,
			false,
			t1SelectableChampions,
			t2SelectableChampions,
		)
		fmt.Println("Processed: ", champion, " with evaluation: ", evaluation, " and num completed states: ", len(completedStates))

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)

		//c.FormatCompletedStates(node.CompletedStates[533116:533119], IdToChampion)
		//c.FormatCompletedStates(node.CompletedStates, IdToChampion)
		//fmt.Println(node.CompletedStates)
		//break
	}

	node.AverageEvaluation = evaluationSum / len(champions)
}

func process(
	numChampions int,
	championNameToId map[string]string,
	idToChampion map[string]c.Champion,
	previousState string,
	chosenChampionId string,
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1HasSupport bool,
	t2HasSupport bool,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) (int, []string) {
	currentState := previousState + chosenChampionId

	// Base case
	if draftStepIdx >= len(c.DraftOrder)-1 {
		evaluation := 9
		// TODO: Evaluate completed state.
		return evaluation, []string{currentState}
	}

	node := c.ScoredTrieNode{
		ChampionName:      idToChampion[chosenChampionId].Name,
		AverageEvaluation: 0,
		CompletedStates:   []string{},
		Children:          make(map[string]*c.ScoredTrieNode),
	}

	selectableChampions := getSelectableChampions(
		draftStepIdx+1,
		numT1Picks,
		numT2Picks,
		t1HasSupport,
		t2HasSupport,
		t1SelectableChampions,
		t2SelectableChampions,
	)

	evaluationSum := 0
	for championId := range selectableChampions {

		deepCopyT1SelectableChampions := c.DeepCopyTeamSelectableChampions(t1SelectableChampions)
		deepCopyT2SelectableChampions := c.DeepCopyTeamSelectableChampions(t2SelectableChampions)

		if draftStepIdx < len(c.DraftOrder)-1 {
			numT1Picks, numT2Picks = updateSelectableChampionsInPlace(
				championId,
				draftStepIdx+1,
				numT1Picks,
				numT2Picks,
				deepCopyT1SelectableChampions,
				deepCopyT2SelectableChampions,
			)
		}

		evaluation, completedStates := process(
			numChampions,
			championNameToId,
			idToChampion,
			currentState,
			championId,
			draftStepIdx+1,
			numT1Picks,
			numT2Picks,
			t1HasSupport || (c.DraftOrder[draftStepIdx+1] == "T1P" && idToChampion[championId].Role == c.SupportRole),
			t2HasSupport || (c.DraftOrder[draftStepIdx+1] == "T2P" && idToChampion[championId].Role == c.SupportRole),
			deepCopyT1SelectableChampions,
			deepCopyT2SelectableChampions,
		)

		evaluationSum += evaluation
		node.CompletedStates = append(node.CompletedStates, completedStates...)

		if draftStepIdx < 3 {
			fmt.Println(currentState, " -> ", idToChampion[championId], "num completed states: ", len(completedStates))
		}
		//break
	}

	node.AverageEvaluation = evaluationSum / numChampions
	return node.AverageEvaluation, node.CompletedStates
}

func getSelectableChampions(
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1HasSupport bool,
	t2HasSupport bool,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) map[string]bool {
	t1NeedsSupportThisStep := !t1HasSupport && numT1Picks >= 2
	t2NeedsSupportThisStep := !t2HasSupport && numT2Picks >= 2

	switch c.DraftOrder[draftStepIdx] {
	case "T1P":
		if t1NeedsSupportThisStep {
			return t1SelectableChampions.PickableSupportChampions
		}
		return t1SelectableChampions.PickableChampions
	case "T1GB", "T1B":
		if t2NeedsSupportThisStep {
			return t1SelectableChampions.BannableSupportChampions
		}
		return t1SelectableChampions.BannableChampions
	case "T2P":
		if t2NeedsSupportThisStep {
			return t2SelectableChampions.PickableSupportChampions
		}
		return t2SelectableChampions.PickableChampions
	case "T2GB", "T2B":
		if t1NeedsSupportThisStep {
			return t2SelectableChampions.BannableSupportChampions
		}
		return t2SelectableChampions.BannableChampions
	default:
		log.Fatal("getSelectableChampions(): Invalid draft step.")
		return nil
	}
}

// Note: Currently does not support global bans after picks.
func updateSelectableChampionsInPlace(
	championId string,
	currDraftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) (int, int) {
	switch c.DraftOrder[currDraftStepIdx] {
	case "T1GB":
		delete(t1SelectableChampions.PickableChampions, championId)
		delete(t2SelectableChampions.BannableChampions, championId)
		delete(t1SelectableChampions.PickableSupportChampions, championId)
		delete(t2SelectableChampions.BannableSupportChampions, championId)
		fallthrough
	case "T1B":
		delete(t1SelectableChampions.BannableChampions, championId)
		delete(t2SelectableChampions.PickableChampions, championId)
		delete(t1SelectableChampions.BannableSupportChampions, championId)
		delete(t2SelectableChampions.PickableSupportChampions, championId)
		return numT1Picks, numT2Picks
	case "T1P":
		delete(t1SelectableChampions.PickableChampions, championId)
		delete(t2SelectableChampions.BannableChampions, championId)
		delete(t1SelectableChampions.PickableSupportChampions, championId)
		delete(t2SelectableChampions.BannableSupportChampions, championId)
		return numT1Picks + 1, numT2Picks
	case "T2GB":
		delete(t1SelectableChampions.BannableChampions, championId)
		delete(t2SelectableChampions.PickableChampions, championId)
		delete(t1SelectableChampions.BannableSupportChampions, championId)
		delete(t2SelectableChampions.PickableSupportChampions, championId)
		fallthrough
	case "T2B":
		delete(t1SelectableChampions.PickableChampions, championId)
		delete(t2SelectableChampions.BannableChampions, championId)
		delete(t1SelectableChampions.PickableSupportChampions, championId)
		delete(t2SelectableChampions.BannableSupportChampions, championId)
		return numT1Picks, numT2Picks
	case "T2P":
		delete(t1SelectableChampions.BannableChampions, championId)
		delete(t2SelectableChampions.PickableChampions, championId)
		delete(t1SelectableChampions.BannableSupportChampions, championId)
		delete(t2SelectableChampions.PickableSupportChampions, championId)
		return numT1Picks, numT2Picks + 1
	default:
		log.Fatal("updateSelectableChampionsInPlace(): Invalid draft step.")
		return numT1Picks, numT2Picks
	}
}
