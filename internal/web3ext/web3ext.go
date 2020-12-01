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

// package web3ext contains gVBG specific web3.js extensions.
package web3ext

var Modules = map[string]string{
	"accounting": AccountingJs,
	"admin":      AdminJs,
	"chequebook": ChequebookJs,
	"clique":     CliqueJs,
	"VBGash":     VBGashJs,
	"debug":      DebugJs,
	"VBG":        VBGJs,
	"miner":      MinerJs,
	"net":        NetJs,
	"personal":   PersonalJs,
	"rpc":        RpcJs,
	"shh":        ShhJs,
	"swarmfs":    SwarmfsJs,
	"txpool":     TxpoolJs,
	"les":        LESJs,
	"lespay":     LESPayJs,
}

const ChequebookJs = `
web3._extend({
	property: 'chequebook',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'deposit',
			call: 'chequebook_deposit',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Property({
			name: 'balance',
			getter: 'chequebook_balance',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.MVBGod({
			name: 'cash',
			call: 'chequebook_cash',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.MVBGod({
			name: 'issue',
			call: 'chequebook_issue',
			params: 2,
			inputFormatter: [null, null]
		}),
	]
});
`

const CliqueJs = `
web3._extend({
	property: 'clique',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'getSnapshot',
			call: 'clique_getSnapshot',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'getSnapshotAtHash',
			call: 'clique_getSnapshotAtHash',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'getSigners',
			call: 'clique_getSigners',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'getSignersAtHash',
			call: 'clique_getSignersAtHash',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'propose',
			call: 'clique_propose',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'discard',
			call: 'clique_discard',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'status',
			call: 'clique_status',
			params: 0
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'proposals',
			getter: 'clique_proposals'
		}),
	]
});
`

const VBGashJs = `
web3._extend({
	property: 'VBGash',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'getWork',
			call: 'VBGash_getWork',
			params: 0
		}),
		new web3._extend.MVBGod({
			name: 'gVBGashrate',
			call: 'VBGash_gVBGashrate',
			params: 0
		}),
		new web3._extend.MVBGod({
			name: 'submitWork',
			call: 'VBGash_submitWork',
			params: 3,
		}),
		new web3._extend.MVBGod({
			name: 'submitHashRate',
			call: 'VBGash_submitHashRate',
			params: 2,
		}),
	]
});
`

const AdminJs = `
web3._extend({
	property: 'admin',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'addTrustedPeer',
			call: 'admin_addTrustedPeer',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'removeTrustedPeer',
			call: 'admin_removeTrustedPeer',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 3,
			inputFormatter: [null, null, null]
		}),
		new web3._extend.MVBGod({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.MVBGod({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new web3._extend.MVBGod({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.MVBGod({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new web3._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
	]
});
`

