package network

import (
	"encoding/json"
	"fmt"
	"mini-docker/container"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.uber.org/zap"
)

// callback mean return origin cyberspace
type callback func()

func CreateNetwork(name, subnet, device string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	// the first ip
	gatewayIP, err := ipamAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = *gatewayIP

	nw, err := devices[device].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	return nw.storage(defaultNetworkPath)
}

// init
func Init() error {
	// check defaultNetworkPath
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(defaultNetworkPath, 0644); err != nil {
				return fmt.Errorf("mkdir dir %s error %v", defaultNetworkPath, err)
			}
		} else {
			return err
		}
	}
	// init networks
	files, err := os.ReadDir(defaultNetworkPath)
	if err != nil {
		return fmt.Errorf("read dir %s error %v", defaultNetworkPath, err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		nw := &NetWork{
			Name: filename,
		}
		if err := nw.load(filepath.Join(defaultNetworkPath, file.Name())); err != nil {
			return err
		}
		networks[filename] = nw
	}
	// init devices
	bridge := &Bridge{}
	devices[bridge.Name()] = bridge
	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIPRange\tDriver\n")
	for _, network := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", network.Name, network.IPRange, network.Driver)
	}
	if err := w.Flush(); err != nil {
		zap.L().Sugar().Errorf("flush error %v", err)
	}
}

func RemoveNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("the network %s don't exist", networkName)
	}
	// release subnet
	err := ipamAllocator.RemoveSubNet(nw.IPRange)
	if err != nil {
		return fmt.Errorf("release ip error %v", err)
	}
	// delete driver
	err = devices[nw.Driver].Delete(*nw)
	if err != nil {
		return fmt.Errorf("delete driver %s error %v", devices[nw.Driver].Name(), err)
	}
	return nw.remove(networkName)
}

func Connect(networkName string, containerMeta *container.ContainerMeta) error {
	// find network
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("don't have this network")
	}
	// allocate ip
	ip, err := ipamAllocator.Allocate(nw.IPRange)
	if err != nil {
		return fmt.Errorf("allocate ip error %v", err)
	}
	// endpoint
	ep := &EndPoint{
		ID:          fmt.Sprintf("%s-%s", containerMeta.ID, networkName),
		IPAddress:   ip,
		PortMapping: strings.Split(containerMeta.Port, " "),
		NetWork:     nw,
	}
	err = devices[nw.Driver].Connect(nw, ep)
	if err != nil {
		return err
	}

	err = configEndPointIPAndRoute(ep, containerMeta)
	if err != nil {
		return err
	}

	err = configPortMap(ep)
	if err != nil {
		return err
	}
	// write network information to config
	var containerIP net.IPNet = *nw.IPRange
	containerIP.IP = *ip
	return container.WriteNetwork(containerIP, containerMeta.Name)
}

// DisConnect to release the ip
func DisConnect(containerName string) error {
	containerMeta, err := container.GetContainerByName(containerName)
	if err != nil {
		zap.L().Sugar().Errorf("get container information error %v", err)
		return err
	}
	ip := containerMeta.IP
	if ip == "" {
		return nil
	}
	containerIP, subnet, _ := net.ParseCIDR(ip)
	return ipamAllocator.Release(subnet, &containerIP)
}

// config port map
func configPortMap(ep *EndPoint) error {
	for _, pm := range ep.PortMapping {
		portMap := strings.Split(pm, ":")
		if len(portMap) == 2 && portMap[0] != "" && portMap[1] != "" {
			host, container := portMap[0], portMap[1]

			iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s", host, ep.IPAddress.String(), container)
			cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
			output, err := cmd.Output()
			if err != nil {
				zap.L().Sugar().Errorf("iptables output %v", output)
			}
		} else {
			zap.L().Sugar().Error("port mapping format error, port mapping: %s", pm)
		}
	}
	return nil
}

// config ip and route in container
func configEndPointIPAndRoute(ep *EndPoint, containerMeta *container.ContainerMeta) error {
	// config veth
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail to config endpoint: %v", err)
	}

	// restore namespace
	defer enterNetNamespace(&peerLink, containerMeta)()
	// net namespace ip
	var interfaceIP net.IPNet = *ep.IPRange
	interfaceIP.IP = *ep.IPAddress

	if err := setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("set interface %s error %v", ep.Device.PeerName, err)
	}
	if err := setInterfaceUP(ep.Device.PeerName); err != nil {
		return fmt.Errorf("set interface %s error %v", ep.Device.PeerName, err)
	}
	// set the lo up
	if err := setInterfaceUP("lo"); err != nil {
		return fmt.Errorf("set interface lo error %v", err)
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// add route
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.IPRange.IP,
		Dst:       cidr,
	}
	// equivalent to: route add default gw {gateway}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("add route error %v", err)
	}
	return nil
}

func enterNetNamespace(eplink *netlink.Link, containerMeta *container.ContainerMeta) callback {
	// find fd
	f, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", containerMeta.PID), os.O_RDONLY, 0)
	if err != nil {
		zap.L().Sugar().Errorf("open /proc/%d/ns/net error %v", containerMeta.PID, err)
	}
	fd := f.Fd()

	runtime.LockOSThread()
	// equivalent to: ip link set {device} netns { PID | NETNSNAME | NETNSFILE }
	if err := netlink.LinkSetNsFd(*eplink, int(fd)); err != nil {
		zap.L().Sugar().Errorf("set link netns error %v", err)
	}
	// get origin net namespace
	origin, err := netns.Get()
	if err != nil {
		zap.L().Sugar().Errorf("get current net namespace error %v", err)
	}
	// entry netns
	if err := netns.Set(netns.NsHandle(fd)); err != nil {
		zap.L().Sugar().Errorf("entry /proc/%d/ns/net netns error %v", containerMeta.PID, err)
	}
	// return origin namespace(restore)
	return func() {
		netns.Set(origin)
		origin.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

// remove method is mean to remove the configuration of network
func (nw *NetWork) remove(networkPath string) error {
	if _, err := os.Stat(networkPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(networkPath)
}

func (nw *NetWork) load(networkPath string) error {
	if _, err := os.Stat(networkPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	f, err := os.Open(networkPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 2000)
	n, err := f.Read(buffer)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buffer[:n], nw)
	if err != nil {
		zap.L().Sugar().Errorf("unmarshal network config file error %v", err)
		return err
	}
	return nil
}

func (nw *NetWork) storage(networkPath string) error {
	if _, err := os.Stat(networkPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(networkPath, 0644); err != nil {
				zap.L().Sugar().Errorf("mkdir dir %s error %v", networkPath, err)
				return err
			}
		} else {
			return err
		}
	}

	fileName := filepath.Join(networkPath, nw.Name+".json")
	f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	networkJSON, err := json.Marshal(nw)
	if err != nil {
		return err
	}

	_, err = f.Write(networkJSON)
	if err != nil {
		return err
	}
	return nil
}
