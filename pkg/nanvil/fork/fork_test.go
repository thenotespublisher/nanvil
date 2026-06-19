package fork_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/nanvil/fork"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestManifestSaveLoad(t *testing.T) {
	m := &fork.Manifest{
		RPCURL:       "http://localhost:20331",
		NetworkMagic: 860833102,
		Index:        100,
		IndexHash:    util.Uint256{1},
		RootHash:     util.Uint256{2},
		Contracts:    []fork.ContractInfo{{ID: 1, Hash: util.Uint160{3}}},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "fork.json")
	require.NoError(t, m.Save(path))
	loaded, err := fork.LoadManifest(path)
	require.NoError(t, err)
	require.Equal(t, m.RPCURL, loaded.RPCURL)
	require.Equal(t, m.Index, loaded.Index)
}

func TestTrackingOverlay(t *testing.T) {
	o := fork.NewTrackingOverlay()
	o.Put("k1", []byte("v1"))
	v, ok := o.Get("k1")
	require.True(t, ok)
	require.Equal(t, []byte("v1"), v)
	o.Delete("k1")
	_, ok = o.Get("k1")
	require.True(t, ok) // deleted marker
	o.Reset()
	require.Equal(t, 0, o.Len())
}

func TestDiskCache(t *testing.T) {
	dir := t.TempDir()
	c, err := fork.NewDiskCache(dir, 42, 100)
	require.NoError(t, err)
	require.NoError(t, c.Put("key", []byte("val")))
	v, ok := c.Get("key")
	require.True(t, ok)
	require.Equal(t, []byte("val"), v)
	require.NotEmpty(t, c.Dir())
	require.Contains(t, fork.CacheKey(util.Uint160{1}, []byte{2}), "_02")
}

func TestTrackingOverlayExportImport(t *testing.T) {
	o := fork.NewTrackingOverlay()
	o.Put("a", []byte("1"))
	o.Delete("b")
	w, d := o.Export()
	require.Len(t, w, 1)
	require.Len(t, d, 1)

	o2 := fork.NewTrackingOverlay()
	o2.Import(w, d)
	v, ok := o2.Get("a")
	require.True(t, ok)
	require.Equal(t, []byte("1"), v)
	_, del := o2.Get("b")
	require.True(t, del)
}

func TestManifestJSON(t *testing.T) {
	raw, err := json.Marshal(fork.Manifest{Index: 1})
	require.NoError(t, err)
	require.Contains(t, string(raw), "index")
	_ = os.Getenv // keep import
}
