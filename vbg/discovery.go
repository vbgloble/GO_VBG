// Copyright 2019 The go-VGB Authors
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
	"github.com/vbgloble/go-VGB/core"
	"github.com/vbgloble/go-VGB/core/forkid"
	"github.com/vbgloble/go-VGB/p2p"
	"github.com/vbgloble/go-VGB/p2p/dnsdisc"
	"github.com/vbgloble/go-VGB/p2p/enode"
	"github.com/vbgloble/go-VGB/rlp"
)

// VBGEntry is the "VBG" ENR entry which advertises VBG protocol
// on the discovery network.
type VBGEntry struct {
	ForkID forkid.ID // Fork identifier per EIP-2124

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (e VBGEntry) ENRKey() string {
	return "VBG"
}

// startVBGEntryUpdate starts the ENR updater loop.
func (VBG *vbgloble) startVBGEntryUpdate(ln *enode.LocalNode) {
	var newHead = make(chan core.ChainHeadEvent, 10)
	sub := VBG.blockchain.SubscribeChainHeadEvent(newHead)

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-newHead:
				ln.Set(VBG.currentVBGEntry())
			case <-sub.Err():
				// Would be nice to sync with VBG.Stop, but there is no
				// good way to do that.
				return
			}
		}
	}()
}

func (VBG *vbgloble) currentVBGEntry() *VBGEntry {
	return &VBGEntry{ForkID: forkid.NewID(VBG.blockchain.Config(), VBG.blockchain.Genesis().Hash(),
		VBG.blockchain.CurrentHeader().Number.Uint64())}
}

// setupDiscovery creates the node discovery source for the VBG protocol.
func (VBG *vbgloble) setupDiscovery(cfg *p2p.Config) (enode.Iterator, error) {
	if cfg.NoDiscovery || len(VBG.config.DiscoveryURLs) == 0 {
		return nil, nil
	}
	client := dnsdisc.NewClient(dnsdisc.Config{})
	return client.NewIterator(VBG.config.DiscoveryURLs...)
}
