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
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/vbgloble/go-VGB/cmd/utils"
	"github.com/vbgloble/go-VGB/VBG"
	"github.com/vbgloble/go-VGB/internal/VBGapi"
	"github.com/vbgloble/go-VGB/log"
	"github.com/vbgloble/go-VGB/node"
	"github.com/vbgloble/go-VGB/params"
	"github.com/naoina/toml"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...), whisperFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type VBGstatsConfig struct {
	URL string `toml:",omitempty"`
}

// whisper has been deprecated, but clients out there might still have [Shh]
// in their config, which will crash. Cut them some slack by keeping the
// config, and displaying a message that those config switches are ineffectual.
// To be removed circa Q1 2021 -- @gballet.
type whisperDeprecatedConfig struct {
	MaxMessageSize                        uint32  `toml:",omitempty"`
	MinimumAcceptedPOW                    float64 `toml:",omitempty"`
	RestrictConnectionBetweenLightClients bool    `toml:",omitempty"`
}

type gVBGConfig struct {
	VBG      VBG.Config
	Shh      whisperDeprecatedConfig
	Node     node.Config
	VBGstats VBGstatsConfig
}

func loadConfig(file string, cfg *gVBGConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "VBG")
	cfg.WSModules = append(cfg.WSModules, "VBG")
	cfg.IPCPath = "gVBG.ipc"
	return cfg
}

// makeConfigNode loads gVBG configuration and creates a blank node instance.
func makeConfigNode(ctx *cli.Context) (*node.Node, gVBGConfig) {
	// Load defaults.
	cfg := gVBGConfig{
		VBG:  VBG.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}

		if cfg.Shh != (whisperDeprecatedConfig{}) {
			log.Warn("Deprecated whisper config detected. Whisper has been moved to github.com/vbgloble/whisper")
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetVBGConfig(ctx, stack, &cfg.VBG)
	if ctx.GlobalIsSet(utils.VBGStatsURLFlag.Name) {
		cfg.VBGstats.URL = ctx.GlobalString(utils.VBGStatsURLFlag.Name)
	}
	utils.SetShhConfig(ctx, stack)

	return stack, cfg
}

// enableWhisper returns true in case one of the whisper flags is set.
func checkWhisper(ctx *cli.Context) {
	for _, flag := range whisperFlags {
		if ctx.GlobalIsSet(flag.GetName()) {
			log.Warn("deprecated whisper flag detected. Whisper has been moved to github.com/vbgloble/whisper")
		}
	}
}

// makeFullNode loads gVBG configuration and creates the vbgloble backend.
func makeFullNode(ctx *cli.Context) (*node.Node, VBGapi.Backend) {
	stack, cfg := makeConfigNode(ctx)

	backend := utils.RegisterVBGService(stack, &cfg.VBG)

	checkWhisper(ctx)
	// Configure GraphQL if requested
	if ctx.GlobalIsSet(utils.GraphQLEnabledFlag.Name) {
		utils.RegisterGraphQLService(stack, backend, cfg.Node)
	}
	// Add the vbgloble Stats daemon if requested.
	if cfg.VBGstats.URL != "" {
		utils.RegisterVBGStatsService(stack, backend, cfg.VBGstats.URL)
	}
	return stack, backend
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.VBG.Genesis != nil {
		cfg.VBG.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}
