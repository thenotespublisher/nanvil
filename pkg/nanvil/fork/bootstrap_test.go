package fork_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nspcc-dev/neo-go/internal/testserdes"
	"github.com/nspcc-dev/neo-go/pkg/config/netmode"
	"github.com/nspcc-dev/neo-go/pkg/core"
	"github.com/nspcc-dev/neo-go/pkg/core/block"
	"github.com/nspcc-dev/neo-go/pkg/core/storage"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/accounts"
	nanvilcfg "github.com/nspcc-dev/neo-go/pkg/nanvil/config"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/fork"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const testForkMagic = uint32(860833102)

func buildForkBlock(t *testing.T, index uint32, magic uint32) *block.Block {
	t.Helper()
	val, err := accounts.NewValidatorSigner()
	require.NoError(t, err)
	blk := &block.Block{
		Header: block.Header{
			PrevHash:      util.Uint256{byte(index)},
			Index:         index,
			Timestamp:     uint64(index) * 1000,
			NextConsensus: val.ScriptHash(),
			Script: transaction.Witness{
				VerificationScript: val.Script(),
			},
		},
	}
	blk.RebuildMerkleRoot()
	blk.Script.InvocationScript = val.SignHashable(magic, blk)
	return blk
}

func newForkRPCServer(t *testing.T, forkBlk *block.Block, root util.Uint256) *httptest.Server {
	t.Helper()
	blockRaw, err := testserdes.EncodeBinary(forkBlk)
	require.NoError(t, err)
	blockB64 := base64.StdEncoding.EncodeToString(blockRaw)
	index := forkBlk.Index
	blockHash := forkBlk.Hash()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "getstateheight":
			_, _ = w.Write([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"result":{"localrootindex":%d,"validatedrootindex":%d}}`, index, index)))
		case "getversion":
			_, _ = w.Write([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"result":{"protocol":{"network":%d,"addressversion":53}}}`, testForkMagic)))
		case "getblockhash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x` + blockHash.StringBE() + `"}`))
		case "getstateroot":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"version":0,"index":` + fmt.Sprint(index) + `,"roothash":"0x` + root.StringBE() + `","witnesses":[]}}`))
		case "getnativecontracts":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"id":-1,"hash":"0x` + util.Uint160{8}.StringBE() + `","manifest":{"name":"ContractManagement"}}]}`))
		case "findstates":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"results":[],"truncated":false}}`))
		case "getblock":
			_, _ = w.Write([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"result":"%s"}`, blockB64)))
		default:
			t.Fatalf("unexpected method %s", req.Method)
		}
	}))
}

func TestBootstrapForkChain(t *testing.T) {
	const index uint32 = 10
	root := util.Uint256{2}
	forkBlk := buildForkBlock(t, index, testForkMagic)
	srv := newForkRPCServer(t, forkBlk, root)
	defer srv.Close()

	ctx := context.Background()
	manifest, err := fork.CreateBranch(ctx, srv.URL, index)
	require.NoError(t, err)
	require.Equal(t, index, manifest.Index)

	opts := nanvilcfg.DefaultStartOptions()
	bcCfg := nanvilcfg.BlockchainConfigForFork(manifest, opts)
	pubHex, err := accounts.ValidatorPublicKeyHex()
	require.NoError(t, err)
	bcCfg.StandbyCommittee = []string{pubHex}

	bc, err := core.NewBlockchain(storage.NewMemoryStore(), bcCfg, zaptest.NewLogger(t))
	require.NoError(t, err)

	overlay := fork.NewTrackingOverlay()
	cache, err := fork.NewDiskCache(t.TempDir(), manifest.NetworkMagic, manifest.Index)
	require.NoError(t, err)
	remote, err := fork.NewRemoteStateStore(ctx, manifest, cache, true)
	require.NoError(t, err)
	defer remote.Close()

	mgr, err := accounts.NewManager(opts.Mnemonic, 2)
	require.NoError(t, err)

	require.NoError(t, fork.Bootstrap(ctx, fork.BootstrapOptions{
		Manifest: manifest,
		Remote:   remote,
		Overlay:  overlay,
		Chain:    bc,
		Accounts: mgr,
		Balance:  opts.Balance,
	}))
	require.Equal(t, index, bc.BlockHeight())
	require.Equal(t, netmode.Magic(testForkMagic), bc.GetConfig().Magic)
}
