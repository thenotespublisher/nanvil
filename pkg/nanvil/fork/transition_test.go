package fork_test

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/accounts"
	"github.com/nspcc-dev/neo-go/pkg/nanvil/fork"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestBuildAndApplyTransition(t *testing.T) {
	pubHex, err := accounts.ValidatorPublicKeyHex()
	require.NoError(t, err)
	pub, err := keys.NewPublicKeyFromString(pubHex)
	require.NoError(t, err)

	m := &fork.Manifest{
		Index:     42,
		IndexHash: util.Uint256{9},
	}
	info := fork.BuildTransition(m, pub)
	require.Equal(t, uint32(42), info.BranchIndex)
	require.Equal(t, m.IndexHash, info.BranchHash)
	require.Equal(t, pub.GetScriptHash(), info.LocalValidator)
	require.Equal(t, uint32(43), info.TransitionHeight)

	info2 := fork.ApplyTransition(m, pub)
	require.Equal(t, info, info2)
}

func TestBranchState(t *testing.T) {
	m := &fork.Manifest{Index: 7}
	bs := fork.NewBranchState(m)
	require.Equal(t, uint32(7), bs.Height())

	empty := fork.NewBranchState(nil)
	require.Equal(t, uint32(0), empty.Height())
}
