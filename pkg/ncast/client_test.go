package ncast_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	nanvilcfg "github.com/nspcc-dev/neo-go/pkg/nanvil/config"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/node"
	"github.com/nspcc-dev/neo-go/pkg/ncast"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResolveHash160(t *testing.T) {
	h, err := ncast.ResolveHash160("NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.NoError(t, err)
	expected, err := address.StringToUint160("NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.NoError(t, err)
	require.Equal(t, expected, h)

	hexHash := "0x" + expected.StringLE()
	h, err = ncast.ResolveHash160(hexHash)
	require.NoError(t, err)
	require.Equal(t, expected, h)
}

func TestResolveHash256(t *testing.T) {
	raw := util.Uint256{1, 2, 3, 4}
	h, err := ncast.ResolveHash256("0x" + raw.StringLE())
	require.NoError(t, err)
	require.Equal(t, raw, h)
}

func TestWSURL(t *testing.T) {
	ws, err := ncast.WSURL("http://127.0.0.1:8545")
	require.NoError(t, err)
	require.Equal(t, "ws://127.0.0.1:8545/ws", ws)

	ws, err = ncast.WSURL("https://example.com/rpc")
	require.NoError(t, err)
	require.Equal(t, "wss://example.com/rpc/ws", ws)

	_, err = ncast.WSURL("ftp://example.com")
	require.Error(t, err)
}

func TestFormatHelpers(t *testing.T) {
	require.Equal(t, "—", ncast.FormatTimeMs(0))
	require.NotEqual(t, "—", ncast.FormatTimeMs(1_700_000_000_000))
	require.Equal(t, "1 GAS", ncast.FormatGAS(100_000_000))
}

func TestRPCClientAgainstDevNode(t *testing.T) {
	opts := nanvilcfg.DefaultStartOptions()
	opts.Port = 0
	opts.Explorer = false
	opts.NoMining = true
	opts.Accounts = 2

	n, err := node.NewDevNode(opts, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, n.Start(context.Background()))
	defer n.Shutdown()

	endpoint := "http://" + n.RPCAddr
	c, err := ncast.RPCClient(context.Background(), endpoint)
	require.NoError(t, err)
	defer c.Close()

	height, err := c.GetBlockCount()
	require.NoError(t, err)
	require.Greater(t, height, uint32(0))

	gasHash, err := ncast.ResolveContract(c, "gas")
	require.NoError(t, err)
	require.NotEqual(t, util.Uint160{}, gasHash)

	devAcc := "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg"
	raw, err := ncast.RawRPC(endpoint, "getnep17balances", []any{devAcc})
	require.NoError(t, err)
	var balances struct {
		Balance []struct {
			Assethash string `json:"assethash"`
		} `json:"balance"`
	}
	require.NoError(t, json.Unmarshal(raw, &balances))
	require.NotEmpty(t, balances.Balance)

	blockHash, err := c.GetBlockHash(1)
	require.NoError(t, err)
	blk, err := c.GetBlockByHash(blockHash)
	require.NoError(t, err)
	require.Equal(t, uint32(1), blk.Index)
}

func TestRawRPCError(t *testing.T) {
	_, err := ncast.RawRPC("http://127.0.0.1:1", "getblockcount", nil)
	require.Error(t, err)
}

func TestParseHash256Alias(t *testing.T) {
	raw := util.Uint256{9}
	h, err := ncast.ParseHash256("0x" + raw.StringLE())
	require.NoError(t, err)
	require.Equal(t, raw, h)
}

func TestResolveContractEmpty(t *testing.T) {
	opts := nanvilcfg.DefaultStartOptions()
	opts.Port = 0
	opts.Explorer = false
	n, err := node.NewDevNode(opts, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, n.Start(context.Background()))
	defer n.Shutdown()

	c, err := ncast.RPCClient(context.Background(), "http://"+n.RPCAddr)
	require.NoError(t, err)
	defer c.Close()

	_, err = ncast.ResolveContract(c, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty contract")

	_, err = ncast.ResolveContract(c, "definitely-not-a-contract")
	require.Error(t, err)
}

func TestDefaultRPCConstant(t *testing.T) {
	require.Equal(t, "http://127.0.0.1:8545", ncast.DefaultRPC)
}

func TestPrintHelpers(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	require.NoError(t, ncast.PrintJSON(map[string]int{"x": 1}))
	ncast.PrintKV("k", "v")
	require.NoError(t, w.Close())
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	require.Contains(t, buf.String(), `"x"`)
	require.Contains(t, buf.String(), "k")
	require.Equal(t, "—", ncast.FormatTimeMs(0))
}
