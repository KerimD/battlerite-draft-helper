package main

import (
	"battlerite-draft-helper/c"
	"battlerite-draft-helper/data"
	"fmt"
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

	vroomVroom(championSet)
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

	evaluationSum := float32(0)
	for championId, champion := range championSet {
		evaluation, childNode := kickOffDraft(championSet, championId)
		fmt.Println(champion.Name, "->", evaluation)

		evaluationSum += evaluation
		node.Children[championId] = childNode

		//c.FormatCompletedStates(node.CompletedStates[533116:533119], IdToChampion)
		//c.FormatCompletedStates(node.CompletedStates, IdToChampion)
		//fmt.Println(node.CompletedStates)
		//break
	}

	node.AverageEvaluation = evaluationSum / float32(len(championSet))
}

func kickOffDraft(championSet map[byte]c.Champion, chosenChampionId byte) (float32, *c.ScoredTrieNode) {
	t1SelectableChampions := c.CreateTeamSelectableChampions(championSet)
	t2SelectableChampions := c.CreateTeamSelectableChampions(championSet)

	numT1Picks, numT2Picks := updateSelectableChampionsInPlace(
		chosenChampionId,
		0,
		0,
		0,
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
) (float32, *c.ScoredTrieNode) {
	currentState := append(previousState, chosenChampionId)

	// Base case
	if draftStepIdx >= len(c.DraftOrder)-1 {
		evaluation := float32(evaluateCompletedState(currentState))
		leafNode := c.ScoredTrieNode{
			ChampionName:      IdToChampion[chosenChampionId].Name,
			AverageEvaluation: evaluation,
		}
		return evaluation, &leafNode
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

	evaluationSum := float32(0)
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

		evaluation, childNode := process(
			currentState,
			championId,
			draftStepIdx+1,
			numT1Picks,
			numT2Picks,
			t1HasSupport || (c.DraftOrder[draftStepIdx+1] == "T1P" && IdToChampion[championId].Role == c.SupportRole),
			t2HasSupport || (c.DraftOrder[draftStepIdx+1] == "T2P" && IdToChampion[championId].Role == c.SupportRole),
			deepCopyT1SelectableChampions,
			deepCopyT2SelectableChampions,
		)
		//if math.IsNaN(float64(evaluation)) {
		//	fmt.Println("evaluation is NaN")
		//}
		evaluationSum += evaluation
		node.Children[championId] = childNode

		if draftStepIdx < 3 {
			//fmt.Println("evaluation:", evaluation)
			//c.PrintMemUsage()
			//break
		}
	}

	if len(selectableChampions) != 0 { // TODO: Remove when using longer champ list.
		node.AverageEvaluation = evaluationSum / float32(len(selectableChampions))
	}
	//if draftStepIdx < 3 {
	//	fmt.Printf("    -> %s %s: %f\n", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, node.AverageEvaluation)
	//}
	if draftStepIdx < 2 {
		fmt.Printf("  -> %s %s: %f\n", c.DraftOrder[draftStepIdx], IdToChampion[chosenChampionId].Name, node.AverageEvaluation)
	}
	return node.AverageEvaluation, &node
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
) map[byte]bool {
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

func updateSelectableChampionsInPlace(
	championId byte,
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
