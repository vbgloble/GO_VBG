// Copyright 2016 The go-VGB Authors
// This file is part of the go-VGB library.
//
// The go-VGB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-VGB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-VGB library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light vbgloble Subprotocol.
package les

import (
	"fmt"
	"time"

	"github.com/vbgloble/go-VGB/accounts"
	"github.com/vbgloble/go-VGB/common"
	"github.com/vbgloble/go-VGB/common/hexutil"
	"github.com/vbgloble/go-VGB/common/mclock"
	"github.com/vbgloble/go-VGB/consensus"
	"github.com/vbgloble/go-VGB/core"
	"github.com/vbgloble/go-VGB/core/bloombits"
	"github.com/vbgloble/go-VGB/core/rawdb"
	"github.com/vbgloble/go-VGB/core/types"
	"github.com/vbgloble/go-VGB/VBG"
	"github.com/vbgloble/go-VGB/VBG/downloader"
	"github.com/vbgloble/go-VGB/VBG/filters"
	"github.com/vbgloble/go-VGB/VBG/gasprice"
	"github.com/vbgloble/go-VGB/event"
	"github.com/vbgloble/go-VGB/internal/VBGapi"
	lpc "github.com/vbgloble/go-VGB/les/lespay/client"
	"github.com/vbgloble/go-VGB/light"
	"github.com/vbgloble/go-VGB/log"
	"github.com/vbgloble/go-VGB/node"
	"github.com/vbgloble/go-VGB/p2p"
	"github.com/vbgloble/go-VGB/p2p/enode"
	"github.com/vbgloble/go-VGB/params"
	"github.com/vbgloble/go-VGB/rpc"
)

type Lightvbgloble struct {
	lesCommons

	peers          *serverPeerSet
	reqDist        *requestDistributor
	retriever      *retrieveManager
	odr            *LesOdr
	relay          *lesTxRelay
	handler        *clientHandler
	txPool         *light.TxPool
	blockchain     *light.LightChain
	serverPool     *serverPool
	valueTracker   *lpc.ValueTracker
	dialCandidates enode.Iterator
	pruner         *pruner

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend     *LesApiBackend
	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager
	netRPCService  *VBGapi.PublicNetAPI

	p2pServer *p2p.Server
}

// New creates an instance of the light client.
func New(stack *node.Node, config *VBG.Config) (*Lightvbgloble, error) {
	chainDb, err := stack.OpenDatabase("lightchaindata", config.DatabaseCache, config.DatabaseHandles, "VBG/db/chaindata/")
	if err != nil {
		return nil, err
	}
	lespayDb, err := stack.OpenDatabase("lespay", 0, 0, "VBG/db/lespay")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newServerPeerSet()
	lVBG := &Lightvbgloble{
		lesCommons: lesCommons{
			genesis:     genesisHash,
			config:      config,
			chainConfig: chainConfig,
			iConfig:     light.DefaultClientIndexerConfig,
			chainDb:     chainDb,
			closeCh:     make(chan struct{}),
		},
		peers:          peers,
		eventMux:       stack.EventMux(),
		reqDist:        newRequestDistributor(peers, &mclock.System{}),
		accountManager: stack.AccountManager(),
		engine:         VBG.CreateConsensusEngine(stack, chainConfig, &config.VBGash, nil, false, chainDb),
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   VBG.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
		valueTracker:   lpc.NewValueTracker(lespayDb, &mclock.System{}, requestList, time.Minute, 1/float64(time.Hour), 1/float64(time.Hour*100), 1/float64(time.Hour*1000)),
		p2pServer:      stack.Server(),
	}
	peers.subscribe((*vtSubscription)(lVBG.valueTracker))

	dnsdisc, err := lVBG.setupDiscovery(&stack.Config().P2P)
	if err != nil {
		return nil, err
	}
	lVBG.serverPool = newServerPool(lespayDb, []byte("serverpool:"), lVBG.valueTracker, dnsdisc, time.Second, nil, &mclock.System{}, config.UltraLightServers)
	peers.subscribe(lVBG.serverPool)
	lVBG.dialCandidates = lVBG.serverPool.dialIterator

	lVBG.retriever = newRetrieveManager(peers, lVBG.reqDist, lVBG.serverPool.getTimeout)
	lVBG.relay = newLesTxRelay(peers, lVBG.retriever)

	lVBG.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, lVBG.retriever)
	lVBG.chtIndexer = light.NewChtIndexer(chainDb, lVBG.odr, params.CHTFrequency, params.HelperTrieConfirmations, config.LightNoPrune)
	lVBG.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, lVBG.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency, config.LightNoPrune)
	lVBG.odr.SetIndexers(lVBG.chtIndexer, lVBG.bloomTrieIndexer, lVBG.bloomIndexer)

	checkpoint := config.Checkpoint
	if checkpoint == nil {
		checkpoint = params.TrustedCheckpoints[genesisHash]
	}
	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if lVBG.blockchain, err = light.NewLightChain(lVBG.odr, lVBG.chainConfig, lVBG.engine, checkpoint); err != nil {
		return nil, err
	}
	lVBG.chainReader = lVBG.blockchain
	lVBG.txPool = light.NewTxPool(lVBG.chainConfig, lVBG.blockchain, lVBG.relay)

	// Set up checkpoint oracle.
	lVBG.oracle = lVBG.setupOracle(stack, genesisHash, config)

	// Note: AddChildIndexer starts the update process for the child
	lVBG.bloomIndexer.AddChildIndexer(lVBG.bloomTrieIndexer)
	lVBG.chtIndexer.Start(lVBG.blockchain)
	lVBG.bloomIndexer.Start(lVBG.blockchain)

	// Start a light chain pruner to delete useless historical data.
	lVBG.pruner = newPruner(chainDb, lVBG.chtIndexer, lVBG.bloomTrieIndexer)

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lVBG.blockchain.SVBGead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lVBG.ApiBackend = &LesApiBackend{stack.Config().ExtRPCEnabled(), lVBG, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.GasPrice
	}
	lVBG.ApiBackend.gpo = gasprice.NewOracle(lVBG.ApiBackend, gpoParams)

	lVBG.handler = newClientHandler(config.UltraLightServers, config.UltraLightFraction, checkpoint, lVBG)
	if lVBG.handler.ulc != nil {
		log.Warn("Ultra light client is enabled", "trustedNodes", len(lVBG.handler.ulc.keys), "minTrustedFraction", lVBG.handler.ulc.fraction)
		lVBG.blockchain.DisableCheckFreq()
	}

	lVBG.netRPCService = VBGapi.NewPublicNetAPI(lVBG.p2pServer, lVBG.config.NetworkId)

	// Register the backend on the node
	stack.RegisterAPIs(lVBG.APIs())
	stack.RegisterProtocols(lVBG.Protocols())
	stack.RegisterLifecycle(lVBG)

	return lVBG, nil
}

