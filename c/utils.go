package c

import (
	"fmt"
	"runtime"
	"slices"
)

func CopyMap(original map[byte]bool) map[byte]bool {
	newMap := make(map[byte]bool, len(original))
	for key, value := range original {
		newMap[key] = value
	}
	return newMap
}

func CreateTeamSelectableChampions(champions map[byte]Champion) TeamSelectableChampions {
	teamSelectableChampions := TeamSelectableChampionsConstructor()

	for championId, champion := range champions {
		teamSelectableChampions.PickableChampions[championId] = true
		teamSelectableChampions.BannableChampions[championId] = true
		if champion.Role == SupportRole {
			teamSelectableChampions.PickableSupportChampions[championId] = true
			teamSelectableChampions.BannableSupportChampions[championId] = true
		}
	}

	return teamSelectableChampions
}

func DeepCopyTeamSelectableChampions(teamSelectableChampions TeamSelectableChampions) TeamSelectableChampions {
	deepCopy := TeamSelectableChampionsConstructor()

	deepCopy.PickableChampions = CopyMap(teamSelectableChampions.PickableChampions)
	deepCopy.BannableChampions = CopyMap(teamSelectableChampions.BannableChampions)
	deepCopy.PickableSupportChampions = CopyMap(teamSelectableChampions.PickableSupportChampions)
	deepCopy.BannableSupportChampions = CopyMap(teamSelectableChampions.BannableSupportChampions)

	return deepCopy
}

func TeamSelectableChampionsConstructor() TeamSelectableChampions {
	return TeamSelectableChampions{
		PickableChampions:        make(map[byte]bool),
		BannableChampions:        make(map[byte]bool),
		PickableSupportChampions: make(map[byte]bool),
		BannableSupportChampions: make(map[byte]bool),
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc = %d KB", bToMb(m.Alloc))
	fmt.Printf(", TotalAlloc = %d KB", bToMb(m.TotalAlloc))
	fmt.Printf(", Sys = %d KB", bToMb(m.Sys))
	fmt.Printf(", NumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

func kbToMb(kb uint64) uint64 {
	return kb / 1024
}

func PrintTree(node *ScoredTrieNode, maxDepth int) {
	printTree(node, 0, maxDepth)
}

func printTree(node *ScoredTrieNode, currentDepth int, maxDepth int) {
	if node == nil || currentDepth > maxDepth {
		return
	}

	indent := ""
	for i := 0; i < currentDepth; i++ {
		indent += "|  "
	}

	if node.ChampionName != "" {
		fmt.Printf("%s%s %s: %.4f\n", indent, DraftOrder[currentDepth-1], node.ChampionName, node.AverageEvaluation)
	}

	sortedList := mapToSortedSlice(node.Children)

	for _, childNode := range sortedList {
		printTree(&childNode, currentDepth+1, maxDepth)
	}
}

func mapToSortedSlice(nodesMap map[byte]*ScoredTrieNode) []ScoredTrieNode {
	nodes := make([]ScoredTrieNode, 0, len(nodesMap))

	for _, node := range nodesMap {
		nodes = append(nodes, *node)
	}

	slices.SortFunc(nodes, func(a, b ScoredTrieNode) int {
		if a.AverageEvaluation < b.AverageEvaluation {
			return 1
		} else if a.AverageEvaluation > b.AverageEvaluation {
			return -1
		} else {
			return 0
		}
	})

	return nodes
}
