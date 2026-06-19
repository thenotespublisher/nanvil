package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func runCapture(args []string) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	runErr := NewApp().Run(args)
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}

func TestNewApp(t *testing.T) {
	app := NewApp()
	require.Equal(t, "ncast", app.Name)
	require.Greater(t, len(app.Commands), 5)
}

func TestHashCmd(t *testing.T) {
	out, err := runCapture([]string{"ncast", "hash", "deadbeef"})
	require.NoError(t, err)
	require.Len(t, strings.TrimSpace(out), 64)
}

func TestAddressCmdFromNeoAddress(t *testing.T) {
	out, err := runCapture([]string{"ncast", "address", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg"})
	require.NoError(t, err)
	require.NotEmpty(t, stringsTrim(out))
}

func TestAddressCmdFromScriptHash(t *testing.T) {
	out, err := runCapture([]string{"ncast", "address", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg"})
	require.NoError(t, err)
	back, err := runCapture([]string{"ncast", "address", strings.TrimSpace(out)})
	require.NoError(t, err)
	require.Contains(t, back, "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
}

func TestToFromDatoshi(t *testing.T) {
	out, err := runCapture([]string{"ncast", "to-datoshi", "1.5"})
	require.NoError(t, err)
	require.Equal(t, "150000000\n", out)

	out, err = runCapture([]string{"ncast", "from-datoshi", "100000000"})
	require.NoError(t, err)
	require.Contains(t, out, "1 GAS")
}

func TestUtilCommandsRequireArgs(t *testing.T) {
	_, err := runCapture([]string{"ncast", "hash"})
	require.Error(t, err)
	_, err = runCapture([]string{"ncast", "address"})
	require.Error(t, err)
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}