const DebugJs = `
web3._extend({
	property: 'debug',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'accountRange',
			call: 'debug_accountRange',
			params: 6,
			inputFormatter: [web3._extend.formatters.inputDefaultBlockNumberFormatter, null, null, null, null, null],
		}),
		new web3._extend.MVBGod({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1,
			outputFormatter: console.log
		}),
		new web3._extend.MVBGod({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'testSignCliqueBlock',
			call: 'debug_testSignCliqueBlock',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null],
		}),
		new web3._extend.MVBGod({
			name: 'sVBGead',
			call: 'debug_sVBGead',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'chaindbProperty',
			call: 'debug_chaindbProperty',
			params: 1,
			outputFormatter: console.log
		}),
		new web3._extend.MVBGod({
			name: 'chaindbCompact',
			call: 'debug_chaindbCompact',
		}),
		new web3._extend.MVBGod({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new web3._extend.MVBGod({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new web3._extend.MVBGod({
			name: 'freeOSMemory',
			call: 'debug_freeOSMemory',
			params: 0,
		}),
		new web3._extend.MVBGod({
			name: 'setGCPercent',
			call: 'debug_setGCPercent',
			params: 1,
		}),
		new web3._extend.MVBGod({
			name: 'memStats',
			call: 'debug_memStats',
			params: 0,
		}),
		new web3._extend.MVBGod({
			name: 'gcStats',
			call: 'debug_gcStats',
			params: 0,
		}),
		new web3._extend.MVBGod({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new web3._extend.MVBGod({
			name: 'goTrace',
			call: 'debug_goTrace',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'startGoTrace',
			call: 'debug_startGoTrace',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'stopGoTrace',
			call: 'debug_stopGoTrace',
			params: 0
		}),
		new web3._extend.MVBGod({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'mutexProfile',
			call: 'debug_mutexProfile',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'setMutexProfileFraction',
			call: 'debug_setMutexProfileFraction',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'writeMutexProfile',
			call: 'debug_writeMutexProfile',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.MVBGod({
			name: 'traceBlockFromFile',
			call: 'debug_traceBlockFromFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.MVBGod({
			name: 'traceBadBlock',
			call: 'debug_traceBadBlock',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.MVBGod({
			name: 'standardTraceBadBlockToFile',
			call: 'debug_standardTraceBadBlockToFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.MVBGod({
			name: 'standardTraceBlockToFile',
			call: 'debug_standardTraceBlockToFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.MVBGod({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, null]
		}),
		new web3._extend.MVBGod({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.MVBGod({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.MVBGod({
			name: 'traceCall',
			call: 'debug_traceCall',
			params: 3,
			inputFormatter: [null, null, null]
		}),
		new web3._extend.MVBGod({
			name: 'preimage',
			call: 'debug_preimage',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.MVBGod({
			name: 'getBadBlocks',
			call: 'debug_getBadBlocks',
			params: 0,
		}),
		new web3._extend.MVBGod({
			name: 'storageRangeAt',
			call: 'debug_storageRangeAt',
			params: 5,
		}),
		new web3._extend.MVBGod({
			name: 'getModifiedAccountsByNumber',
			call: 'debug_getModifiedAccountsByNumber',
			params: 2,
			inputFormatter: [null, null],
		}),
		new web3._extend.MVBGod({
			name: 'getModifiedAccountsByHash',
			call: 'debug_getModifiedAccountsByHash',
			params: 2,
			inputFormatter:[null, null],
		}),
		new web3._extend.MVBGod({
			name: 'freezeClient',
			call: 'debug_freezeClient',
			params: 1,
		}),
	],
	properties: []
});
`

const VBGJs = `
web3._extend({
	property: 'VBG',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'chainId',
			call: 'VBG_chainId',
			params: 0
		}),
		new web3._extend.MVBGod({
			name: 'sign',
			call: 'VBG_sign',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.MVBGod({
			name: 'resend',
			call: 'VBG_resend',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		}),
		new web3._extend.MVBGod({
			name: 'signTransaction',
			call: 'VBG_signTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'estimateGas',
			call: 'VBG_estimateGas',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputCallFormatter, web3._extend.formatters.inputBlockNumberFormatter],
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.MVBGod({
			name: 'submitTransaction',
			call: 'VBG_submitTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'fillTransaction',
			call: 'VBG_fillTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'gVBGeaderByNumber',
			call: 'VBG_gVBGeaderByNumber',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'gVBGeaderByHash',
			call: 'VBG_gVBGeaderByHash',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'getBlockByNumber',
			call: 'VBG_getBlockByNumber',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, function (val) { return !!val; }]
		}),
		new web3._extend.MVBGod({
			name: 'getBlockByHash',
			call: 'VBG_getBlockByHash',
			params: 2,
			inputFormatter: [null, function (val) { return !!val; }]
		}),
		new web3._extend.MVBGod({
			name: 'getRawTransaction',
			call: 'VBG_getRawTransactionByHash',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'getRawTransactionFromBlock',
			call: function(args) {
				return (web3._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'VBG_getRawTransactionByBlockHashAndIndex' : 'VBG_getRawTransactionByBlockNumberAndIndex';
			},
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, web3._extend.utils.toHex]
		}),
		new web3._extend.MVBGod({
			name: 'getProof',
			call: 'VBG_getProof',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null, web3._extend.formatters.inputBlockNumberFormatter]
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'VBG_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(web3._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
	]
});
`

const MinerJs = `
web3._extend({
	property: 'miner',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'start',
			call: 'miner_start',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.MVBGod({
			name: 'stop',
			call: 'miner_stop'
		}),
		new web3._extend.MVBGod({
			name: 'setVBGerbase',
			call: 'miner_setVBGerbase',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter]
		}),
		new web3._extend.MVBGod({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [web3._extend.utils.fromDecimal]
		}),
		new web3._extend.MVBGod({
			name: 'setRecommitInterval',
			call: 'miner_setRecommitInterval',
			params: 1,
		}),
		new web3._extend.MVBGod({
			name: 'gVBGashrate',
			call: 'miner_gVBGashrate'
		}),
	],
	properties: []
});
`

