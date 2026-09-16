package worktree

import "strings"

// fuzzyScore ranks how well query matches text. It returns 0 when text does
// not contain every rune of query in order (case-insensitive). A higher score
// is a better match: consecutive matches and matches after a separator score
// more, and shorter texts win ties so a tight match beats an incidental one.
func fuzzyScore(query, text string) int {
	if query == "" {
		return 0
	}
	q := []rune(strings.ToLower(query))
	t := []rune(strings.ToLower(text))
	score := 0
	queryIndex := 0
	run := 0
	for textIndex := 0; textIndex < len(t) && queryIndex < len(q); textIndex++ {
		if t[textIndex] != q[queryIndex] {
			run = 0
			continue
		}
		bonus := 1 + 2*run
		if textIndex == 0 || isSeparator(t[textIndex-1]) {
			bonus += 3
		}
		score += bonus
		run++
		queryIndex++
	}
	if queryIndex < len(q) {
		return 0
	}
	return score*100 - len(t)
}

func isSeparator(r rune) bool {
	return strings.ContainsRune("/-_. ", r)
}
