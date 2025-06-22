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
	numChampions     int
	ChampionNameToId = make(map[string]byte)
	IdToChampion     = make(map[byte]c.Champion)
	ChampionMatchups = make(map[byte]map[byte]int)
	ChampionSynergys = make(map[byte]map[byte]int)
	t1               c.Team
	t2               c.Team
)

// Note: Currently does not support global bans after picks.
func main() {
	//champions := data.GetChampionsFromCsv(DataDir + data.ChampionsShortListCsvFilename)
	champions := data.GetChampionsFromCsv(DataDir + data.ChampionsCsvFilename)
	initializeGlobalVariables(champions)

	//player1, player2, player3 := data.GetPlayerChampions(ChampionNameToId, "deniz", "vet", "bo4")
	//populateTeamPickPoolsUsingPlayerChampionPools(&t1, player1, player2, player3)
	//populateTeamPickPoolsUsingPlayerChampionPools(&t2, player1, player2, player3) // TODO: 3 new players.

	championSet := make(map[byte]c.Champion, len(champions))
	for _, champion := range champions {
		championSet[champion.Id] = champion
	}

	start := time.Now()
	vroomVroom(championSet)
	defer fmt.Printf("Program completed in %v", time.Since(start))
}

func initializeGlobalVariables(champions []c.Champion) {
	numChampions = len(champions)
	for _, champion := range champions {
		ChampionNameToId[champion.Name] = champion.Id
		IdToChampion[champion.Id] = champion
	}
	ChampionMatchups = data.FormatCsvData(ChampionNameToId, DataDir+data.MatchupsCsvFilename, false)
	ChampionSynergys = data.FormatCsvData(ChampionNameToId, DataDir+data.SynergiesCsvFilename, true)
}

func populateTeamPickPoolsUsingPlayerChampionPools(
	team *c.Team,
	player1 c.Player,
	player2 c.Player,
	player3 c.Player,
) {
	team.Pick1Pool = make([]int8, numChampions)
	team.Pick2Pool = make([]int8, numChampions*numChampions)
	team.Pick3Pool = make([]int8, numChampions*numChampions*numChampions)

	for _, player := range []c.Player{player1, player2, player3} {
		populatePick1Pool(team, player)
	}

	for i, p1 := range []c.Player{player1, player2, player3} {
		for j, p2 := range []c.Player{player1, player2, player3} {
			if i == j {
				continue
			}
			populatePick2Pool(team, p1.ChampionPool, p2.ChampionPool)
		}
	}

	for i, p1 := range []c.Player{player1, player2, player3} {
		for j, p2 := range []c.Player{player1, player2, player3} {
			for k, p3 := range []c.Player{player1, player2, player3} {
				if i == j || i == k || j == k {
					continue
				}
				populatePick3Pool(team, p1.ChampionPool, p2.ChampionPool, p3.ChampionPool)
			}
		}
	}
}

func populatePick1Pool(team *c.Team, player c.Player) {
	for championId, evaluation := range player.ChampionPool {
		if evaluation > team.Pick1Pool[championId] {
			team.Pick1Pool[championId] = evaluation
		}
	}
}

func populatePick2Pool(team *c.Team, championPool1, championPool2 map[byte]int8) {
	for champion1Id, evaluation1 := range championPool1 {
		for champion2Id, evaluation2 := range championPool2 {
			if champion1Id == champion2Id {
				continue
			}

			evaluationSum := evaluation1 + evaluation2
			index := int(champion1Id) + int(champion2Id)*numChampions
			if evaluationSum > team.Pick2Pool[index] {
				team.Pick2Pool[index] = evaluationSum
			}
		}
	}
}

func populatePick3Pool(team *c.Team, championPool1, championPool2, championPool3 map[byte]int8) {
	for champion1Id, evaluation1 := range championPool1 {
		for champion2Id, evaluation2 := range championPool2 {
			for champion3Id, evaluation3 := range championPool3 {
				if champion1Id == champion2Id || champion1Id == champion3Id || champion2Id == champion3Id {
					continue
				}

				evaluationSum := evaluation1 + evaluation2 + evaluation3
				index := int(champion1Id) + int(champion2Id)*numChampions + int(champion3Id)*numChampions*numChampions
				if evaluationSum > team.Pick3Pool[index] {
					team.Pick3Pool[index] = evaluationSum
				}
			}
		}
	}
}

