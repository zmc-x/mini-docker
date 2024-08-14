package network

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllocate(t *testing.T) {
	assert := assert.New(t)
	home := os.Getenv("HOME")
	ipam := NewIPAM(filepath.Join(home, "mini-docker/network/ipam/subnet.json"))
	_, ipnet, _ := net.ParseCIDR("172.18.0.0/24")
	for i := 0; i < 255; i++ {
		ip, err := ipam.Allocate(ipnet)
		if i < 254 {
			assert.Nil(err, "ip allocate should return nil")
			assert.Equal(ipnet.IP, ip.Mask(ipnet.Mask), "these two values should be equal")
		} else {
			assert.Equal(fmt.Errorf("the subnet don't have enough ip address"), err, "correct error")
		}
	}
}

func TestRelease(t *testing.T) {
	home := os.Getenv("HOME")
	ipam := NewIPAM(filepath.Join(home, "mini-docker/network/ipam/subnet.json"))
	ip, ipnet, _ := net.ParseCIDR("172.18.0.1/24")
	ipam.Release(ipnet, &ip)
}
