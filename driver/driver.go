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
	"fmt"
	. "github.com/choppsv1/docker-network-p2p/logging" // nolint
	"github.com/docker/go-plugins-helpers/network"
	"github.com/vishvananda/netlink"
	"math/bits"
	"os"
	"sync"
	"syscall"
)

type bitArray uint

type p2pEndpoint struct {
	ID         string                     `json:"endpoint-id"`
	Ord        uint                       `json:"ordinal"`
	I          *network.EndpointInterface `json:"interface,omitempty"`
	sandboxKey string
}

type p2pNetwork struct {
	ID        string                  `json:"network-id"`
	Ord       uint                    `json:"ordinal"`
	Endpoints map[string]*p2pEndpoint `json:"endpoints"`
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
	d := &driver{
		networks: make(map[string]*p2pNetwork),
	}
	_ = os.Mkdir(stateDir, 0755)
	d.loadNetworks()
	return d, nil
}

func intfName(netOrd, ifOrd uint) string {
	return fmt.Sprintf("p2p%d-%d", netOrd, ifOrd)
}

func (d *driver) recreateNetwork(n *p2pNetwork, existsOk bool) error {
	// Create the veth pair
	a := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: intfName(n.Ord, 0),
		},

		PeerName: intfName(n.Ord, 1),
	}
	if err := netlink.LinkAdd(a); err != nil {
		if !existsOk || err != syscall.EEXIST {
			Debug("Error creating veth interface pair for %s: %v", n.ID, err)
			return errFmt("Creating veth interface pair for %s: %v", n.ID, err)
		}
		Info("Ignoring existing of interfaces on network recreate")
	}

	d.alloc |= (1 << n.Ord)
	d.networks[n.ID] = n

	return nil
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

	n := &p2pNetwork{
		ID:        r.NetworkID,
		Ord:       uint(ord),
		Endpoints: make(map[string]*p2pEndpoint),
	}

	if err := n.saveNetworkState(); err != nil {
		return errFmt("Saving state for network %s", n.ID)
	}

	if err := d.recreateNetwork(n, false); err != nil {
		n.deleteNetworkState()
		return err
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

	if count := len(n.Endpoints); count != 0 {
		return errFmt("Network %s still has %d endpoints", n.ID, count)
	}

	Debug("Deleting network: p2p%d: %s", n.Ord, n.ID)

	d.alloc &= ^(1 << n.Ord)
	delete(d.networks, n.ID)
	n.deleteNetworkState()

	// Delete the veth pair
	a := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: intfName(n.Ord, 0),
		},

		PeerName: intfName(n.Ord, 1),
	}
	if err := netlink.LinkDel(a); err != nil {
		return errFmt("Removing veth interface pair for %s: %v", n.ID, err)
	}
	return nil
}

// Gets called when creating a container
func (d *driver) CreateEndpoint(r *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	Trace("CreateEndpoint(%+v)", r)

	n, ok := d.networks[r.NetworkID]
	if !ok {
		return nil, errFmt("Network %s does not exist", r.NetworkID)
	}
	if len(n.Endpoints) > 1 {
		return nil, errFmt("2 endpoints allready attached to P2P network %s", r.NetworkID)
	}

	// Find the next available ordinal number for the network
	ord := uint(0)
	for _, e := range n.Endpoints {
		ord = (e.Ord + 1) % 2
	}

	Debug("Creating endpoint: %d on p2p%d: (%s, %s)", ord, n.Ord, r.EndpointID, n.ID)

	e := &p2pEndpoint{
		ID:  r.EndpointID,
		Ord: ord,
		I:   r.Interface,
	}
	n.Endpoints[r.EndpointID] = e

	if err := n.saveNetworkState(); err != nil {
		// XXX What so now we have bad network state?
		delete(n.Endpoints, r.EndpointID)
		return nil, errFmt("Saving state for endpoint %s", e.ID)
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
	e, ok := n.Endpoints[r.EndpointID]
	if !ok {
		return errFmt("Endpoint %s on network %s does not exist", r.EndpointID, n.ID)
	}

	Debug("Deleting endpoint %d on p2p%d (%s, %s)", e.Ord, n.Ord, e.ID, n.ID)

	delete(n.Endpoints, e.ID)
	return nil
}

// Move the container interface into a namespace
func (d *driver) Join(r *network.JoinRequest) (*network.JoinResponse, error) {
	Trace("Join(%+v)", r)

	n, ok := d.networks[r.NetworkID]
	if !ok {
		return nil, errFmt("Network %s does not exist", r.NetworkID)
	}

	var e *p2pEndpoint
	if e, ok = n.Endpoints[r.EndpointID]; !ok {
		return nil, errFmt("Endpoint %s does not exist", r.EndpointID)
	}

	// type JoinRequest struct {
	// 	NetworkID  string
	// 	EndpointID string
	// 	SandboxKey string
	// 	Options    map[string]interface{}
	// }
	// // InterfaceName consists of the name of the interface in the global netns and
	// // the desired prefix to be appended to the interface inside the container netns

	res := &network.JoinResponse{
		InterfaceName: network.InterfaceName{
			SrcName:   intfName(n.Ord, e.Ord),
			DstPrefix: "p2p",
		},
		DisableGatewayService: true,
	}
	// The response is used to modify the input values, nil for no modification.
	return res, nil

}

// Remove the container interface from a namespace
func (d *driver) Leave(r *network.LeaveRequest) error {
	Trace("Leave(%+v)", r)
	return nil
}

//
// Rest of API unimplemented
//

func (d *driver) AllocateNetwork(r *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	Trace("AllocateNetwork(%+v)", r)
	return nil, nil
}

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

	n, ok := d.networks[r.NetworkID]
	if !ok {
		return nil, errFmt("Network %s does not exist", r.NetworkID)
	}

	var e *p2pEndpoint
	if e, ok = n.Endpoints[r.EndpointID]; !ok {
		return nil, errFmt("Endpoint %s does not exist", r.EndpointID)
	}

	b, err := json.Marshal(e)
	if err != nil {
		return nil, errFmt("Marshall of endpoint %s failed: %v", r.EndpointID, err)
	}
	res := &network.InfoResponse{make(map[string]string)}
	res.Value["data"] = string(b)
	return res, nil
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