func vroomVroom(championSet map[byte]c.Champion) {
	node := c.ScoredTrieNode{
		AverageEvaluation: 0,
		Children:          make(map[byte]*c.ScoredTrieNode),
	}

	sumNumCompletedStates := 0
	tempNumChampionsRan := 0
	evaluationSum := float32(0)
	for championId := range championSet {
		if championId != TestDraft[0] {
			continue
		}

		tempNumChampionsRan += 1

		childNode, numCompletedStates := kickOffDraft(championSet, championId)

		evaluationSum += childNode.AverageEvaluation
		node.Children[championId] = childNode
		sumNumCompletedStates += numCompletedStates
		break
	}

	//node.AverageEvaluation = evaluationSum / float32(len(championSet))
	node.AverageEvaluation = evaluationSum / float32(tempNumChampionsRan)
	fmt.Println("numCompletedStates:", sumNumCompletedStates)
	c.PrintTree(IdToChampion, &node, 6)
}

func kickOffDraft(championSet map[byte]c.Champion, chosenChampionId byte) (*c.ScoredTrieNode, int) {
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

// var TestDraft = []byte{7, 8, 0, 0, 2, 2, 1, 1, 4, 4}
// /////////////////// b, b, b, b, p, p, p, p, b, b
var TestDraft = []byte{3, 4, 24, 11, 1, 1}

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
) (*c.ScoredTrieNode, int) {
	start := time.Now()
	//fmt.Println("process()", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name)
	currentState := append(previousState, chosenChampionId)

	// Base case
	if draftStepIdx >= len(c.DraftOrder)-1 {
		evaluation := float32(evaluateCompletedState(currentState))
		leafNode := c.ScoredTrieNode{
			AverageEvaluation: evaluation,
		}
		return &leafNode, 1
	}

	node := c.ScoredTrieNode{
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
	//fmt.Println(c.DraftOrder[draftStepIdx+1], selectableChampions)

	sumNumCompletedStates := 0
	numChampionsRanWithCompletedStates := 0
	evaluationSum := float32(0)
	for _, championId := range selectableChampions {
		if (draftStepIdx+1 < len(TestDraft)) && championId != TestDraft[draftStepIdx+1] {
			continue
		} else {
			//fmt.Println("Forcing", c.DraftOrder[draftStepIdx+1], IdToChampion[championId].Name)
		}

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

		childNode, numCompletedStates := process(
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
		if draftStepIdx+1 < 3 {
			//fmt.Println("evaluation:", evaluation)
			//c.PrintMemUsage()
			break
		}
	}

	if sumNumCompletedStates != 0 {
		node.AverageEvaluation = evaluationSum / float32(numChampionsRanWithCompletedStates)
	}

	//if draftStepIdx+1 == 7 {
	//	fmt.Println(c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan, "evaluation", evaluationSum/float32(tempNumChampionsRan))
	//}
	if draftStepIdx+1 == 7 {
		fmt.Printf("%s2 %s, Time: %v", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, time.Since(start))
		fmt.Println(", eval:", node.AverageEvaluation, "num states:", sumNumCompletedStates)
	}
	if draftStepIdx+1 == 6 {
		fmt.Printf("%s %s, Time: %v", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, time.Since(start))
		fmt.Println(", eval:", node.AverageEvaluation, "num states:", sumNumCompletedStates)
	}
	//if draftStepIdx+1 == 2 {
	//	fmt.Println(IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan)
	//}
	//if draftStepIdx+1 == 1 {
	//	fmt.Println(IdToChampion[chosenChampionId].Name, "evaluationSum:", evaluationSum, "tempNumChampionsRan:", tempNumChampionsRan)
	//}

	return &node, sumNumCompletedStates
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
