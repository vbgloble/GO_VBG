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

package accounts

import (
	"testing"
)

func TestURLParsing(t *testing.T) {
	url, err := parseURL("https://vbgloble.org")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "vbgloble.org" {
		t.Errorf("expected: %v, got: %v", "vbgloble.org", url.Path)
	}

	_, err = parseURL("vbgloble.org")
	if err == nil {
		t.Error("expected err, got: nil")
	}
}

func TestURLString(t *testing.T) {
	url := URL{Scheme: "https", Path: "vbgloble.org"}
	if url.String() != "https://vbgloble.org" {
		t.Errorf("expected: %v, got: %v", "https://vbgloble.org", url.String())
	}

	url = URL{Scheme: "", Path: "vbgloble.org"}
	if url.String() != "vbgloble.org" {
		t.Errorf("expected: %v, got: %v", "vbgloble.org", url.String())
	}
}

func TestURLMarshalJSON(t *testing.T) {
	url := URL{Scheme: "https", Path: "vbgloble.org"}
	json, err := url.MarshalJSON()
	if err != nil {
		t.Errorf("unexpcted error: %v", err)
	}
	if string(json) != "\"https://vbgloble.org\"" {
		t.Errorf("expected: %v, got: %v", "\"https://vbgloble.org\"", string(json))
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	url := &URL{}
	err := url.UnmarshalJSON([]byte("\"https://vbgloble.org\""))
	if err != nil {
		t.Errorf("unexpcted error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "vbgloble.org" {
		t.Errorf("expected: %v, got: %v", "https", url.Path)
	}
}

func TestURLComparison(t *testing.T) {
	tests := []struct {
		urlA   URL
		urlB   URL
		expect int
	}{
		{URL{"https", "vbgloble.org"}, URL{"https", "vbgloble.org"}, 0},
		{URL{"http", "vbgloble.org"}, URL{"https", "vbgloble.org"}, -1},
		{URL{"https", "vbgloble.org/a"}, URL{"https", "vbgloble.org"}, 1},
		{URL{"https", "abc.org"}, URL{"https", "vbgloble.org"}, -1},
	}

	for i, tt := range tests {
		result := tt.urlA.Cmp(tt.urlB)
		if result != tt.expect {
			t.Errorf("test %d: cmp mismatch: expected: %d, got: %d", i, tt.expect, result)
		}
	}
}
