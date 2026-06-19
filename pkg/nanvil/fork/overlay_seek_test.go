package fork_test

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/core/storage"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/fork"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestStoreSeekAndSeekGC(t *testing.T) {
	base := storage.NewMemoryStore()
	overlay := fork.NewTrackingOverlay()
	store := fork.NewStore(base, overlay)

	id := int32(-6)
	keyA := makeStorageKey(id, []byte{1})
	keyB := makeStorageKey(id, []byte{2})
	require.NoError(t, base.PutChangeSet(nil, map[string][]byte{
		string(keyA): []byte("a"),
		string(keyB): []byte("b"),
	}))
	fork.PutStorageOverlay(overlay, id, []byte{3}, []byte("overlay"))

	var seen []string
	store.Seek(storage.SeekRange{Prefix: keyA[:5]}, func(k, v []byte) bool {
		seen = append(seen, string(k))
		return true
	})
	require.NotEmpty(t, seen)

	require.NoError(t, store.SeekGC(storage.SeekRange{Prefix: keyA[:5]}, func(k, v []byte) (bool, bool) {
		return string(k) != string(keyA), true
	}))
	_, err := store.Get(keyA)
	require.Error(t, err)
	gotB, err := store.Get(keyB)
	require.NoError(t, err)
	require.Equal(t, []byte("b"), gotB)

	acc := util.Uint160{4}
	require.Error(t, fork.PutGasBalanceOverlay(overlay, acc, -1))
}
