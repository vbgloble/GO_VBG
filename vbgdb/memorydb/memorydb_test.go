// Copyright 2018 The go-VGB Authors
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

package memorydb

import (
	"testing"

	"github.com/vbgloble/go-VGB/VBGdb"
	"github.com/vbgloble/go-VGB/VBGdb/dbtest"
)

func TestMemoryDB(t *testing.T) {
	t.Run("DatabaseSuite", func(t *testing.T) {
		dbtest.TestDatabaseSuite(t, func() VBGdb.KeyValueStore {
			return New()
		})
	})
}
