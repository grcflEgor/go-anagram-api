package anagram

import (
	"sort"
	"strings"
)

func normalizeWord(word string) string {
	lowerWord := strings.ToLower(word)

	runes := []rune(lowerWord)

	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	return string(runes)
}

func Group(words []string) map[string][]string {
	groups := make(map[string][]string)

	for _, word := range words {
		if word == "" {
			continue
		}

		key := normalizeWord(word)
		
		groups[key] = append(groups[key], word)
	}

	return groups
}