// -*- coding: utf-8 -*-
//
// May 12 2020, Christian E. Hopps <chopps@gmail.com>
//
// Copyright (c) 2020, Christian E. Hopps
// All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package driver

import (
	"encoding/json"
	. "github.com/choppsv1/docker-network-p2p/logging" // nolint
	"io/ioutil"
	"os"
	"path/filepath"
)

const stateDir = "/etc/docker/docker-network-p2p"
const stateGlob = stateDir + "/*"

func loadNetworkState(fn string) (*p2pNetwork, error) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	n := &p2pNetwork{}
	if err = json.Unmarshal(b, n); err != nil {
		Debug("Error unmarshaling data: %v", err)
		return nil, err
	}
	return n, nil
}

func (d *driver) loadNetworks() error {
	m, _ := filepath.Glob(stateGlob)
	for _, fn := range m {
		n, err := loadNetworkState(fn)
		if err != nil {
			Err("Error loading data for network %s: %v", fn, err)
		}

		if err = d.recreateNetwork(n, true); err != nil {
			return err
		}

		Debug("Restored network: %s", n.ID)
	}
	return nil
}

func (n *p2pNetwork) saveNetworkState() error {
	b, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		Debug("Error marshaling data for network %s: %v", n, err)
		return err
	}
	Debug("Marshaling data for network %s: %v", n.ID, string(b))

	fn := filepath.Join(stateDir, n.ID)
	if err = ioutil.WriteFile(fn, b, 0644); err != nil {
		Debug("Error writing marshaling data to %s: %v", fn, err)
		return err
	}
	return nil
}

func (d *driver) saveState() error {
	for _, n := range d.networks {
		if err := n.saveNetworkState(); err != nil {
			Err("Error marshaling data: %v", err)
		}
	}
	return nil
}

func (n *p2pNetwork) deleteNetworkState() error {
	return os.Remove(filepath.Join(stateDir, n.ID))
}
