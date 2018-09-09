package inject

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type moduleForCombine1 struct{}
type moduleForCombine2 struct{}

func TestCombineModules(t *testing.T) {
	var combined Module = CombineModules(moduleForCombine1{}, moduleForCombine2{})
	require.NotNil(t, combined)
	_, ok := combined.(moduleCollection)
	require.True(t, ok)
}
