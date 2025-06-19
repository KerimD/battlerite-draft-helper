package c

type Role string

type Champion struct {
	Name string
	Role Role
	Id   byte
}

type ScoredTrieNode struct {
	ChampionName      string
	AverageEvaluation int
	CompletedStates   [][]byte
	Children          map[byte]*ScoredTrieNode
}

type TeamSelectableChampions struct {
	PickableChampions        map[byte]bool
	BannableChampions        map[byte]bool
	PickableSupportChampions map[byte]bool
	BannableSupportChampions map[byte]bool
}
