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
	map[byte]map[byte]int,
	map[byte]map[byte]int,
	c.Team,
	c.Team,
) {
	numChampions := len(champions)
	championNameToId := make(map[string]byte)
	idToChampion := make(map[byte]c.Champion)
	championMatchups := make(map[byte]map[byte]int)
	championSynergys := make(map[byte]map[byte]int)

	for _, champion := range champions {
		championNameToId[champion.Name] = champion.Id
		idToChampion[champion.Id] = champion
	}

	championMatchups = data.FormatCsvData(championNameToId, DataDir+data.MatchupsCsvFilename, false)
	championSynergys = data.FormatCsvData(championNameToId, DataDir+data.SynergiesCsvFilename, true)

	player1, player2, player3 := data.GetPlayerChampions(championNameToId, "deniz", "vet", "bo4")
	team1 := populateTeamPickPoolsUsingPlayerChampionPools(numChampions, player1, player2, player3)
	team2 := populateTeamPickPoolsUsingPlayerChampionPools(numChampions, player1, player2, player3) // TODO: 3 new players.

	return numChampions, championNameToId, idToChampion, championMatchups, championSynergys, team1, team2
}

func populateTeamPickPoolsUsingPlayerChampionPools(
	numChampions int,
	player1 c.Player,
	player2 c.Player,
	player3 c.Player,
) c.Team {
	team := c.Team{
		Pick1Pool: make([]int8, numChampions),
		Pick2Pool: make([]int8, numChampions*numChampions),
		Pick3Pool: make([]int8, numChampions*numChampions*numChampions),
	}

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

	return team
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

func populatePick3Pool(numChampions int, team *c.Team, championPool1, championPool2, championPool3 map[byte]int8) {
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
