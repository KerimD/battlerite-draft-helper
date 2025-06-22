package c

type Role string

type Champion struct {
	Name string
	Role Role
	Id   byte
}

type ScoredTrieNode struct {
	AverageEvaluation float32
	Children          map[byte]*ScoredTrieNode
}

type TeamSelectableChampions struct {
	PickableChampions        map[byte]bool
	BannableChampions        map[byte]bool
	PickableSupportChampions map[byte]bool
	BannableSupportChampions map[byte]bool
}

type Player struct {
	Name         string
	ChampionPool map[byte]int8 // Champion ID to Evaluation.
}

type Team struct {
	Pick1Pool []int8
	Pick2Pool []int8
	Pick3Pool []int8
}
