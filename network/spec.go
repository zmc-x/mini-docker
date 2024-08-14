package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

// network driver
type NetWorkDriver interface {
	// return the driver name
	Name() string
	// create the network driver
	Create(subnet string, name string) (*NetWork, error)
	// delete the network driver
	Delete(network NetWork) error
	// connect endPoint to network
	Connect(network *NetWork, endPoint *EndPoint) error
	// disconnect endPoint from network
	Disconnect(network *NetWork, endPoint *EndPoint) error
}

// net
type NetWork struct {
	// network name
	Name string
	// ipnet
	IPRange *net.IPNet
	// network driver
	Driver string
}

// net point(entry)
type EndPoint struct {
	ID          string            `json:"id"`
	Device      netlink.Veth      `json:"device"`
	IPAddress   *net.IP           `json:"ip"`
	MACAddress  *net.HardwareAddr `json:"mac"`
	PortMapping []string          `json:"port_map"`
	*NetWork
}

var (
	devices  map[string]NetWorkDriver = make(map[string]NetWorkDriver)
	networks map[string]*NetWork      = make(map[string]*NetWork)
)

const (
	ipamDefaultAllocatorPath = "/var/run/mini-docker/network/ipam/subnet.json"
	defaultNetworkPath       = "/var/run/mini-docker/network"
)
