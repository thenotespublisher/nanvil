package fork_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/nanvil/fork"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestRemoteStorePrefetchAndCache(t *testing.T) {
	contract := util.Uint160{0x11}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "getversion":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"protocol":{"network":42,"addressversion":53}}}`))
		case "getnativecontracts":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"id":1,"hash":"0x` + contract.StringBE() + `","manifest":{"name":"MyContract"}}]}`))
		case "findstates":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"results":[{"key":"0x01","value":"0xab","proof":[]}],"truncated":false}}`))
		default:
			t.Fatalf("unexpected %s", req.Method)
		}
	}))
	defer srv.Close()

	m := &fork.Manifest{
		RPCURL:       srv.URL,
		NetworkMagic: 42,
		Index:        1,
		RootHash:     util.Uint256{1},
		Contracts:    []fork.ContractInfo{{ID: 1, Hash: contract}},
	}
	cache, err := fork.NewDiskCache(t.TempDir(), m.NetworkMagic, m.Index)
	require.NoError(t, err)
	require.NotEmpty(t, cache.Dir())
	require.Contains(t, fork.CacheKey(contract, []byte{1}), contract.StringBE())

	rs, err := fork.NewRemoteStateStore(context.Background(), m, cache, false)
	require.NoError(t, err)
	defer rs.Close()

	require.NoError(t, rs.PrefetchContract(contract))
	require.Equal(t, 1, rs.CachedCount())

	h, ok := rs.ContractHash(1)
	require.True(t, ok)
	require.NotEqual(t, util.Uint160{}, h)
}

func TestEnumerateContractsFromMock(t *testing.T) {
	mgmt := util.Uint160{8}
	contractHash := util.Uint160{0x22}
	keyBytes := []byte{0x08, 0, 0, 0, 5}
	valBytes := contractHash.BytesBE()
	keyB64 := base64.StdEncoding.EncodeToString(keyBytes)
	valB64 := base64.StdEncoding.EncodeToString(valBytes)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "getstateheight":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"localrootindex":5,"validatedrootindex":5}}`))
		case "getversion":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"protocol":{"network":860833102,"addressversion":53}}}`))
		case "getblockhash":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x` + util.Uint256{1}.StringBE() + `"}`))
		case "getstateroot":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"version":0,"index":5,"roothash":"0x` + util.Uint256{2}.StringBE() + `","witnesses":[]}}`))
		case "getnativecontracts":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[{"id":-1,"hash":"0x` + mgmt.StringBE() + `","manifest":{"name":"ContractManagement"}}]}`))
		case "findstates":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"results":[{"key":"` + keyB64 + `","value":"` + valB64 + `","proof":[]}],"truncated":false}}`))
		default:
			t.Fatalf("unexpected %s", req.Method)
		}
	}))
	defer srv.Close()

	m, err := fork.CreateBranch(context.Background(), srv.URL, 5)
	require.NoError(t, err)
	require.Len(t, m.Contracts, 1)
	require.Equal(t, int32(5), m.Contracts[0].ID)
}
