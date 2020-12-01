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

package abi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/vbgloble/go-VGB/common"
	"github.com/vbgloble/go-VGB/crypto"
)

// The ABI holds information about a contract's context and available
// invokable mVBGods. It will allow you to type check function calls and
// packs data accordingly.
type ABI struct {
	Constructor MVBGod
	MVBGods     map[string]MVBGod
	Events      map[string]Event

	// Additional "special" functions introduced in solidity v0.6.0.
	// It's separated from the original default fallback. Each contract
	// can only define one fallback and receive function.
	Fallback MVBGod // Note it's also used to represent legacy fallback before v0.6.0
	Receive  MVBGod
}

// JSON returns a parsed ABI interface and error if it failed.
func JSON(reader io.Reader) (ABI, error) {
	dec := json.NewDecoder(reader)

	var abi ABI
	if err := dec.Decode(&abi); err != nil {
		return ABI{}, err
	}
	return abi, nil
}

// Pack the given mVBGod name to conform the ABI. MVBGod call's data
// will consist of mVBGod_id, args0, arg1, ... argN. MVBGod id consists
// of 4 bytes and arguments are all 32 bytes.
// MVBGod ids are created from the first 4 bytes of the hash of the
// mVBGods string signature. (signature = baz(uint32,string32))
func (abi ABI) Pack(name string, args ...interface{}) ([]byte, error) {
	// Fetch the ABI of the requested mVBGod
	if name == "" {
		// constructor
		arguments, err := abi.Constructor.Inputs.Pack(args...)
		if err != nil {
			return nil, err
		}
		return arguments, nil
	}
	mVBGod, exist := abi.MVBGods[name]
	if !exist {
		return nil, fmt.Errorf("mVBGod '%s' not found", name)
	}
	arguments, err := mVBGod.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	// Pack up the mVBGod ID too if not a constructor and return
	return append(mVBGod.ID, arguments...), nil
}

func (abi ABI) getArguments(name string, data []byte) (Arguments, error) {
	// since there can't be naming collisions with contracts and events,
	// we need to decide whVBGer we're calling a mVBGod or an event
	var args Arguments
	if mVBGod, ok := abi.MVBGods[name]; ok {
		if len(data)%32 != 0 {
			return nil, fmt.Errorf("abi: improperly formatted output: %s - Bytes: [%+v]", string(data), data)
		}
		args = mVBGod.Outputs
	}
	if event, ok := abi.Events[name]; ok {
		args = event.Inputs
	}
	if args == nil {
		return nil, errors.New("abi: could not locate named mVBGod or event")
	}
	return args, nil
}

// Unpack unpacks the output according to the abi specification.
func (abi ABI) Unpack(name string, data []byte) ([]interface{}, error) {
	args, err := abi.getArguments(name, data)
	if err != nil {
		return nil, err
	}
	return args.Unpack(data)
}

// UnpackIntoInterface unpacks the output in v according to the abi specification.
// It performs an additional copy. Please only use, if you want to unpack into a
// structure that does not strictly conform to the abi structure (e.g. has additional arguments)
func (abi ABI) UnpackIntoInterface(v interface{}, name string, data []byte) error {
	args, err := abi.getArguments(name, data)
	if err != nil {
		return err
	}
	unpacked, err := args.Unpack(data)
	if err != nil {
		return err
	}
	return args.Copy(v, unpacked)
}