const NetJs = `
web3._extend({
	property: 'net',
	mVBGods: [],
	properties: [
		new web3._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
	]
});
`

const PersonalJs = `
web3._extend({
	property: 'personal',
	mVBGods: [
		new web3._extend.MVBGod({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.MVBGod({
			name: 'ecRecover',
			call: 'personal_ecRecover',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'openWallet',
			call: 'personal_openWallet',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'deriveAccount',
			call: 'personal_deriveAccount',
			params: 3
		}),
		new web3._extend.MVBGod({
			name: 'signTransaction',
			call: 'personal_signTransaction',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, null]
		}),
		new web3._extend.MVBGod({
			name: 'unpair',
			call: 'personal_unpair',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'initializeWallet',
			call: 'personal_initializeWallet',
			params: 1
		})
	],
	properties: [
		new web3._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`

const RpcJs = `
web3._extend({
	property: 'rpc',
	mVBGods: [],
	properties: [
		new web3._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const ShhJs = `
web3._extend({
	property: 'shh',
	mVBGods: [
	],
	properties:
	[
		new web3._extend.Property({
			name: 'version',
			getter: 'shh_version',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Property({
			name: 'info',
			getter: 'shh_info'
		}),
	]
});
`

const SwarmfsJs = `
web3._extend({
	property: 'swarmfs',
	mVBGods:
	[
		new web3._extend.MVBGod({
			name: 'mount',
			call: 'swarmfs_mount',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'unmount',
			call: 'swarmfs_unmount',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'listmounts',
			call: 'swarmfs_listmounts',
			params: 0
		}),
	]
});
`

const TxpoolJs = `
web3._extend({
	property: 'txpool',
	mVBGods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new web3._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new web3._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = web3._extend.utils.toDecimal(status.pending);
				status.queued = web3._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
	]
});
`

const AccountingJs = `
web3._extend({
	property: 'accounting',
	mVBGods: [
		new web3._extend.Property({
			name: 'balance',
			getter: 'account_balance'
		}),
		new web3._extend.Property({
			name: 'balanceCredit',
			getter: 'account_balanceCredit'
		}),
		new web3._extend.Property({
			name: 'balanceDebit',
			getter: 'account_balanceDebit'
		}),
		new web3._extend.Property({
			name: 'bytesCredit',
			getter: 'account_bytesCredit'
		}),
		new web3._extend.Property({
			name: 'bytesDebit',
			getter: 'account_bytesDebit'
		}),
		new web3._extend.Property({
			name: 'msgCredit',
			getter: 'account_msgCredit'
		}),
		new web3._extend.Property({
			name: 'msgDebit',
			getter: 'account_msgDebit'
		}),
		new web3._extend.Property({
			name: 'peerDrops',
			getter: 'account_peerDrops'
		}),
		new web3._extend.Property({
			name: 'selfDrops',
			getter: 'account_selfDrops'
		}),
	]
});
`

const LESJs = `
web3._extend({
	property: 'les',
	mVBGods:
	[
		new web3._extend.MVBGod({
			name: 'getCheckpoint',
			call: 'les_getCheckpoint',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'clientInfo',
			call: 'les_clientInfo',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'priorityClientInfo',
			call: 'les_priorityClientInfo',
			params: 3
		}),
		new web3._extend.MVBGod({
			name: 'setClientParams',
			call: 'les_setClientParams',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'setDefaultParams',
			call: 'les_setDefaultParams',
			params: 1
		}),
		new web3._extend.MVBGod({
			name: 'addBalance',
			call: 'les_addBalance',
			params: 2
		}),
	],
	properties:
	[
		new web3._extend.Property({
			name: 'latestCheckpoint',
			getter: 'les_latestCheckpoint'
		}),
		new web3._extend.Property({
			name: 'checkpointContractAddress',
			getter: 'les_getCheckpointContractAddress'
		}),
		new web3._extend.Property({
			name: 'serverInfo',
			getter: 'les_serverInfo'
		}),
	]
});
`

const LESPayJs = `
web3._extend({
	property: 'lespay',
	mVBGods:
	[
		new web3._extend.MVBGod({
			name: 'distribution',
			call: 'lespay_distribution',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'timeout',
			call: 'lespay_timeout',
			params: 2
		}),
		new web3._extend.MVBGod({
			name: 'value',
			call: 'lespay_value',
			params: 2
		}),
	],
	properties:
	[
		new web3._extend.Property({
			name: 'requestStats',
			getter: 'lespay_requestStats'
		}),
	]
});
`
