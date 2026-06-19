package cli_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/nanvil/accounts"
	nanvilcfg "github.com/nspcc-dev/neo-go/pkg/nanvil/config"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/node"
	ncastcli "github.com/nspcc-dev/neo-go/pkg/ncast/cli"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func startNode(t *testing.T) (*node.DevNode, string) {
	t.Helper()
	opts := nanvilcfg.DefaultStartOptions()
	opts.Port = 0
	opts.Explorer = false
	opts.NoMining = true
	opts.Accounts = 2
	n, err := node.NewDevNode(opts, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, n.Start(context.Background()))
	t.Cleanup(n.Shutdown)
	return n, "http://" + n.RPCAddr
}

func startRPC(t *testing.T) string {
	_, rpc := startNode(t)
	return rpc
}

func runCLIWithCapture(t *testing.T, rpc string, args ...string) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	full := append([]string{"ncast", "--rpc", rpc}, args...)
	runErr := ncastcli.NewApp().Run(full)
	require.NoError(t, w.Close())
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}

func TestBlockNumberAgainstDevNode(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "block-number")
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestRPCGetBlockCount(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "rpc", "getblockcount", "[]")
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestContractGasOnDevNode(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "--json", "contract", "gas")
	require.NoError(t, err)
	require.Contains(t, out, "GasToken")
}

func TestChainIDCmd(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "chain-id")
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestMempoolCmd(t *testing.T) {
	rpc := startRPC(t)
	_, err := runCLIWithCapture(t, rpc, "mempool")
	require.NoError(t, err)
}

func TestBalanceCmd(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "balance", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.NoError(t, err)
	require.Contains(t, out, "GAS")
}

func TestRPCJSONOutput(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "--json", "block-number")
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestBlockByIndex(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "block", "0")
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(out), "index")
}

func TestEstimateGasBalanceOf(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "estimate", "gas", "balanceOf", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestBalanceJSON(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "--json", "balance", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.NoError(t, err)
	require.Contains(t, out, "balance")
}

func TestMempoolJSON(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "--json", "mempool")
	require.NoError(t, err)
	require.Contains(t, out, "[")
}

func TestBlockFullJSON(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "--json", "block", "0", "--full")
	require.NoError(t, err)
	require.Contains(t, out, "index")
}

func TestCallCmd(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "call", "gas", "balanceOf", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(out), "halt")
}

func TestTxCmdAfterTransfer(t *testing.T) {
	n, rpc := startNode(t)
	mgr, err := accounts.NewManager(n.Opts.Mnemonic, n.Opts.Accounts)
	require.NoError(t, err)
	tx, err := mgr.SignedGASTransfer(n.Chain, mgr.Accounts[0], mgr.Accounts[1].Signer.ScriptHash(), 1_0000_0000)
	require.NoError(t, err)
	require.NoError(t, n.NetServer.RelayTxn(tx))
	require.NoError(t, n.MineBlock(1))

	out, err := runCLIWithCapture(t, rpc, "tx", "0x"+tx.Hash().StringLE())
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestSendCallMissingWIF(t *testing.T) {
	rpc := startRPC(t)
	_, err := runCLIWithCapture(t, rpc, "send-call", "gas", "transfer", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg", "1", "NRBs4Wc8vuFJY5sN7WvJNRzVduAeXQ8TQg")
	require.Error(t, err)
}

func TestTxCmdJSON(t *testing.T) {
	n, rpc := startNode(t)
	mgr, err := accounts.NewManager(n.Opts.Mnemonic, n.Opts.Accounts)
	require.NoError(t, err)
	tx, err := mgr.SignedGASTransfer(n.Chain, mgr.Accounts[0], mgr.Accounts[1].Signer.ScriptHash(), 1_0000_0000)
	require.NoError(t, err)
	require.NoError(t, n.NetServer.RelayTxn(tx))
	require.NoError(t, n.MineBlock(1))

	out, err := runCLIWithCapture(t, rpc, "--json", "tx", "0x"+tx.Hash().StringLE())
	require.NoError(t, err)
	require.Contains(t, out, "hash")
}

func TestCallCmdJSON(t *testing.T) {
	rpc := startRPC(t)
	out, err := runCLIWithCapture(t, rpc, "--json", "call", "gas", "decimals")
	require.NoError(t, err)
	require.Contains(t, out, "stack")
}

func TestSendWithWIF(t *testing.T) {
	n, rpc := startNode(t)
	mgr, err := accounts.NewManager(n.Opts.Mnemonic, n.Opts.Accounts)
	require.NoError(t, err)
	out, err := runCLIWithCapture(t, rpc, "send",
		"--wif", mgr.Accounts[0].WIF,
		mgr.Accounts[1].Address, "1",
	)
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}

func TestHashCmdScript(t *testing.T) {
	out, err := runCLIWithCapture(t, "http://127.0.0.1:1", "hash", "--type", "script", strings.Repeat("ab", 33))
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
}
