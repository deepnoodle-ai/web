package web

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocument_Markdown(t *testing.T) {
	doc, err := NewDocument(`
		<html>
			<body>
				<h1>Hello, world!</h1>
				<button>Click me</button>
			</body>
		</html>
	`)
	require.NoError(t, err)

	header := doc.H1()
	require.Equal(t, "Hello, world!", header)
}
