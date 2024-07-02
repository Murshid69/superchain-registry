package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

func testDataAvailability(t *testing.T, chain *ChainConfig) {
	require.Nil(t, chain.Plasma, "Standard chains use Ethereum L1 calldata or blobs for data availability (plasma not permitted)")
}
