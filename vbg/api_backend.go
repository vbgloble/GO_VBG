// Copyright 2015 The go-VGB Authors
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

package VBG

import (
	"context"
	"errors"
	"math/big"

	"github.com/vbgloble/go-VGB/accounts"
	"github.com/vbgloble/go-VGB/common"
	"github.com/vbgloble/go-VGB/consensus"
	"github.com/vbgloble/go-VGB/core"
	"github.com/vbgloble/go-VGB/core/bloombits"
	"github.com/vbgloble/go-VGB/core/rawdb"
	"github.com/vbgloble/go-VGB/core/state"
	"github.com/vbgloble/go-VGB/core/types"
	"github.com/vbgloble/go-VGB/core/vm"
	"github.com/vbgloble/go-VGB/VBG/downloader"
	"github.com/vbgloble/go-VGB/VBG/gasprice"
	"github.com/vbgloble/go-VGB/VBGdb"
	"github.com/vbgloble/go-VGB/event"
	"github.com/vbgloble/go-VGB/miner"
	"github.com/vbgloble/go-VGB/params"
	"github.com/vbgloble/go-VGB/rpc"
)

// VBGAPIBackend implements VBGapi.Backend for full nodes
type VBGAPIBackend struct {
	extRPCEnabled bool
	VBG           *vbgloble
	gpo           *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *VBGAPIBackend) ChainConfig() *params.ChainConfig {
	return b.VBG.blockchain.Config()
}

func (b *VBGAPIBackend) CurrentBlock() *types.Block {
	return b.VBG.blockchain.CurrentBlock()
}

func (b *VBGAPIBackend) SVBGead(number uint64) {
	b.VBG.protocolManager.downloader.Cancel()
	b.VBG.blockchain.SVBGead(number)
}

func (b *VBGAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.VBG.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.VBG.blockchain.CurrentBlock().Header(), nil
	}
	return b.VBG.blockchain.GVBGeaderByNumber(uint64(number)), nil
}

func (b *VBGAPIBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.HeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.VBG.blockchain.GVBGeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.VBG.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		return header, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *VBGAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.VBG.blockchain.GVBGeaderByHash(hash), nil
}

func (b *VBGAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block := b.VBG.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.VBG.blockchain.CurrentBlock(), nil
	}
	return b.VBG.blockchain.GetBlockByNumber(uint64(number)), nil
}

func (b *VBGAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.VBG.blockchain.GetBlockByHash(hash), nil
}

func (b *VBGAPIBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.BlockByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.VBG.blockchain.GVBGeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.VBG.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		block := b.VBG.blockchain.GetBlock(hash, header.Number.Uint64())
		if block == nil {
			return nil, errors.New("header found, but block body is missing")
		}
		return block, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *VBGAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, state := b.VBG.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, nil, err
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb, err := b.VBG.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *VBGAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.StateAndHeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header, err := b.HeaderByHash(ctx, hash)
		if err != nil {
			return nil, nil, err
		}
		if header == nil {
			return nil, nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.VBG.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, nil, errors.New("hash is not currently canonical")
		}
		stateDb, err := b.VBG.BlockChain().StateAt(header.Root)
		return stateDb, header, err
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *VBGAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.VBG.blockchain.GetReceiptsByHash(hash), nil
}

func (b *VBGAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	receipts := b.VBG.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *VBGAPIBackend) GetTd(ctx context.Context, hash common.Hash) *big.Int {
	return b.VBG.blockchain.GetTdByHash(hash)
}

func (b *VBGAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error) {
	vmError := func() error { return nil }

	txContext := core.NewEVMTxContext(msg)
	context := core.NewEVMBlockContext(header, b.VBG.BlockChain(), nil)
	return vm.NewEVM(context, txContext, state, b.VBG.blockchain.Config(), *b.VBG.blockchain.GetVMConfig()), vmError, nil
}

func (b *VBGAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.VBG.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *VBGAPIBackend) SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.VBG.miner.SubscribePendingLogs(ch)
}

func (b *VBGAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.VBG.BlockChain().SubscribeChainEvent(ch)
}

func (b *VBGAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.VBG.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *VBGAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.VBG.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *VBGAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.VBG.BlockChain().SubscribeLogsEvent(ch)
}

func (b *VBGAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.VBG.txPool.AddLocal(signedTx)
}

func (b *VBGAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.VBG.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *VBGAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.VBG.txPool.Get(hash)
}

func (b *VBGAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(b.VBG.ChainDb(), txHash)
	return tx, blockHash, blockNumber, index, nil
}

func (b *VBGAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.VBG.txPool.Nonce(addr), nil
}

func (b *VBGAPIBackend) Stats() (pending int, queued int) {
	return b.VBG.txPool.Stats()
}

func (b *VBGAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.VBG.TxPool().Content()
}

func (b *VBGAPIBackend) TxPool() *core.TxPool {
	return b.VBG.TxPool()
}

func (b *VBGAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.VBG.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *VBGAPIBackend) Downloader() *downloader.Downloader {
	return b.VBG.Downloader()
}

func (b *VBGAPIBackend) ProtocolVersion() int {
	return b.VBG.VBGVersion()
}

func (b *VBGAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *VBGAPIBackend) ChainDb() VBGdb.Database {
	return b.VBG.ChainDb()
}

func (b *VBGAPIBackend) EventMux() *event.TypeMux {
	return b.VBG.EventMux()
}

func (b *VBGAPIBackend) AccountManager() *accounts.Manager {
	return b.VBG.AccountManager()
}

func (b *VBGAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *VBGAPIBackend) RPCGasCap() uint64 {
	return b.VBG.config.RPCGasCap
}

func (b *VBGAPIBackend) RPCTxFeeCap() float64 {
	return b.VBG.config.RPCTxFeeCap
}

func (b *VBGAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.VBG.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *VBGAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.VBG.bloomRequests)
	}
}

func (b *VBGAPIBackend) Engine() consensus.Engine {
	return b.VBG.engine
}

func (b *VBGAPIBackend) CurrentHeader() *types.Header {
	return b.VBG.blockchain.CurrentHeader()
}

func (b *VBGAPIBackend) Miner() *miner.Miner {
	return b.VBG.Miner()
}

func (b *VBGAPIBackend) StartMining(threads int) error {
	return b.VBG.StartMining(threads)
}
