package validator

import "testing"

func TestValidateByName(t *testing.T) {
	tc := []struct {
		name           string
		forbiddenWords []string
		words          []string
		expected       bool
	}{
		{
			name:           "Should forbid russian flag emoji",
			forbiddenWords: []string{"🇷🇺"},
			words:          []string{"🇷🇺"},
			expected:       false,
		},
		{
			name:           "Should forbid russian matryoshka emoji",
			forbiddenWords: []string{"🪆"},
			words:          []string{"🪆🇺"},
			expected:       false,
		},
		{
			name:           "Should forbid word 'Russia'",
			forbiddenWords: []string{"Russia"},
			words:          []string{"Russia"},
			expected:       false,
		},
		{
			name:           "Should forbid word 'russia'",
			forbiddenWords: []string{"russia"},
			words:          []string{"russia"},
			expected:       false,
		},
		{
			name:           "Should allow 'Vadym' name to bypass",
			forbiddenWords: []string{"russia"},
			words:          []string{"Vadym"},
			expected:       true,
		},
		{
			name:           "Should allow 'Ruslan' name to bypass",
			forbiddenWords: []string{"russia"},
			words:          []string{"Ruslan"},
			expected:       true,
		},
		{
			name:           "Should allow anything when forbidden words are empty",
			forbiddenWords: []string{},
			words:          []string{"russia", "Russia"},
			expected:       true,
		},
		{
			name:           "Should allow anything when nil slice of forbidden words are passed",
			forbiddenWords: nil,
			words:          []string{"russia", "Russia"},
			expected:       true,
		},
		{
			name:           "Should allow when empty slice of words are passed",
			forbiddenWords: []string{"russia"},
			words:          []string{},
			expected:       true,
		},
		{
			name:           "Should allow when nil slice of words are passed",
			forbiddenWords: []string{"russia"},
			words:          nil,
			expected:       true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New(tt.forbiddenWords)
			if got := v.Validate(tt.words...); got != tt.expected {
				t.Errorf("Expected to get: %v, but got: %v", tt.expected, got)
			}
		})
	}
}
