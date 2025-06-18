package c

import (
	"fmt"
	"strings"
)

func CopyMap(original map[string]bool) map[string]bool {
	newMap := make(map[string]bool, len(original))
	for key, value := range original {
		newMap[key] = value
	}
	return newMap
}

func CreateTeamSelectableChampions(champions map[string]Champion) TeamSelectableChampions {
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
		PickableChampions:        make(map[string]bool),
		BannableChampions:        make(map[string]bool),
		PickableSupportChampions: make(map[string]bool),
		BannableSupportChampions: make(map[string]bool),
	}
}

func FormatCompletedStates(states []string, idToChampion map[string]Champion) {
	for i, state := range states {
		fmt.Printf("=== State %d ===\n%s\n\n", i+1, FormatCompletedState(state, idToChampion))
	}
}

func FormatCompletedState(state string, idToChampion map[string]Champion) string {
	var t1Bans, t1Picks, t2Bans, t2Picks []string

	for i := 0; i < len(DraftOrder); i++ {
		championId := string(state[i])

		champion, exists := idToChampion[championId]
		championName := "Unknown"
		if exists {
			championName = champion.Name
		}

		draftStep := DraftOrder[i]

		switch {
		case strings.HasPrefix(draftStep, "T1") && (strings.Contains(draftStep, "B") || strings.Contains(draftStep, "GB")):
			// Team 1 ban
			if strings.Contains(draftStep, "GB") {
				championName += " (G)" // Mark global bans
			}
			t1Bans = append(t1Bans, championName)
		case strings.HasPrefix(draftStep, "T1") && strings.Contains(draftStep, "P"):
			// Team 1 pick
			t1Picks = append(t1Picks, championName)
		case strings.HasPrefix(draftStep, "T2") && (strings.Contains(draftStep, "B") || strings.Contains(draftStep, "GB")):
			// Team 2 ban
			if strings.Contains(draftStep, "GB") {
				championName += " (G)" // Mark global bans
			}
			t2Bans = append(t2Bans, championName)
		case strings.HasPrefix(draftStep, "T2") && strings.Contains(draftStep, "P"):
			// Team 2 pick
			t2Picks = append(t2Picks, championName)
		}
	}

	var result strings.Builder
	result.WriteString("Team 1\n")
	result.WriteString(fmt.Sprintf("Bans: %s\n", strings.Join(t1Bans, ", ")))
	result.WriteString(fmt.Sprintf("Picks: %s\n", strings.Join(t1Picks, ", ")))
	result.WriteString("\nTeam 2\n")
	result.WriteString(fmt.Sprintf("Bans: %s\n", strings.Join(t2Bans, ", ")))
	result.WriteString(fmt.Sprintf("Picks: %s", strings.Join(t2Picks, ", ")))

	return result.String()
}