// vtSubscription implements serverPeerSubscriber
type vtSubscription lpc.ValueTracker

// registerPeer implements serverPeerSubscriber
func (v *vtSubscription) registerPeer(p *serverPeer) {
	vt := (*lpc.ValueTracker)(v)
	p.setValueTracker(vt, vt.Register(p.ID()))
	p.updateVtParams()
}

// unregisterPeer implements serverPeerSubscriber
func (v *vtSubscription) unregisterPeer(p *serverPeer) {
	vt := (*lpc.ValueTracker)(v)
	vt.Unregister(p.ID())
	p.setValueTracker(nil, nil)
}

type LightDummyAPI struct{}

// VBGerbase is the address that mining rewards will be send to
func (s *LightDummyAPI) VBGerbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Coinbase is the address that mining rewards will be send to (alias for VBGerbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("mining is not supported in light mode")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the vbgloble package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Lightvbgloble) APIs() []rpc.API {
	apis := VBGapi.GetAPIs(s.ApiBackend)
	apis = append(apis, s.engine.APIs(s.BlockChain().HeaderChain())...)
	return append(apis, []rpc.API{
		{
			Namespace: "VBG",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "VBG",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.handler.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "VBG",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		}, {
			Namespace: "les",
			Version:   "1.0",
			Service:   NewPrivateLightAPI(&s.lesCommons),
			Public:    false,
		}, {
			Namespace: "lespay",
			Version:   "1.0",
			Service:   lpc.NewPrivateClientAPI(s.valueTracker),
			Public:    false,
		},
	}...)
}

func (s *Lightvbgloble) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Lightvbgloble) BlockChain() *light.LightChain      { return s.blockchain }
func (s *Lightvbgloble) TxPool() *light.TxPool              { return s.txPool }
func (s *Lightvbgloble) Engine() consensus.Engine           { return s.engine }
func (s *Lightvbgloble) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *Lightvbgloble) Downloader() *downloader.Downloader { return s.handler.downloader }
func (s *Lightvbgloble) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols returns all the currently configured network protocols to start.
func (s *Lightvbgloble) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions, s.handler.runPeer, func(id enode.ID) interface{} {
		if p := s.peers.peer(id.String()); p != nil {
			return p.Info()
		}
		return nil
	}, s.dialCandidates)
}

// Start implements node.Lifecycle, starting all internal goroutines needed by the
// light vbgloble protocol implementation.
func (s *Lightvbgloble) Start() error {
	log.Warn("Light client mode is an experimental feature")

	s.serverPool.start()
	// Start bloom request workers.
	s.wg.Add(bloomServicVBGreads)
	s.startBloomHandlers(params.BloomBitsBlocksClient)
	s.handler.start()

	return nil
}

// Stop implements node.Lifecycle, terminating all internal goroutines used by the
// vbgloble protocol.
func (s *Lightvbgloble) Stop() error {
	close(s.closeCh)
	s.serverPool.stop()
	s.valueTracker.Stop()
	s.peers.close()
	s.reqDist.close()
	s.odr.Stop()
	s.relay.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.handler.stop()
	s.txPool.Stop()
	s.engine.Close()
	s.pruner.close()
	s.eventMux.Stop()
	s.chainDb.Close()
	s.wg.Wait()
	log.Info("Light vbgloble stopped")
	return nil
}
