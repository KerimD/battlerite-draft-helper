package c

type Role string

type Champion struct {
	Name string
	Role Role
	Id   string
}

type ScoredTrieNode struct {
	ChampionName      string
	AverageEvaluation int
	CompletedStates   []string
	Children          map[string]*ScoredTrieNode
}

type TeamSelectableChampions struct {
	PickableChampions        map[string]bool
	BannableChampions        map[string]bool
	PickableSupportChampions map[string]bool
	BannableSupportChampions map[string]bool
}
