package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type Bridge struct{}

func (b *Bridge) Name() string {
	return "bridge"
}

func (b *Bridge) Create(subnet string, name string) (*NetWork, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	network := &NetWork{
		Name:    name,
		IPRange: ipRange,
		Driver:  b.Name(),
	}

	err := b.initBridge(network)
	if err != nil {
		zap.L().Sugar().Errorf("init bridge error %v", err)
		return nil, fmt.Errorf("init bridge error")
	}
	return network, nil
}

func (b *Bridge) Delete(network NetWork) error {
	bridgeName := network.Name
	device, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	if err := netlink.LinkDel(device); err != nil {
		return err
	}
	return removeIPTables(bridgeName, network.IPRange)
}

func (b *Bridge) Connect(network *NetWork, endPoint *EndPoint) error {
	bridgeName := network.Name
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// create veth
	la := netlink.NewLinkAttrs()
	la.Name = endPoint.ID[:5]
	la.MasterIndex = bridge.Attrs().Index

	endPoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "netns-" + endPoint.ID[:5],
	}
	// equivalent to: ip link add {veth name} type veth peer {peer veth name}
	// 				  ip link set {veth name} master {bridge name}
	if err := netlink.LinkAdd(&endPoint.Device); err != nil {
		return fmt.Errorf("add endpoint error %v", err)
	}
	// equivalent to: ip link set {veth} up
	if err := netlink.LinkSetUp(&endPoint.Device); err != nil {
		return fmt.Errorf("set device up error %v", err)
	}
	return nil
}

func (b *Bridge) Disconnect(network *NetWork, endPoint *EndPoint) error {
	return nil
}

func (b *Bridge) initBridge(network *NetWork) error {
	// create the bridge
	bridgeName := network.Name
	if err := createBridge(bridgeName); err != nil {
		return fmt.Errorf("error add bridge %s, error is %v", bridgeName, err)
	}
	// set the ip to the bridge
	gatewayip := *network.IPRange

	if err := setInterfaceIP(bridgeName, gatewayip.String()); err != nil {
		return fmt.Errorf("set bridge %s ip error %v", bridgeName, err)
	}
	// set the bridge up
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("set bridge %s up error %v", bridgeName, err)
	}
	// iptables
	if err := setIPTables(bridgeName, network.IPRange); err != nil {
		return fmt.Errorf("set iptables error %v", err)
	}
	return nil
}

// create the linux bridge
func createBridge(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// init ip link
	iplink := netlink.NewLinkAttrs()
	iplink.Name = bridgeName
	// create bridge
	br := &netlink.Bridge{
		LinkAttrs: iplink,
	}
	// equivalent to: ip link add br type bridge
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge %s create error %v", bridgeName, err)
	}
	return nil
}

// set the interface ip
func setInterfaceIP(bridgeName string, ip string) error {
	device, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	ipNet, err := netlink.ParseIPNet(ip)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{
		IPNet: ipNet,
		Peer:  ipNet,
	}
	// equivalent to: ip addr add {ip_address} dev {device}
	return netlink.AddrAdd(device, addr)
}

// set the interface up
func setInterfaceUP(bridgeName string) error {
	device, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	// equivalent to: ip link set {device} up
	err = netlink.LinkSetUp(device)
	if err != nil {
		return err
	}
	return nil
}

func setIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesArgs := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesArgs, " ")...)
	output, err := cmd.Output()
	if err != nil {
		zap.L().Sugar().Errorf("iptables output %v", output)
		return fmt.Errorf("set iptables error %v", err)
	}
	return nil
}

func removeIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesArgs := fmt.Sprintf("-t nat -D POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesArgs, " ")...)
	output, err := cmd.Output()
	if err != nil {
		zap.L().Sugar().Errorf("iptables output %v", output)
		return fmt.Errorf("remove iptables error %v", err)
	}
	return nil
}