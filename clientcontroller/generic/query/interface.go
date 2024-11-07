package query

import "github.com/babylonlabs-io/finality-gadget/types"

type IQuery interface {
	QueryIsBlockBabylonFinalized(block *types.Block) (bool, error)
	QueryBtcStakingActivatedTimestamp() (uint64, error)
}
