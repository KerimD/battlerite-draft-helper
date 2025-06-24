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

func PrintTree(idToChampionName map[byte]Champion, node *ScoredTrieNode, maxDepth int) {
	fmt.Println("T1 Pick Eval:", node.AverageEvaluation)
	printTree(idToChampionName, node, maxDepth, 0)
}

func printTree(idToChampionName map[byte]Champion, node *ScoredTrieNode, maxDepth, currentDepth int) {
	if node == nil || currentDepth > maxDepth {
		return
	}

	indent := ""
	for i := 0; i < currentDepth; i++ {
		indent += "|  "
	}

	for championId, childNode := range node.Children {
		fmt.Printf("%s%s %s: %.4f\n", indent, DraftOrder[currentDepth], idToChampionName[championId].Name, childNode.AverageEvaluation)
	}

	sortedList := mapToSortedSlice(node.Children)

	for _, childNode := range sortedList {
		printTree(idToChampionName, &childNode, maxDepth, currentDepth+1)
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
