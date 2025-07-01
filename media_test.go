package web

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsMediaURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "image file",
			url:      "https://example.com/image.jpg",
			expected: true,
		},
		{
			name:     "video file",
			url:      "https://example.com/video.mp4",
			expected: true,
		},
		{
			name:     "audio file",
			url:      "https://example.com/audio.mp3",
			expected: true,
		},
		{
			name:     "document file",
			url:      "https://example.com/doc.pdf",
			expected: true,
		},
		{
			name:     "uppercase extension",
			url:      "https://example.com/IMAGE.JPG",
			expected: true,
		},
		{
			name:     "html file",
			url:      "https://example.com/page.html",
			expected: false,
		},
		{
			name:     "no extension",
			url:      "https://example.com/page",
			expected: false,
		},
		{
			name:     "path with dot but no extension",
			url:      "https://example.com/path.with.dots/page",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse(tt.url)
			result := IsMediaURL(u)
			require.Equal(t, tt.expected, result)
		})
	}
}
