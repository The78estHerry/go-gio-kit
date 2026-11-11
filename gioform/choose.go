package gioform

import "strings"

// filterOptions returns the indexes of the options whose label or value
// contains query, ignoring case; all of them when query is blank.
func filterOptions(opts []Option, query string) []int {
	q := strings.ToLower(strings.TrimSpace(query))
	idx := make([]int, 0, len(opts))
	for i, o := range opts {
		if q == "" || strings.Contains(strings.ToLower(o.Label), q) || strings.Contains(strings.ToLower(o.Value), q) {
			idx = append(idx, i)
		}
	}
	return idx
}

// optionIndex returns the index of the option with value v, or -1.
func optionIndex(opts []Option, v string) int {
	for i, o := range opts {
		if o.Value == v {
			return i
		}
	}
	return -1
}

// moveHighlight moves a highlight in a list of n items by dir and keeps it
// inside the list; -1 means nothing is highlighted.
func moveHighlight(hl, n, dir int) int {
	if n == 0 {
		return -1
	}
	return max(0, min(hl+dir, n-1))
}
