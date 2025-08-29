package anagram

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestGroup(t *testing.T) {
	testCases := []struct {
		name     string
		words    []string
		expected map[string][]string
	}{
		{
			name:  "Base case from TZ",
			words: []string{"ток", "рост", "кот", "торс", "Кто", "фывап", "рок"},
			expected: map[string][]string{
				"кот":   {"ток", "кот", "Кто"},
				"орст":  {"рост", "торс"},
				"авпфы": {"фывап"},
				"кор":   {"рок"},
			},
		},
		{
			name:     "Empty input slice",
			words:    []string{},
			expected: map[string][]string{},
		},
		{
			name:  "No anagrams",
			words: []string{"hello", "world"},
			expected: map[string][]string{
				"ehllo": {"hello"},
				"dlorw": {"world"},
			},
		},
		{
			name:  "Words with special characters and numbers",
			words: []string{"ав12", "21ва", "тест"},
			expected: map[string][]string{
				"12ав": {"ав12", "21ва"},
				"естт": {"тест"},
			},
		},
		{
			name:  "Slice with empty strings",
			words: []string{"first", "", "second", ""},
			expected: map[string][]string{
				"first":  {"first"},
				"cdenos": {"second"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Group(context.Background(), tc.words)
			if err != nil {
				t.Fatalf("Group() error = %v", err)
			}

			for k := range result {
				sort.Strings(result[k])
			}
			for k := range tc.expected {
				sort.Strings(tc.expected[k])
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Group() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func BenchmarkGroup(b *testing.B) {
	words := []string{"ток", "рост", "кот", "торс", "Кто", "фывап", "рок", "сор", "рот", "кофе"}
	largeInput := make([]string, 0, 10000)
	for i := 0; i < 1000; i++ {
		largeInput = append(largeInput, words...)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Group(context.Background(), largeInput)
	}
}

func FuzzNormalizeWord(f *testing.F) {
	f.Add("hello")
	f.Add("World")
	f.Add("Тест")
	f.Add("123!@#")
	f.Add("test one")
	f.Add("   leading and trailing spaces   ")

	f.Fuzz(func(t *testing.T, word string) {
		normalized := normalizeWord(word)

		if strings.ContainsRune(normalized, ' ') {
			t.Errorf("Нормализованная строка содержит пробел: %q", normalized)
		}

		runes := []rune(normalized)
		if !sort.SliceIsSorted(runes, func(i, j int) bool {
			return runes[i] < runes[j]
		}) {
			t.Errorf("Нормализованная строка не отсортирована: %q", normalized)
		}
	})
}
