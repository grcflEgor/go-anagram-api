package anagram

import (
	"context"
	"sort"
	"strings"
	"unicode"
)

func normalizeWord(word string, caseSensitive bool) string {
	var base string

	if caseSensitive {
		base = word
	} else {
		base = strings.ToLower(word)
	}

	runes := make([]rune, 0, len(base))

	for _, r := range base {
		if !unicode.IsSpace(r) {
			runes = append(runes, r)
		}
	}

	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	return string(runes)
}

func Group(ctx context.Context, words []string, caseSensitive bool) (map[string][]string, error) {
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

		key := normalizeWord(word, caseSensitive)

		groups[key] = append(groups[key], word)
	}

	return groups, nil
}
