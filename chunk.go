package web

import (
	"strings"
	"unicode"
)

// Chunk splits a string into chunks of approximately the given size. Attempts
// to split on periods or spaces if present, near the split points.
func Chunk(text string, size int) []string {
	if size < 2 {
		size = 2
	}
	windowSize := size / 4
	var chunks []string
	runes := []rune(text)
	for {
		if len(runes) <= size {
			chunks = append(chunks, string(runes))
			break
		}
		cutoff := size - 1
		minCutoff := size - windowSize
		if minCutoff < 0 {
			minCutoff = 0
		}
		// first look for a period
		found := false
		for cutoff > minCutoff {
			if cutoff < len(runes) && runes[cutoff] == '.' {
				found = true
				break
			}
			cutoff--
		}
		// if no period found, look for a space
		if !found {
			cutoff = size
			for cutoff > minCutoff {
				if cutoff < len(runes) && unicode.IsSpace(runes[cutoff]) {
					found = true
					break
				}
				cutoff--
			}
		}
		if !found {
			cutoff = size
		} else {
			cutoff += 1
		}
		chunks = append(chunks, string(runes[:cutoff]))
		runes = runes[cutoff:]
	}
	for i, chunk := range chunks {
		chunks[i] = strings.TrimSpace(chunk)
	}
	return chunks
}
