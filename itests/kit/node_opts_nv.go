package kit

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/network"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node"
	"github.com/ipfs/go-cid"
)

// DefaultTestUpgradeSchedule
var DefaultTestUpgradeSchedule = stmgr.UpgradeSchedule{{
	Network:   network.Version9,
	Height:    1,
	Migration: stmgr.UpgradeActorsV2,
}, {
	Network:   network.Version10,
	Height:    2,
	Migration: stmgr.UpgradeActorsV3,
}, {
	Network:   network.Version12,
	Height:    3,
	Migration: stmgr.UpgradeActorsV4,
}, {
	Network:   network.Version13,
	Height:    4,
	Migration: stmgr.UpgradeActorsV5,
}}

func LatestActorsAt(upgradeHeight abi.ChainEpoch) node.Option {
	return NetworkUpgradeAt(build.NewestNetworkVersion, upgradeHeight)
}

// InstantaneousNetworkVersion starts the network instantaneously at the
// specified version in height 1.
func InstantaneousNetworkVersion(version network.Version) node.Option {
	// composes all migration functions
	var mf stmgr.MigrationFunc = func(ctx context.Context, sm *stmgr.StateManager, cache stmgr.MigrationCache, cb stmgr.ExecMonitor, oldState cid.Cid, height abi.ChainEpoch, ts *types.TipSet) (newState cid.Cid, err error) {
		var state = oldState
		for _, u := range DefaultTestUpgradeSchedule {
			if u.Network > version {
				break
			}
			state, err = u.Migration(ctx, sm, cache, cb, state, height, ts)
			if err != nil {
				return cid.Undef, err
			}
		}
		return state, nil
	}
	return node.Override(new(stmgr.UpgradeSchedule), stmgr.UpgradeSchedule{
		{Network: version, Height: 1, Migration: mf},
	})
}

func NetworkUpgradeAt(version network.Version, upgradeHeight abi.ChainEpoch) node.Option {
	schedule := stmgr.UpgradeSchedule{}
	for _, upgrade := range DefaultTestUpgradeSchedule {
		if upgrade.Network > version {
			break
		}

		schedule = append(schedule, upgrade)
	}
	if len(schedule) == 0 {
		panic("empty upgrade schedule")
	}
	targetUpgrade := &schedule[len(schedule)-1]
	if targetUpgrade.Network != version {
		panic(fmt.Sprintf("failed to upgrade to target version %d, last version is %d",
			version, targetUpgrade.Network))
	}

	if upgradeHeight > 0 {
		// We can't go lower because our default upgrades are sequential.
		if upgradeHeight < targetUpgrade.Height {
			panic(fmt.Sprintf("target upgrade height %d for version %d less than minimum %d",
				upgradeHeight, version, targetUpgrade.Height))
		}
		targetUpgrade.Height = upgradeHeight
	}

	return node.Override(new(stmgr.UpgradeSchedule), schedule)
}

func SDRUpgradeAt(calico, persian abi.ChainEpoch) node.Option {
	return node.Override(new(stmgr.UpgradeSchedule), stmgr.UpgradeSchedule{{
		Network:   network.Version6,
		Height:    1,
		Migration: stmgr.UpgradeActorsV2,
	}, {
		Network:   network.Version7,
		Height:    calico,
		Migration: stmgr.UpgradeCalico,
	}, {
		Network: network.Version8,
		Height:  persian,
	}})
}
