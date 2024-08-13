package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string][]int64
}

var ipamAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func NewIPAM(ipamAllocatorPath string) *IPAM {
	return &IPAM{
		SubnetAllocatorPath: ipamAllocatorPath,
	}
}

// load the ipnet ip allocation
func (m *IPAM) load() error {
	if _, err := os.Stat(m.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	f, err := os.Open(m.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 2000)
	n, err := f.Read(buffer)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer[:n], m.Subnets)
	if err != nil {
		zap.L().Sugar().Errorf("unmarshal ipam config file error %v", err)
		return err
	}
	return nil
}

// storage ip allocate
func (m *IPAM) storage() error {
	ipamConfigPath, _ := filepath.Split(m.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ipamConfigPath, 0644); err != nil {
				zap.L().Sugar().Errorf("mkdir dir %s error %v", ipamConfigPath, err)
				return err
			}
		} else {
			return err
		}
	}

	f, err := os.OpenFile(m.SubnetAllocatorPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	ipamJSON, err := json.Marshal(m.Subnets)
	if err != nil {
		return err
	}

	_, err = f.Write(ipamJSON)
	if err != nil {
		return err
	}
	return nil
}

// allocate ip address
func (m *IPAM) Allocate(subnet *net.IPNet) (*net.IP, error) {
	subnetIP, subnet, _ := net.ParseCIDR(subnet.String())
	subnetIP = subnetIP.To4()
	one, bits := subnet.Mask.Size()
	delta := bits - one 
	var cnt, total int64 
	cnt = 1 << max(delta - 6, 0)
	total = 1 << delta - 2
	
	m.Subnets = &map[string][]int64{}
	err := m.load()
	if err != nil {
		zap.L().Sugar().Warnf("load the config file error %v", err)
	}

	// init the bitmap
	if _, ok := (*m.Subnets)[subnet.String()]; !ok {
		(*m.Subnets)[subnet.String()] = make([]int64, cnt)
	}
	
	for idx, v := range (*m.Subnets)[subnet.String()] {
		if v == -1 {continue}
		for i := 63; i >= 0; i-- {
			if v >> i & 1 == 0 {
				(*m.Subnets)[subnet.String()][idx] = (*m.Subnets)[subnet.String()][idx] | 1 << i
				n := int64(64 * idx + 64 - i)
				if n <= total {
					for j := 3; j >= 0; j-- {
						subnetIP[j] += uint8(n >> (8 * (3 - j)))
					}
					err = m.storage()
					if err != nil {
						zap.L().Sugar().Warnf("storage error %v", err)
					}
					return &subnetIP, nil
				}
				return nil, fmt.Errorf("the subnet don't have enough ip address")
			}
		}
	}
	err = m.storage()
	if err != nil {
		zap.L().Sugar().Warnf("storage error %v", err)
	}
	return nil, fmt.Errorf("the subnet don't have enough ip address")
}

// release ip address
func (m *IPAM) Release(subnet *net.IPNet, oldIP *net.IP) error {
	m.Subnets = &map[string][]int64{}
	err := m.load()
	if err != nil {
		zap.L().Sugar().Warnf("load the ipam config error %v", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())
	mask := []byte(subnet.Mask)
	ipv4 := []byte(oldIP.To4())
	var total int64 
	for i := 0; i < 4; i++ {
		diff := ipv4[i] - (mask[i] & ipv4[i])
		total += int64(diff) * (1 << ((3 - i) * 8))
	}
	
	for i := range (*m.Subnets)[subnet.String()] {
		if total > 64 {
			total -= 64
			continue
		} 
		(*m.Subnets)[subnet.String()][i] = ^int64(1 << (64 - total)) & (*m.Subnets)[subnet.String()][i]
		break
	}
	err = m.storage()
	if err != nil {
		zap.L().Sugar().Warnf("storage error %v", err)
	}
	return nil
} 
