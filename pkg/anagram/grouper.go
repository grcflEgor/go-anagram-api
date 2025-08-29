package anagram

import (
	"context"
	"sort"
	"strings"
	"unicode"
)

func normalizeWord(word string) string {
	lowerWord := strings.ToLower(word)

	runes := make([]rune, 0, len(lowerWord))

	for _, r := range lowerWord {
		if !unicode.IsSpace(r) {
			runes = append(runes, r)
		}
	}

	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	return string(runes)
}

func Group(ctx context.Context, words []string) (map[string][]string, error) {
	groups := make(map[string][]string)

	for _, word := range words {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if word == "" {
			continue
		}

		key := normalizeWord(word)
		
		groups[key] = append(groups[key], word)
	}

	return groups, nil
}