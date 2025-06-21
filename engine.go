package main

import (
	"battlerite-draft-helper/c"
	"battlerite-draft-helper/data"
	"fmt"
	"time"

	//"fmt"
	"log"
)

const DataDir = "data/"

var (
	NumChampions     = 0
	ChampionNameToId = make(map[string]byte)
	IdToChampion     = make(map[byte]c.Champion)
	ChampionMatchups = make(map[byte]map[byte]int)
	ChampionSynergys = make(map[byte]map[byte]int)
)

// Note: Currently does not support global bans after picks.
func main() {
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	//champions := data.GetChampionsFromCsv(DataDir + data.ChampionsCsvFilename)
	initializeGlobalVariables(champions)

	championSet := make(map[byte]c.Champion, len(champions))
	for _, champion := range champions {
		championSet[champion.Id] = champion
	}

	start := time.Now()
	vroomVroom(championSet)
	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func initializeGlobalVariables(champions []c.Champion) {
	NumChampions = len(champions)
	for _, champion := range champions {
		ChampionNameToId[champion.Name] = champion.Id
		IdToChampion[champion.Id] = champion
	}
	ChampionMatchups = data.FormatCsvData(ChampionNameToId, DataDir+data.MatchupsCsvFilename, false)
	ChampionSynergys = data.FormatCsvData(ChampionNameToId, DataDir+data.SynergiesCsvFilename, true)
}

func vroomVroom(championSet map[byte]c.Champion) {
	node := c.ScoredTrieNode{
		ChampionName:      "",
		AverageEvaluation: 0,
		Children:          make(map[byte]*c.ScoredTrieNode),
	}

	tempNumChampionsRan := 0
	numCompletedStates := 0
	evaluationSum := float32(0)
	for championId, _ := range championSet {
		if championId != 6 {
			continue
		}

		tempNumChampionsRan += 1

		evaluation, childNode, asdf := kickOffDraft(championSet, championId)

		evaluationSum += evaluation
		node.Children[championId] = childNode
		numCompletedStates += asdf
		break
	}

	//node.AverageEvaluation = evaluationSum / float32(len(championSet))
	node.AverageEvaluation = evaluationSum / float32(tempNumChampionsRan)
	fmt.Println("numCompletedStates 6350400:", numCompletedStates)
	fmt.Println("Team 1:", node.AverageEvaluation)
	c.PrintTree(&node, 12)
}

func kickOffDraft(championSet map[byte]c.Champion, chosenChampionId byte) (float32, *c.ScoredTrieNode, int) {
	t1SelectableChampions := c.CreateTeamSelectableChampions(championSet)
	t2SelectableChampions := c.CreateTeamSelectableChampions(championSet)
	numT1Picks, numT2Picks := 0, 0

	deleteChampionIdFromSelectableChampionsInPlace(
		chosenChampionId,
		0,
		&numT1Picks,
		&numT2Picks,
		t1SelectableChampions,
		t2SelectableChampions,
	)

	return process(
		[]byte{},
		chosenChampionId,
		0,
		numT1Picks,
		numT2Picks,
		false,
		false,
		t1SelectableChampions,
		t2SelectableChampions,
	)
}

var TestDraft = []byte{7, 8, 0, 0, 2, 2, 1, 1, 4, 4}

func process(
	previousState []byte,
	chosenChampionId byte,
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1HasSupport bool,
	t2HasSupport bool,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) (float32, *c.ScoredTrieNode, int) {
	fmt.Println("process()", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name)
	currentState := append(previousState, chosenChampionId)

	// Base case
	if draftStepIdx >= len(c.DraftOrder)-1 {
		evaluation := float32(evaluateCompletedState(currentState))
		leafNode := c.ScoredTrieNode{
			ChampionName:      IdToChampion[chosenChampionId].Name,
			AverageEvaluation: evaluation,
		}
		return evaluation, &leafNode, 1
	}

	node := c.ScoredTrieNode{
		ChampionName:      IdToChampion[chosenChampionId].Name,
		AverageEvaluation: 0,
		Children:          make(map[byte]*c.ScoredTrieNode),
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
	fmt.Println(c.DraftOrder[draftStepIdx+1], selectableChampions)

	tempNumChampionsRan := 0
	numCompletedStates := 0
	evaluationSum := float32(0)
	for _, championId := range selectableChampions {
		if (draftStepIdx+1 < len(TestDraft)) && championId != TestDraft[draftStepIdx+1] {
			continue
		}

		tempNumChampionsRan += 1

		copyOfNumT1Picks := numT1Picks
		copyOfNumT2Picks := numT2Picks

		var affectedMaps []*map[byte]bool
		if draftStepIdx < len(c.DraftOrder)-1 {
			affectedMaps = deleteChampionIdFromSelectableChampionsInPlace(
				championId,
				draftStepIdx+1,
				&copyOfNumT1Picks,
				&copyOfNumT2Picks,
				t1SelectableChampions,
				t2SelectableChampions,
			)
		}

		evaluation, childNode, asdf := process(
			currentState,
			championId,
			draftStepIdx+1,
			copyOfNumT1Picks,
			copyOfNumT2Picks,
			t1HasSupport || (c.DraftOrder[draftStepIdx+1] == "T1P" && IdToChampion[championId].Role == c.SupportRole),
			t2HasSupport || (c.DraftOrder[draftStepIdx+1] == "T2P" && IdToChampion[championId].Role == c.SupportRole),
			t1SelectableChampions,
			t2SelectableChampions,
		)

		evaluationSum += evaluation
		node.Children[championId] = childNode
		numCompletedStates += asdf

		addChampionIdToSelectableChampionsInPlace(championId, affectedMaps)

		if draftStepIdx+1 == 2 {
			//fmt.Println(c.DraftOrder[draftStepIdx+1], IdToChampion[championId].Name)
			//fmt.Println("inside for loop", c.DraftOrder[draftStepIdx+1], IdToChampion[championId].Name, "evaluation:", evaluation, "evaluationSum", evaluationSum)
		}
		if draftStepIdx+1 < 10 {
			//fmt.Println("evaluation:", evaluation)
			//c.PrintMemUsage()
			break
		}
	}

	//if draftStepIdx+1 == 3 {
	//	fmt.Println(c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan, "evaluation", evaluationSum/float32(tempNumChampionsRan))
	//}
	//if draftStepIdx+1 == 2 {
	//	fmt.Println(IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan)
	//}
	//if draftStepIdx+1 == 1 {
	//	fmt.Println(IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan)
	//}
	// TODO: Remove when using longer champ list.
	if len(selectableChampions) != 0 {
		node.AverageEvaluation = evaluationSum / float32(tempNumChampionsRan)
	}
	//node.AverageEvaluation = evaluationSum / float32(len(selectableChampions))
	//if draftStepIdx < 3 {
	//	fmt.Printf("    -> %s %s: %f\n", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, node.AverageEvaluation)
	//}
	//if draftStepIdx < 2 {
	//	fmt.Printf("  -> %s %s: %f\n", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, node.AverageEvaluation)
	//}
	return node.AverageEvaluation, &node, numCompletedStates
}

// From T1's perspective.
func evaluateCompletedState(completedState []byte) int {
	evaluation := 0
	t1Picks := []byte{
		completedState[c.T1PIdxs[0]],
		completedState[c.T1PIdxs[1]],
		completedState[c.T1PIdxs[2]],
	}
	t2Picks := []byte{
		completedState[c.T2PIdxs[0]],
		completedState[c.T2PIdxs[1]],
		completedState[c.T2PIdxs[2]],
	}

	for i, t1ChampionId := range t1Picks {
		for _, t2ChampionId := range t2Picks {
			matchup := ChampionMatchups[t1ChampionId][t2ChampionId]
			evaluation += matchup
		}

		for _, t1ChampionId2 := range t1Picks[i+1:] {
			synergy := ChampionSynergys[t1ChampionId][t1ChampionId2]
			evaluation += synergy
		}
	}

	for i, t2ChampionId := range t2Picks {
		for _, t2ChampionId2 := range t2Picks[i+1:] {
			synergy := ChampionSynergys[t2ChampionId][t2ChampionId2]
			evaluation += (-1) * synergy
		}
	}

	//fmt.Println(completedState, "evaluation:", evaluation)
	return evaluation
}

func getSelectableChampions(
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1HasSupport bool,
	t2HasSupport bool,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) []byte {
	var selectableChampions map[byte]bool
	t1NeedsSupportThisStep := !t1HasSupport && numT1Picks >= 2
	t2NeedsSupportThisStep := !t2HasSupport && numT2Picks >= 2

	switch c.DraftOrder[draftStepIdx] {
	case "T1P":
		if t1NeedsSupportThisStep {
			selectableChampions = t1SelectableChampions.PickableSupportChampions
			break
		}
		selectableChampions = t1SelectableChampions.PickableChampions
	case "T1GB", "T1B":
		if t2NeedsSupportThisStep {
			selectableChampions = t1SelectableChampions.BannableSupportChampions
			break
		}
		selectableChampions = t1SelectableChampions.BannableChampions
	case "T2P":
		if t2NeedsSupportThisStep {
			selectableChampions = t2SelectableChampions.PickableSupportChampions
			break
		}
		selectableChampions = t2SelectableChampions.PickableChampions
	case "T2GB", "T2B":
		if t1NeedsSupportThisStep {
			selectableChampions = t2SelectableChampions.BannableSupportChampions
			break
		}
		selectableChampions = t2SelectableChampions.BannableChampions
	default:
		log.Fatal("getSelectableChampions(): Invalid draft step.")
	}

	selectableChampionsSlice := make([]byte, 0, len(selectableChampions))
	for championId := range selectableChampions {
		selectableChampionsSlice = append(selectableChampionsSlice, championId)
	}
	return selectableChampionsSlice
}

func deleteChampionIdFromSelectableChampionsInPlace(
	championId byte,
	currDraftStepIdx int,
	numT1Picks *int,
	numT2Picks *int,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) []*map[byte]bool {
	switch c.DraftOrder[currDraftStepIdx] {
	case "T1GB", "T2GB":
		return globalBan(championId, t1SelectableChampions, t2SelectableChampions)
	case "T1P":
		*numT1Picks += 1
		return t1PickT2Ban(championId, t1SelectableChampions, t2SelectableChampions)
	case "T1B":
		return t1BanT2Pick(championId, t1SelectableChampions, t2SelectableChampions)
	case "T2P":
		*numT2Picks += 1
		return t1BanT2Pick(championId, t1SelectableChampions, t2SelectableChampions)
	case "T2B":
		return t1PickT2Ban(championId, t1SelectableChampions, t2SelectableChampions)
	default:
		log.Fatal("updateSelectableChampionsInPlace(): Invalid draft step.")
		return []*map[byte]bool{}
	}
}

func globalBan(
	championId byte,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) []*map[byte]bool {
	var affectedMaps []*map[byte]bool

	delete(t1SelectableChampions.PickableChampions, championId)
	delete(t1SelectableChampions.BannableChampions, championId)
	delete(t2SelectableChampions.PickableChampions, championId)
	delete(t2SelectableChampions.BannableChampions, championId)

	delete(t1SelectableChampions.PickableSupportChampions, championId)
	delete(t1SelectableChampions.BannableSupportChampions, championId)
	delete(t2SelectableChampions.PickableSupportChampions, championId)
	delete(t2SelectableChampions.BannableSupportChampions, championId)

	affectedMaps = append(affectedMaps, &t1SelectableChampions.PickableChampions)
	affectedMaps = append(affectedMaps, &t1SelectableChampions.BannableChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.PickableChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.BannableChampions)

	affectedMaps = append(affectedMaps, &t1SelectableChampions.PickableSupportChampions)
	affectedMaps = append(affectedMaps, &t1SelectableChampions.BannableSupportChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.PickableSupportChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.BannableSupportChampions)

	return affectedMaps
}

func t1BanT2Pick(
	championId byte,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) []*map[byte]bool {
	var affectedMaps []*map[byte]bool

	delete(t1SelectableChampions.BannableChampions, championId)
	delete(t2SelectableChampions.PickableChampions, championId)

	delete(t1SelectableChampions.BannableSupportChampions, championId)
	delete(t2SelectableChampions.PickableSupportChampions, championId)

	affectedMaps = append(affectedMaps, &t1SelectableChampions.BannableChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.PickableChampions)

	affectedMaps = append(affectedMaps, &t1SelectableChampions.BannableSupportChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.PickableSupportChampions)

	return affectedMaps
}

func t1PickT2Ban(
	championId byte,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) []*map[byte]bool {
	var affectedMaps []*map[byte]bool

	delete(t1SelectableChampions.PickableChampions, championId)
	delete(t2SelectableChampions.BannableChampions, championId)

	delete(t1SelectableChampions.PickableSupportChampions, championId)
	delete(t2SelectableChampions.BannableSupportChampions, championId)

	affectedMaps = append(affectedMaps, &t1SelectableChampions.PickableChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.BannableChampions)

	affectedMaps = append(affectedMaps, &t1SelectableChampions.PickableSupportChampions)
	affectedMaps = append(affectedMaps, &t2SelectableChampions.BannableSupportChampions)

	return affectedMaps
}

func addChampionIdToSelectableChampionsInPlace(
	championId byte,
	affectedMaps []*map[byte]bool,
) {
	for _, affectedMap := range affectedMaps {
		(*affectedMap)[championId] = true
	}
}
