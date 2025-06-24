package prep

import (
	"battlerite-draft-helper/c"
	"battlerite-draft-helper/data"
)

const DataDir = "data/"

func InitializeGlobalVariables(champions []c.Champion) (
	int,
	map[string]byte,
	map[byte]c.Champion,
	[]int8,
	c.Team,
	c.Team,
) {
	championNameToId := make(map[string]byte)
	idToChampion := make(map[byte]c.Champion)
	championMatchups := make(map[byte]map[byte]int8)
	championSynergys := make(map[byte]map[byte]int8)

	for _, champion := range champions {
		championNameToId[champion.Name] = champion.Id
		idToChampion[champion.Id] = champion
	}

	championMatchups = data.FormatCsvData(championNameToId, DataDir+data.MatchupsCsvFilename, false)
	championSynergys = data.FormatCsvData(championNameToId, DataDir+data.SynergiesCsvFilename, true)

	player1, player2, player3 := data.GetPlayerChampions(championNameToId, "deniz", "vet", "bo4")
	player4, player5, player6 := data.GetPlayerChampions(championNameToId, "lepnix", "peasprout", "faris")
	team1 := populateTeamPickPoolsUsingPlayerChampionPools(champions, championSynergys, player1, player2, player3)
	team2 := populateTeamPickPoolsUsingPlayerChampionPools(champions, championSynergys, player4, player5, player6)

	flatChampionMatchups := createFlatChampionMatchups(champions, championMatchups)

	return len(champions), championNameToId, idToChampion, flatChampionMatchups, team1, team2
}

func populateTeamPickPoolsUsingPlayerChampionPools(
	champions []c.Champion,
	championSynergys map[byte]map[byte]int8,
	player1 c.Player,
	player2 c.Player,
	player3 c.Player,
) c.Team {
	numChampions := len(champions)
	team := c.Team{
		Pick1Pool: make([]int8, numChampions),
		Pick2Pool: make([]int8, numChampions*numChampions),
		Pick3Pool: make([]int8, numChampions*numChampions*numChampions),
	}

	fillSliceWithNilValues(team.Pick1Pool)
	fillSliceWithNilValues(team.Pick2Pool)
	fillSliceWithNilValues(team.Pick3Pool)

	for _, player := range []c.Player{player1, player2, player3} {
		populatePick1Pool(&team, player)
	}

	for i, p1 := range []c.Player{player1, player2, player3} {
		for j, p2 := range []c.Player{player1, player2, player3} {
			if i == j {
				continue
			}
			populatePick2Pool(numChampions, &team, p1.ChampionPool, p2.ChampionPool)
		}
	}

	for i, p1 := range []c.Player{player1, player2, player3} {
		for j, p2 := range []c.Player{player1, player2, player3} {
			for k, p3 := range []c.Player{player1, player2, player3} {
				if i == j || i == k || j == k {
					continue
				}
				populatePick3Pool(numChampions, &team, p1.ChampionPool, p2.ChampionPool, p3.ChampionPool)
			}
		}
	}
	addChampionSynergiesToPick3Pool(champions, championSynergys, &team)

	return team
}

func fillSliceWithNilValues(slice []int8) {
	for i := range slice {
		slice[i] = -128
	}
}

func populatePick1Pool(team *c.Team, player c.Player) {
	for championId, evaluation := range player.ChampionPool {
		if evaluation > team.Pick1Pool[championId] {
			team.Pick1Pool[championId] = evaluation
		}
	}
}

func populatePick2Pool(numChampions int, team *c.Team, championPool1, championPool2 map[byte]int8) {
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

func populatePick3Pool(
	numChampions int,
	team *c.Team,
	championPool1, championPool2, championPool3 map[byte]int8,
) {
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

func addChampionSynergiesToPick3Pool(champions []c.Champion, championSynergys map[byte]map[byte]int8, team *c.Team) {
	numChampions := len(champions)

	for i, champion1 := range champions {
		for j, champion2 := range champions {
			for k, champion3 := range champions {
				if i == j || i == k || j == k {
					continue
				}

				index := int(champion1.Id) + int(champion2.Id)*numChampions + int(champion3.Id)*numChampions*numChampions
				if team.Pick3Pool[index] == -128 {
					continue
				}

				team.Pick3Pool[index] += evaluateTeamChampionSynergy(championSynergys, champion1.Id, champion2.Id, champion3.Id)
			}
		}
	}
}

func evaluateTeamChampionSynergy(
	championSynergys map[byte]map[byte]int8,
	champion1Id, champion2Id, champion3Id byte,
) int8 {
	var evaluation int8 = 0

	evaluation += championSynergys[champion1Id][champion2Id]
	evaluation += championSynergys[champion1Id][champion3Id]
	evaluation += championSynergys[champion2Id][champion3Id]

	return evaluation
}

func createFlatChampionMatchups(champions []c.Champion, championMatchups map[byte]map[byte]int8) []int8 {
	numChampions := len(champions)
	flatChampionMatchups := make([]int8, numChampions*numChampions*numChampions*numChampions)

	for _, champion1 := range champions {
		for _, champion2 := range champions {
			for _, champion3 := range champions {
				for _, champion4 := range champions {
					index := int(champion1.Id) + int(champion2.Id)*numChampions + int(champion3.Id)*numChampions*numChampions + int(champion4.Id)*numChampions*numChampions*numChampions
					flatChampionMatchups[index] = championMatchups[champion1.Id][champion2.Id] + championMatchups[champion1.Id][champion3.Id] + championMatchups[champion1.Id][champion4.Id]
				}
			}
		}
	}

	return flatChampionMatchups
}

func CreateTeamSelectableChampions(champions map[byte]c.Champion, t1PickPool, t2PickPool []int8) (
	c.TeamSelectableChampions,
	c.TeamSelectableChampions,
) {
	t1SelectableChampions := TeamSelectableChampionsConstructor()
	t2SelectableChampions := TeamSelectableChampionsConstructor()

	populateTeamSelectableChampions(champions, t1PickPool, t1SelectableChampions, t2SelectableChampions)
	populateTeamSelectableChampions(champions, t2PickPool, t2SelectableChampions, t1SelectableChampions)

	return t1SelectableChampions, t2SelectableChampions
}

func populateTeamSelectableChampions(
	champions map[byte]c.Champion,
	pickPool []int8,
	t1SelectableChampions, t2SelectableChampions c.TeamSelectableChampions,
) {
	for championId, evaluation := range pickPool {
		if evaluation == -128 {
			continue
		}

		t1SelectableChampions.PickableChampions[byte(championId)] = true
		t2SelectableChampions.BannableChampions[byte(championId)] = true

		if champions[byte(championId)].Role == c.SupportRole {
			t1SelectableChampions.PickableSupportChampions[byte(championId)] = true
			t2SelectableChampions.BannableSupportChampions[byte(championId)] = true
		}
	}
}

func TeamSelectableChampionsConstructor() c.TeamSelectableChampions {
	return c.TeamSelectableChampions{
		PickableChampions:        make(map[byte]bool),
		BannableChampions:        make(map[byte]bool),
		PickableSupportChampions: make(map[byte]bool),
		BannableSupportChampions: make(map[byte]bool),
	}
}
