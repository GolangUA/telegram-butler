package validator

import "strings"

func New(forbiddenWords []string) *ByName {
	for i, w := range forbiddenWords {
		forbiddenWords[i] = strings.ToLower(w)
	}

	return &ByName{
		forbiddenWords: forbiddenWords,
	}
}

type ByName struct {
	forbiddenWords []string
}

func (v *ByName) Validate(names ...string) bool {
	for _, name := range names {
		lowerName := strings.ToLower(name)
		for _, word := range v.forbiddenWords {
			if strings.Contains(lowerName, word) {
				return false
			}
		}
	}
	return true
}