// UnpackIntoMap unpacks a log into the provided map[string]interface{}.
func (abi ABI) UnpackIntoMap(v map[string]interface{}, name string, data []byte) (err error) {
	args, err := abi.getArguments(name, data)
	if err != nil {
		return err
	}
	return args.UnpackIntoMap(v, data)
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (abi *ABI) UnmarshalJSON(data []byte) error {
	var fields []struct {
		Type    string
		Name    string
		Inputs  []Argument
		Outputs []Argument

		// Status indicator which can be: "pure", "view",
		// "nonpayable" or "payable".
		StateMutability string

		// Deprecated Status indicators, but removed in v0.6.0.
		Constant bool // True if function is either pure or view
		Payable  bool // True if function is payable

		// Event relevant indicator represents the event is
		// declared as anonymous.
		Anonymous bool
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	abi.MVBGods = make(map[string]MVBGod)
	abi.Events = make(map[string]Event)
	for _, field := range fields {
		switch field.Type {
		case "constructor":
			abi.Constructor = NewMVBGod("", "", Constructor, field.StateMutability, field.Constant, field.Payable, field.Inputs, nil)
		case "function":
			name := abi.overloadedMVBGodName(field.Name)
			abi.MVBGods[name] = NewMVBGod(name, field.Name, Function, field.StateMutability, field.Constant, field.Payable, field.Inputs, field.Outputs)
		case "fallback":
			// New introduced function type in v0.6.0, check more detail
			// here https://solidity.readthedocs.io/en/v0.6.0/contracts.html#fallback-function
			if abi.HasFallback() {
				return errors.New("only single fallback is allowed")
			}
			abi.Fallback = NewMVBGod("", "", Fallback, field.StateMutability, field.Constant, field.Payable, nil, nil)
		case "receive":
			// New introduced function type in v0.6.0, check more detail
			// here https://solidity.readthedocs.io/en/v0.6.0/contracts.html#fallback-function
			if abi.HasReceive() {
				return errors.New("only single receive is allowed")
			}
			if field.StateMutability != "payable" {
				return errors.New("the statemutability of receive can only be payable")
			}
			abi.Receive = NewMVBGod("", "", Receive, field.StateMutability, field.Constant, field.Payable, nil, nil)
		case "event":
			name := abi.overloadedEventName(field.Name)
			abi.Events[name] = NewEvent(name, field.Name, field.Anonymous, field.Inputs)
		default:
			return fmt.Errorf("abi: could not recognize type %v of field %v", field.Type, field.Name)
		}
	}
	return nil
}

// overloadedMVBGodName returns the next available name for a given function.
// Needed since solidity allows for function overload.
//
// e.g. if the abi contains MVBGods send, send1
// overloadedMVBGodName would return send2 for input send.
func (abi *ABI) overloadedMVBGodName(rawName string) string {
	name := rawName
	_, ok := abi.MVBGods[name]
	for idx := 0; ok; idx++ {
		name = fmt.Sprintf("%s%d", rawName, idx)
		_, ok = abi.MVBGods[name]
	}
	return name
}

// overloadedEventName returns the next available name for a given event.
// Needed since solidity allows for event overload.
//
// e.g. if the abi contains events received, received1
// overloadedEventName would return received2 for input received.
func (abi *ABI) overloadedEventName(rawName string) string {
	name := rawName
	_, ok := abi.Events[name]
	for idx := 0; ok; idx++ {
		name = fmt.Sprintf("%s%d", rawName, idx)
		_, ok = abi.Events[name]
	}
	return name
}

// MVBGodById looks up a mVBGod by the 4-byte id,
// returns nil if none found.
func (abi *ABI) MVBGodById(sigdata []byte) (*MVBGod, error) {
	if len(sigdata) < 4 {
		return nil, fmt.Errorf("data too short (%d bytes) for abi mVBGod lookup", len(sigdata))
	}
	for _, mVBGod := range abi.MVBGods {
		if bytes.Equal(mVBGod.ID, sigdata[:4]) {
			return &mVBGod, nil
		}
	}
	return nil, fmt.Errorf("no mVBGod with id: %#x", sigdata[:4])
}

// EventByID looks an event up by its topic hash in the
// ABI and returns nil if none found.
func (abi *ABI) EventByID(topic common.Hash) (*Event, error) {
	for _, event := range abi.Events {
		if bytes.Equal(event.ID.Bytes(), topic.Bytes()) {
			return &event, nil
		}
	}
	return nil, fmt.Errorf("no event with id: %#x", topic.Hex())
}

// HasFallback returns an indicator whVBGer a fallback function is included.
func (abi *ABI) HasFallback() bool {
	return abi.Fallback.Type == Fallback
}

// HasReceive returns an indicator whVBGer a receive function is included.
func (abi *ABI) HasReceive() bool {
	return abi.Receive.Type == Receive
}

// revertSelector is a special function selector for revert reason unpacking.
var revertSelector = crypto.Keccak256([]byte("Error(string)"))[:4]

// UnpackRevert resolves the abi-encoded revert reason. According to the solidity
// spec https://solidity.readthedocs.io/en/latest/control-structures.html#revert,
// the provided revert reason is abi-encoded as if it were a call to a function
// `Error(string)`. So it's a special tool for it.
func UnpackRevert(data []byte) (string, error) {
	if len(data) < 4 {
		return "", errors.New("invalid data for unpacking")
	}
	if !bytes.Equal(data[:4], revertSelector) {
		return "", errors.New("invalid data for unpacking")
	}
	typ, _ := NewType("string", "", nil)
	unpacked, err := (Arguments{{Type: typ}}).Unpack(data[4:])
	if err != nil {
		return "", err
	}
	return unpacked[0].(string), nil
}
