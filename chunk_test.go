package web

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		size     int
		expected []string
	}{
		{
			name:     "short text",
			text:     "Hello world",
			size:     100,
			expected: []string{"Hello world"},
		},
		{
			name:     "text split at period",
			text:     "First sentence. Second sentence.",
			size:     18,
			expected: []string{"First sentence.", "Second sentence."},
		},
		{
			name:     "text split at space",
			text:     "This is a long sentence without periods that should be split at spaces",
			size:     30,
			expected: []string{"This is a long sentence withou", "t periods that should be split", "at spaces"},
		},
		{
			name:     "text with no good split points",
			text:     "Thisisaverylongword",
			size:     5,
			expected: []string{"Thisi", "saver", "ylong", "word"},
		},
		{
			name:     "empty text",
			text:     "",
			size:     100,
			expected: []string{""},
		},
		{
			name:     "size too small (should default to 2)",
			text:     "Hello",
			size:     1,
			expected: []string{"He", "ll", "o"},
		},
		{
			name:     "exact size match",
			text:     "Hello",
			size:     5,
			expected: []string{"Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.text, tt.size)
			require.Equal(t, tt.expected, result)
		})
	}
}
