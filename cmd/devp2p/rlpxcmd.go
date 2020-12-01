// Copyright 2020 The go-VGB Authors
// This file is part of go-VGB.
//
// go-VGB is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-VGB is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-VGB. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"net"

	"github.com/vbgloble/go-VGB/cmd/devp2p/internal/VBGtest"
	"github.com/vbgloble/go-VGB/crypto"
	"github.com/vbgloble/go-VGB/p2p"
	"github.com/vbgloble/go-VGB/p2p/rlpx"
	"github.com/vbgloble/go-VGB/rlp"
	"gopkg.in/urfave/cli.v1"
)

var (
	rlpxCommand = cli.Command{
		Name:  "rlpx",
		Usage: "RLPx Commands",
		Subcommands: []cli.Command{
			rlpxPingCommand,
			rlpxVBGTestCommand,
		},
	}
	rlpxPingCommand = cli.Command{
		Name:   "ping",
		Usage:  "ping <node>",
		Action: rlpxPing,
	}
	rlpxVBGTestCommand = cli.Command{
		Name:      "VBG-test",
		Usage:     "Runs tests against a node",
		ArgsUsage: "<node> <chain.rlp> <genesis.json>",
		Action:    rlpxVBGTest,
		Flags: []cli.Flag{
			testPatternFlag,
			testTAPFlag,
		},
	}
)

func rlpxPing(ctx *cli.Context) error {
	n := getNodeArg(ctx)
	fd, err := net.Dial("tcp", fmt.Sprintf("%v:%d", n.IP(), n.TCP()))
	if err != nil {
		return err
	}
	conn := rlpx.NewConn(fd, n.Pubkey())
	ourKey, _ := crypto.GenerateKey()
	_, err = conn.Handshake(ourKey)
	if err != nil {
		return err
	}
	code, data, _, err := conn.Read()
	if err != nil {
		return err
	}
	switch code {
	case 0:
		var h VBGtest.Hello
		if err := rlp.DecodeBytes(data, &h); err != nil {
			return fmt.Errorf("invalid handshake: %v", err)
		}
		fmt.Printf("%+v\n", h)
	case 1:
		var msg []p2p.DiscReason
		if rlp.DecodeBytes(data, &msg); len(msg) == 0 {
			return fmt.Errorf("invalid disconnect message")
		}
		return fmt.Errorf("received disconnect message: %v", msg[0])
	default:
		return fmt.Errorf("invalid message code %d, expected handshake (code zero)", code)
	}
	return nil
}

// rlpxVBGTest runs the VBG protocol test suite.
func rlpxVBGTest(ctx *cli.Context) error {
	if ctx.NArg() < 3 {
		exit("missing path to chain.rlp as command-line argument")
	}
	suite := VBGtest.NewSuite(getNodeArg(ctx), ctx.Args()[1], ctx.Args()[2])
	return runTests(ctx, suite.AllTests())
}