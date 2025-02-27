// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package satellite

import (
	"context"
	"errors"
	"net"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"storj.io/common/identity"
	"storj.io/common/peertls/extensions"
	"storj.io/common/peertls/tlsopts"
	"storj.io/common/rpc"
	"storj.io/common/signing"
	"storj.io/common/storj"
	"storj.io/private/debug"
	"storj.io/private/version"
	"storj.io/storj/private/lifecycle"
	version_checker "storj.io/storj/private/version/checker"
	"storj.io/storj/satellite/metabase"
	"storj.io/storj/satellite/metainfo"
	"storj.io/storj/satellite/orders"
	"storj.io/storj/satellite/overlay"
	"storj.io/storj/satellite/repair/queue"
	"storj.io/storj/satellite/repair/repairer"
	"storj.io/storj/satellite/reputation"
)

// Repairer is the repairer process.
//
// architecture: Peer
type Repairer struct {
	Log      *zap.Logger
	Identity *identity.FullIdentity

	Servers  *lifecycle.Group
	Services *lifecycle.Group

	Dialer rpc.Dialer

	Version struct {
		Chore   *version_checker.Chore
		Service *version_checker.Service
	}

	Debug struct {
		Listener net.Listener
		Server   *debug.Server
	}

	Metainfo   *metainfo.Service
	Overlay    *overlay.Service
	Reputation *reputation.Service
	Orders     struct {
		DB      orders.DB
		Service *orders.Service
		Chore   *orders.Chore
	}
	SegmentRepairer *repairer.SegmentRepairer
	Repairer        *repairer.Service
}

// NewRepairer creates a new repairer peer.
func NewRepairer(log *zap.Logger, full *identity.FullIdentity,
	metabaseDB *metabase.DB,
	revocationDB extensions.RevocationDB, repairQueue queue.RepairQueue,
	bucketsDB metainfo.BucketsDB, overlayCache overlay.DB,
	reputationdb reputation.DB, rollupsWriteCache *orders.RollupsWriteCache,
	versionInfo version.Info, config *Config, atomicLogLevel *zap.AtomicLevel) (*Repairer, error) {
	peer := &Repairer{
		Log:      log,
		Identity: full,

		Servers:  lifecycle.NewGroup(log.Named("servers")),
		Services: lifecycle.NewGroup(log.Named("services")),
	}

	{ // setup debug
		var err error
		if config.Debug.Address != "" {
			peer.Debug.Listener, err = net.Listen("tcp", config.Debug.Address)
			if err != nil {
				withoutStack := errors.New(err.Error())
				peer.Log.Debug("failed to start debug endpoints", zap.Error(withoutStack))
			}
		}
		debugConfig := config.Debug
		debugConfig.ControlTitle = "Repair"
		peer.Debug.Server = debug.NewServerWithAtomicLevel(log.Named("debug"), peer.Debug.Listener, monkit.Default, debugConfig, atomicLogLevel)
		peer.Servers.Add(lifecycle.Item{
			Name:  "debug",
			Run:   peer.Debug.Server.Run,
			Close: peer.Debug.Server.Close,
		})
	}

	{
		peer.Log.Info("Version info",
			zap.Stringer("Version", versionInfo.Version.Version),
			zap.String("Commit Hash", versionInfo.CommitHash),
			zap.Stringer("Build Timestamp", versionInfo.Timestamp),
			zap.Bool("Release Build", versionInfo.Release),
		)
		peer.Version.Service = version_checker.NewService(log.Named("version"), config.Version, versionInfo, "Satellite")
		peer.Version.Chore = version_checker.NewChore(peer.Version.Service, config.Version.CheckInterval)

		peer.Services.Add(lifecycle.Item{
			Name: "version",
			Run:  peer.Version.Chore.Run,
		})
	}

	{ // setup dialer
		sc := config.Server

		tlsOptions, err := tlsopts.NewOptions(peer.Identity, sc.Config, revocationDB)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}

		peer.Dialer = rpc.NewDefaultDialer(tlsOptions)
	}

	{ // setup metainfo
		peer.Metainfo = metainfo.NewService(log.Named("metainfo"), bucketsDB, metabaseDB)
	}

	{ // setup overlay
		var err error
		peer.Overlay, err = overlay.NewService(log.Named("overlay"), overlayCache, config.Overlay)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}
		peer.Services.Add(lifecycle.Item{
			Name:  "overlay",
			Close: peer.Overlay.Close,
		})
	}

	{ // setup reputation
		peer.Reputation = reputation.NewService(log.Named("reputation:service"),
			overlayCache,
			reputationdb,
			config.Reputation,
		)

		peer.Services.Add(lifecycle.Item{
			Name:  "reputation",
			Close: peer.Reputation.Close,
		})
	}

	{ // setup orders
		peer.Orders.DB = rollupsWriteCache
		peer.Orders.Chore = orders.NewChore(log.Named("orders:chore"), rollupsWriteCache, config.Orders)
		peer.Services.Add(lifecycle.Item{
			Name:  "orders:chore",
			Run:   peer.Orders.Chore.Run,
			Close: peer.Orders.Chore.Close,
		})
		peer.Debug.Server.Panel.Add(
			debug.Cycle("Orders Chore", peer.Orders.Chore.Loop))

		var err error
		peer.Orders.Service, err = orders.NewService(
			log.Named("orders"),
			signing.SignerFromFullIdentity(peer.Identity),
			peer.Overlay,
			peer.Orders.DB,
			bucketsDB,
			config.Orders,
		)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}
	}

	{ // setup repairer
		peer.SegmentRepairer = repairer.NewSegmentRepairer(
			log.Named("segment-repair"),
			metabaseDB,
			peer.Orders.Service,
			peer.Overlay,
			peer.Reputation,
			peer.Dialer,
			config.Repairer.Timeout,
			config.Repairer.MaxExcessRateOptimalThreshold,
			config.Checker.RepairOverrides,
			config.Repairer.DownloadTimeout,
			config.Repairer.InMemoryRepair,
			signing.SigneeFromPeerIdentity(peer.Identity.PeerIdentity()),
		)
		peer.Repairer = repairer.NewService(log.Named("repairer"), repairQueue, &config.Repairer, peer.SegmentRepairer)

		peer.Services.Add(lifecycle.Item{
			Name:  "repair",
			Run:   peer.Repairer.Run,
			Close: peer.Repairer.Close,
		})
		peer.Debug.Server.Panel.Add(
			debug.Cycle("Repair Worker", peer.Repairer.Loop))
	}

	return peer, nil
}

// Run runs the repair process until it's either closed or it errors.
func (peer *Repairer) Run(ctx context.Context) (err error) {
	defer mon.Task()(&ctx)(&err)

	group, ctx := errgroup.WithContext(ctx)

	peer.Servers.Run(ctx, group)
	peer.Services.Run(ctx, group)

	return group.Wait()
}

// Close closes all the resources.
func (peer *Repairer) Close() error {
	return errs.Combine(
		peer.Servers.Close(),
		peer.Services.Close(),
	)
}

// ID returns the peer ID.
func (peer *Repairer) ID() storj.NodeID { return peer.Identity.ID }
