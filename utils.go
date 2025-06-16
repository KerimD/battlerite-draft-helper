package main

func copyMap(original map[string]bool) map[string]bool {
	newMap := make(map[string]bool, len(original))
	for key, value := range original {
		newMap[key] = value
	}
	return newMap
}
