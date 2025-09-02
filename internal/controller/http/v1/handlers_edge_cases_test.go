package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlersEdgeCases(t *testing.T) {
	t.Run("GroupAnagrams_EdgeCases_Tabular", func(t *testing.T) {
		cases := []struct {
			name          string
			words         []string
			caseSensitive bool
		}{
			{
				name: "VeryLargeWordsArray",
				words: func() []string {
					w := make([]string, 10000)
					for i := range w {
						w[i] = fmt.Sprintf("word%d", i)
					}
					return w
				}(),
				caseSensitive: false,
			},
			{
				name:          "SpecialCharacters",
				words:         []string{"!@#$%^&*()", "test", "привет", "123"},
				caseSensitive: false,
			},
			{
				name:          "UnicodeCharacters",
				words:         []string{"café", "naïve", "résumé", "test"},
				caseSensitive: false,
			},
			{
				name:          "MixedCaseSensitivity",
				words:         []string{"Test", "test", "TEST"},
				caseSensitive: true,
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mockService, _, handlers := setupTestHandlers()
				mockService.On("CreateTask", mock.Anything, tc.words, tc.caseSensitive).Return("task123", nil)

				request := GroupRequest{Words: tc.words, CaseSensitive: tc.caseSensitive}
				req := createJSONRequest("POST", "/api/v1/anagrams/group", request)
				rec := httptest.NewRecorder()

				handlers.GroupAnagrams(rec, req)

				assert.Equal(t, http.StatusAccepted, rec.Code)
				mockService.AssertExpectations(t)
			})
		}
	})

	t.Run("UploadFile_EdgeCases_Tabular", func(t *testing.T) {
		cases := []struct {
			name          string
			fileContent   string
			expectedWords []string
			caseSensitive bool
		}{
			{
				name:          "SpecialCharacters",
				fileContent:   "!@#$%\n%$#@!\nпривет\nтевирп",
				expectedWords: []string{"!@#$%", "%$#@!", "привет", "тевирп"},
				caseSensitive: false,
			},
			{
				name:          "UnicodeCharacters",
				fileContent:   "café\néfac\nnaïve\nnaïve",
				expectedWords: []string{"café", "éfac", "naïve", "naïve"},
				caseSensitive: false,
			},
			{
				name:          "MixedCaseSensitivity",
				fileContent:   "Test\ntest\nTEST",
				expectedWords: []string{"Test", "test", "TEST"},
				caseSensitive: true,
			},
			{
				name:          "EmptyLinesInFile",
				fileContent:   "word1\n\nword2\n\n\nword3",
				expectedWords: []string{"word1", "word2", "word3"},
				caseSensitive: false,
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mockService, _, handlers := setupTestHandlers()
				mockService.On("CreateTask", mock.Anything, tc.expectedWords, tc.caseSensitive).Return("task123", nil)

				req := createMultipartRequest("test.txt", tc.fileContent, tc.caseSensitive)
				rec := httptest.NewRecorder()

				handlers.UploadFile(rec, req)

				assert.Equal(t, http.StatusAccepted, rec.Code)
				mockService.AssertExpectations(t)
			})
		}
	})

	t.Run("Validation_EdgeCases_Tabular", func(t *testing.T) {
		cases := []struct {
			name          string
			words         []string
			caseSensitive bool
		}{
			{
				name:          "SpecialCharacters",
				words:         []string{"!@#$%^&*()", "test", "привет", "123"},
				caseSensitive: false,
			},
			{
				name:          "UnicodeCharacters",
				words:         []string{"café", "naïve", "résumé", "test"},
				caseSensitive: false,
			},
			{
				name:          "VeryLongWord",
				words:         []string{strings.Repeat("a", 1000), "test"},
				caseSensitive: false,
			},
			{
				name:          "NumbersAndSymbols",
				words:         []string{"123", "321", "!@#", "#@!", "test"},
				caseSensitive: false,
			},
		}
		validator := validator.New()
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				request := GroupRequest{Words: tc.words, CaseSensitive: tc.caseSensitive}
				err := validator.Struct(request)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("JSONSerialization_EdgeCases_Tabular", func(t *testing.T) {
		cases := []struct {
			name  string
			value interface{}
		}{
			{
				name:  "SpecialCharacters",
				value: GroupRequest{Words: []string{"!@#$%", "привет", "café", "123"}, CaseSensitive: true},
			},
			{
				name:  "EmptyResult",
				value: GroupResponse{TaskID: "task123", Status: "completed", Result: [][]string{}, ProcessingTime: 0, GroupsCount: 0},
			},
			{
				name:  "VeryLongStrings",
				value: GroupRequest{Words: []string{strings.Repeat("a", 1000), "test"}, CaseSensitive: false},
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(tc.value)
				require.NoError(t, err)

				var unmarshaled interface{}
				err = json.Unmarshal(data, &unmarshaled)
				require.NoError(t, err)

				_, err = json.Marshal(unmarshaled)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Performance_EdgeCases_Tabular", func(t *testing.T) {
		cases := []struct {
			name          string
			words         []string
			fileContent   string
			useFile       bool
			caseSensitive bool
		}{
			{
				name: "LargeNumberOfWords",
				words: func() []string {
					w := make([]string, 5000)
					for i := range w {
						w[i] = fmt.Sprintf("word%d", i)
					}
					return w
				}(),
				caseSensitive: false,
				useFile:       false,
			},
			{
				name: "LargeFileUpload",
				fileContent: func() string {
					b := strings.Builder{}
					for i := 0; i < 1000; i++ {
						b.WriteString(fmt.Sprintf("word%d\n", i))
					}
					return b.String()
				}(),
				words: func() []string {
					w := make([]string, 1000)
					for i := range w {
						w[i] = fmt.Sprintf("word%d", i)
					}
					return w
				}(),
				caseSensitive: false,
				useFile:       true,
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mockService, _, handlers := setupTestHandlers()
				mockService.On("CreateTask", mock.Anything, tc.words, tc.caseSensitive).Return("task123", nil)
				var req *http.Request
				if tc.useFile {
					req = createMultipartRequest("large.txt", tc.fileContent, tc.caseSensitive)
				} else {
					request := GroupRequest{Words: tc.words, CaseSensitive: tc.caseSensitive}
					req = createJSONRequest("POST", "/api/v1/anagrams/group", request)
				}
				rec := httptest.NewRecorder()

				if tc.useFile {
					handlers.UploadFile(rec, req)
				} else {
					handlers.GroupAnagrams(rec, req)
				}

				assert.Equal(t, http.StatusAccepted, rec.Code)
				mockService.AssertExpectations(t)
			})
		}
	})
}
