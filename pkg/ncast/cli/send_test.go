package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCallArgs(t *testing.T) {
	args, err := parseCallArgs([]string{"true", "false", "42", "hello", "0x0102"})
	require.NoError(t, err)
	require.Len(t, args, 5)
	require.Equal(t, true, args[0])
	require.Equal(t, false, args[1])
	require.Equal(t, int64(42), args[2])
	require.Equal(t, "hello", args[3])
	require.Equal(t, "0x0102", args[4])

	_, err = parseCallArgs([]string{"NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg"})
	require.NoError(t, err)
}

func TestTrimHex(t *testing.T) {
	require.Equal(t, "abcd", trimHex("0xabcd"))
	require.Equal(t, "abcd", trimHex("0Xabcd"))
	require.Equal(t, "abcd", trimHex("abcd"))
}
