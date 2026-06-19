package fork_test

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/core/storage"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/fork"
	"github.com/stretchr/testify/require"
)

func TestStoreSetRemoteAndClose(t *testing.T) {
	base := storage.NewMemoryStore()
	overlay := fork.NewTrackingOverlay()
	store := fork.NewStore(base, overlay)
	store.SetRemote(nil)
	require.NoError(t, store.Close())
}
