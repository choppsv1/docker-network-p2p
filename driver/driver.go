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
	"fmt"
	. "github.com/choppsv1/p2p-network-driver/logging" // nolint
	"github.com/docker/go-plugins-helpers/network"
	"math/bits"
	"sync"
)

type bitArray uint

type p2pEndpoint struct {
	id  string
	ord uint
	n   *p2pNetwork
	i   *network.EndpointInterface
}

type p2pNetwork struct {
	id        string
	ord       uint
	endpoints map[string]*p2pEndpoint
}

type driver struct {
	alloc    bitArray
	networks map[string]*p2pNetwork
	sync.Mutex
}

func errFmt(format string, a ...interface{}) error {
	return fmt.Errorf("P2P: "+format, a...)
}

func findFirstFreeBit(b bitArray) int {
	for i := 0; i < bits.UintSize; i++ {
		if (b & (1 << i)) == 0 {
			return i
		}
	}
	return -1
}

func Init() (*driver, error) {
	driver := &driver{
		networks: make(map[string]*p2pNetwork),
	}
	return driver, nil
}

func (d *driver) AllocateNetwork(r *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	Trace("AllocateNetwork(%+v)", r)
	return nil, nil
}

// Gets called when docker creates a network
func (d *driver) CreateNetwork(r *network.CreateNetworkRequest) error {
	Trace("CreateNetwork(%+v)", r)

	if _, ok := d.networks[r.NetworkID]; ok {
		return errFmt("Network %s already exists", r.NetworkID)
	}

	ord := findFirstFreeBit(d.alloc)
	if ord < 0 {
		return errFmt("Maximum (%d) networks allocated", bits.UintSize)
	}

	Debug("Creating network: p2p%d: %s", ord, r.NetworkID)

	d.networks[r.NetworkID] = &p2pNetwork{
		id:        r.NetworkID,
		ord:       uint(ord),
		endpoints: make(map[string]*p2pEndpoint),
	}

	return nil
}

// Gets called when deleting a network
func (d *driver) DeleteNetwork(r *network.DeleteNetworkRequest) error {
	Trace("DeleteNetwork(%+v)", r)

	n, ok := d.networks[r.NetworkID]
	if !ok {
		return errFmt("Network %s does not exist", r.NetworkID)
	}

	if count := len(n.endpoints); count != 0 {
		return errFmt("Network %s still has %d endpoints", n.id, count)
	}

	Debug("Deleting network: p2p%d: %s", n.ord, n.id)
	d.alloc &= ^(1 << n.ord)
	delete(d.networks, n.id)
	return nil
}

// Gets called when creating a container
func (d *driver) CreateEndpoint(r *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	Trace("CreateEndpoint(%+v)", r)

	n, ok := d.networks[r.NetworkID]
	if !ok {
		return nil, errFmt("Network %s does not exist", r.NetworkID)
	}
	if len(n.endpoints) > 1 {
		return nil, errFmt("2 endpoints allready attached to P2P network %s", r.NetworkID)
	}

	// Find the next available ordinal number for the network
	ord := uint(0)
	for _, e := range n.endpoints {
		ord = (e.ord + 1) % 2
	}

	Debug("Creating endpoint: %d on p2p%d: (%s, %s)", ord, n.ord, r.EndpointID, n.id)

	n.endpoints[r.EndpointID] = &p2pEndpoint{
		id:  r.EndpointID,
		ord: ord,
		n:   n,
		i:   r.Interface,
	}

	// The response is used to modify the input values, nil for no modification.
	return &network.CreateEndpointResponse{nil}, nil
}

// Gets called when deleting a container
func (d *driver) DeleteEndpoint(r *network.DeleteEndpointRequest) error {
	Trace("DeleteEndpoint(%+v)", r)

	n, ok := d.networks[r.NetworkID]
	if !ok {
		return errFmt("Network %s does not exist", r.NetworkID)
	}
	e, ok := n.endpoints[r.EndpointID]
	if !ok {
		return errFmt("Endpoint %s on network %s does not exist", r.EndpointID, n.id)
	}

	Debug("Deleting endpoint %d on p2p%d (%s, %s)", e.ord, n.ord, e.id, n.id)

	delete(n.endpoints, e.id)
	return nil
}

// Move the container interface into a namespace
func (d *driver) Join(r *network.JoinRequest) (*network.JoinResponse, error) {
	Trace("Join(%+v)", r)
	return nil, nil

	// type JoinRequest struct {
	// 	NetworkID  string
	// 	EndpointID string
	// 	SandboxKey string
	// 	Options    map[string]interface{}
	// }
	// // InterfaceName consists of the name of the interface in the global netns and
	// // the desired prefix to be appended to the interface inside the container netns
	// resp := &network.JoinResponse{
	// 	InterfaceName: network.InterfaceName{
	// 		SrcName:   srcName,
	// 		DstPrefix: dstPrefix,
	// 	},
	// 	DisableGatewayService: true,
	// }
	// // The response is used to modify the input values, nil for no modification.
	// return resp, nil

}

// Remove the container interface from a namespace
func (d *driver) Leave(r *network.LeaveRequest) error {
	Trace("Leave(%+v)", r)
	return nil
}

//
// Rest of API unimplemented
//

func (d *driver) DiscoverDelete(r *network.DiscoveryNotification) error {
	Trace("DiscoverDelete(%+v)", r)
	return nil
}

func (d *driver) DiscoverNew(r *network.DiscoveryNotification) error {
	Trace("DiscoverNew(%+v)", r)
	return nil
}

func (d *driver) EndpointInfo(r *network.InfoRequest) (*network.InfoResponse, error) {
	Trace("EndpointInfo(%+v)", r)
	return nil, nil
}

func (d *driver) FreeNetwork(r *network.FreeNetworkRequest) error {
	Trace("FreeNetwork(%+v)", r)
	return nil
}

func (d *driver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	Trace("GetCapabilities()")
	return &network.CapabilitiesResponse{Scope: network.LocalScope}, nil
}

func (d *driver) ProgramExternalConnectivity(r *network.ProgramExternalConnectivityRequest) error {
	Trace("ProgramExternalConnectivity(%+v)", r)
	return nil
}

func (d *driver) RevokeExternalConnectivity(r *network.RevokeExternalConnectivityRequest) error {
	Trace("RevokeExternalConnectivity(%+v)", r)
	return nil
}
