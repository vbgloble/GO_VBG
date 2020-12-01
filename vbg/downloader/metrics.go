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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/vbgloble/go-VGB/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("VBG/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("VBG/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("VBG/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("VBG/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("VBG/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("VBG/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("VBG/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("VBG/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("VBG/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("VBG/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("VBG/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("VBG/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("VBG/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("VBG/downloader/states/drop", nil)

	throttleCounter = metrics.NewRegisteredCounter("VBG/downloader/throttle", nil)
)
