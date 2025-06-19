package c

const (
	MeleeRole   Role = "Melee"
	RangedRole  Role = "Ranged"
	SupportRole Role = "Support"
)

var (
	//DraftOrder2 = []string{"T1GB1", "T2GB1", "T2B2", "T1B2", "T1P1", "T2P1", "T2P2", "T1P2", "T1B3", "T2B3", "T1P3", "T2P3"}
	DraftOrder = []string{"T1GB", "T2GB", "T2B", "T1B", "T1P", "T2P", "T2P", "T1P", "T1B", "T2B", "T1P", "T2P"}
	T1PIdxs    = []int{4, 7, 10}
	T2PIdxs    = []int{5, 6, 11}
)
