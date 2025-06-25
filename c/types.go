package c

type Role string

type Champion struct {
	Name string
	Role Role
	Id   byte
}

type ScoredTrieNode struct {
	AverageEvaluation int16
	Children          map[byte]*ScoredTrieNode
}

// TODO: Somehow compute every possible combination regardless of the player's pools.
type TeamSelectableChampions struct {
	PickableChampions        map[byte]bool // The union of the champions in our pools.
	BannableChampions        map[byte]bool // The union of the champions in the opposing team's pools.
	PickableSupportChampions map[byte]bool // Only the supports from PickableChampions.
	BannableSupportChampions map[byte]bool // Only the supports from BannableChampions.
}

type Player struct {
	Name         string
	ChampionPool map[byte]int8 // Champion ID to Evaluation.
}

type Team struct {
	Pick1Pool []int8 // Used to build selectable champions.
	Pick2Pool []int8 // All combinations of 2 picks the team can play. Used for pruning compositions the team doesn't play.
	Pick3Pool []int8 // All combinations of 3 picks the team can play mapped to the best eval.
}
