package main

import (
	"battlerite-draft-helper/c"
	"battlerite-draft-helper/data"
	"battlerite-draft-helper/prep"
	"fmt"
	"log"
	"time"
)

const DataDir = "data/"

var (
	NumChampions         int
	ChampionNameToId     = make(map[string]byte)
	IdToChampion         = make(map[byte]c.Champion)
	FlatChampionMatchups []int8
	T1                   c.Team
	T2                   c.Team
)

// Note: Currently does not support global bans after picks.
func main() {
	//champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsCsvFilename)
	NumChampions, ChampionNameToId, IdToChampion, FlatChampionMatchups, T1, T2 = prep.InitializeGlobalVariables(champions)

	championSet := make(map[byte]c.Champion, len(champions))
	for _, champion := range champions {
		championSet[champion.Id] = champion
	}

	start := time.Now()
	vroomVroom(championSet)
	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func vroomVroom(championSet map[byte]c.Champion) {
	t1SelectableChampions, t2SelectableChampions := prep.CreateTeamSelectableChampions(
		championSet,
		T1.Pick1Pool,
		T2.Pick1Pool,
	)

	rootNode, numCompletedStates := process(
		[]byte{},
		0,
		0,
		0,
		false,
		false,
		t1SelectableChampions,
		t2SelectableChampions,
	)

	fmt.Println("numCompletedStates:", numCompletedStates)
	c.PrintTree(IdToChampion, rootNode, len(TestDraft))
}

// var TestDraft = []byte{1, 4, 24, 16, 3}
// /////////////////// b, b, b, b, p, p, p, p, b, b
var TestDraft = []byte{1, 4, 24, 16}

func process(
	currentState []byte,
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	t1HasSupport bool,
	t2HasSupport bool,
	t1SelectableChampions c.TeamSelectableChampions,
	t2SelectableChampions c.TeamSelectableChampions,
) (*c.ScoredTrieNode, int) {

	// Base case
	if draftStepIdx >= len(c.DraftOrder) {
		evaluation := float32(evaluateCompletedState(currentState))
		leafNode := c.ScoredTrieNode{
			AverageEvaluation: evaluation,
		}
		return &leafNode, 1
	}

	start := time.Now()
	//fmt.Println("process()", c.DraftOrder[draftStepIdx], currentState)

	node := c.ScoredTrieNode{
		Children: make(map[byte]*c.ScoredTrieNode),
	}

	selectableChampions := getSelectableChampions(
		draftStepIdx,
		numT1Picks,
		numT2Picks,
		t1HasSupport,
		t2HasSupport,
		t1SelectableChampions,
		t2SelectableChampions,
	)

	sumNumCompletedStates := 0
	numChampionsRanWithCompletedStates := 0
	evaluationSum := float32(0)
	for _, championId := range selectableChampions {
		// TODO: Remove.
		if (draftStepIdx < len(TestDraft)) && championId != TestDraft[draftStepIdx] {
			continue
		}

		if !wouldTeamRealisticallyMakeThisSelection(currentState, draftStepIdx, numT1Picks, numT2Picks, championId) {
			continue
		}

		copyOfNumT1Picks := numT1Picks
		copyOfNumT2Picks := numT2Picks

		var affectedMaps []*map[byte]bool
		if draftStepIdx < len(c.DraftOrder)-1 {
			affectedMaps = deleteChampionIdFromSelectableChampionsInPlace(
				championId,
				draftStepIdx,
				&copyOfNumT1Picks,
				&copyOfNumT2Picks,
				t1SelectableChampions,
				t2SelectableChampions,
			)
		}

		futureState := append(currentState, championId)
		childNode, numCompletedStates := process(
			futureState,
			draftStepIdx+1,
			copyOfNumT1Picks,
			copyOfNumT2Picks,
			t1HasSupport || (c.DraftOrder[draftStepIdx] == "T1P" && IdToChampion[championId].Role == c.SupportRole),
			t2HasSupport || (c.DraftOrder[draftStepIdx] == "T2P" && IdToChampion[championId].Role == c.SupportRole),
			t1SelectableChampions,
			t2SelectableChampions,
		)

		if numCompletedStates != 0 {
			numChampionsRanWithCompletedStates += 1
			evaluationSum += childNode.AverageEvaluation
			node.Children[championId] = childNode
			sumNumCompletedStates += numCompletedStates
		}

		addChampionIdToSelectableChampionsInPlace(championId, affectedMaps)

		//if draftStepIdx+1 == 7 {
		//	fmt.Println("inside for loop", c.DraftOrder[draftStepIdx+1], IdToChampion[championId].Name, "childNode.AverageEvaluation:", childNode.AverageEvaluation, "evaluationSum", evaluationSum)
		//}
		//if draftStepIdx+1 < 3 {
		//fmt.Println("evaluation:", evaluation)
		//c.PrintMemUsage()
		//break
		//}
	}

	if sumNumCompletedStates != 0 {
		node.AverageEvaluation = evaluationSum / float32(numChampionsRanWithCompletedStates)
	}

	//if draftStepIdx == 7 {
	//	fmt.Println(c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan, "evaluation", evaluationSum/float32(tempNumChampionsRan))
	//}
	//if draftStepIdx == 7 {
	//	fmt.Printf("%s2 %s, Time: %v", c.DraftOrder[draftStepIdx-1], IdToChampion[currentState[len(currentState)-1]].Name, time.Since(start))
	//	fmt.Println(", eval:", node.AverageEvaluation, "num states:", sumNumCompletedStates)
	//}
	//if draftStepIdx == 6 {
	//	fmt.Printf("%s %s, Time: %v", c.DraftOrder[draftStepIdx-1], IdToChampion[currentState[len(currentState)-1]].Name, time.Since(start))
	//	fmt.Println(", eval:", node.AverageEvaluation, "num states:", sumNumCompletedStates)
	//}
	if draftStepIdx == 5 {
		fmt.Printf("%s %s, Time: %v", c.DraftOrder[draftStepIdx-1], IdToChampion[currentState[len(currentState)-1]].Name, time.Since(start))
		fmt.Println(", eval:", node.AverageEvaluation, "num states:", sumNumCompletedStates)
	}
	//if draftStepIdx == 2 {
	//	fmt.Println(IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan)
	//}
	//if draftStepIdx == 1 {
	//	fmt.Println(IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan)
	//}

	return &node, sumNumCompletedStates
}

// From T1's perspective.
func evaluateCompletedState(completedState []byte) int8 {
	var evaluation int8 = 0
	t1ChampionIds := []byte{
		completedState[c.T1PIdxs[0]],
		completedState[c.T1PIdxs[1]],
		completedState[c.T1PIdxs[2]],
	}
	t2ChampionIds := []byte{
		completedState[c.T2PIdxs[0]],
		completedState[c.T2PIdxs[1]],
		completedState[c.T2PIdxs[2]],
	}

	t2Index := int(t2ChampionIds[0])*NumChampions + int(t2ChampionIds[1])*NumChampions*NumChampions + int(t2ChampionIds[2])*NumChampions*NumChampions*NumChampions
	evaluation += FlatChampionMatchups[int(t1ChampionIds[0])+t2Index]
	evaluation += FlatChampionMatchups[int(t1ChampionIds[1])+t2Index]
	evaluation += FlatChampionMatchups[int(t1ChampionIds[2])+t2Index]

	evaluation += T1.Pick3Pool[int(t1ChampionIds[0])+int(t1ChampionIds[1])*NumChampions+int(t1ChampionIds[2])*NumChampions*NumChampions]
	evaluation -= T2.Pick3Pool[int(t2ChampionIds[0])+int(t2ChampionIds[1])*NumChampions+int(t2ChampionIds[2])*NumChampions*NumChampions]

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

func wouldTeamRealisticallyMakeThisSelection(
	currentState []byte,
	draftStepIdx int,
	numT1Picks int,
	numT2Picks int,
	championId byte,
) bool {
	switch c.DraftOrder[draftStepIdx] {
	case "T1P", "T2B":
		if numT1Picks == 1 {
			index := int(currentState[c.T1PIdxs[0]]) + int(championId)*NumChampions
			return T1.Pick2Pool[index] != -128
		}
		if numT1Picks == 2 {
			index := int(currentState[c.T1PIdxs[0]]) + int(currentState[c.T1PIdxs[1]])*NumChampions + int(championId)*NumChampions*NumChampions
			return T1.Pick3Pool[index] != -128
		}
	case "T2P", "T1B":
		if numT2Picks == 1 {
			index := int(currentState[c.T2PIdxs[0]]) + int(championId)*NumChampions
			return T2.Pick2Pool[index] != -128
		}
		if numT2Picks == 2 {
			index := int(currentState[c.T2PIdxs[0]]) + int(currentState[c.T2PIdxs[1]])*NumChampions + int(championId)*NumChampions*NumChampions
			return T2.Pick3Pool[index] != -128
		}
	}

	return true
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
