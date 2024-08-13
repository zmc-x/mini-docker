package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBridgeCreate(t *testing.T) {
	assert := assert.New(t)
	b := Bridge{}
	_, err := b.Create("172.24.0.1/24", "mini-docker")
	assert.Nil(err, "error is nil")
}

func TestBridgeConnect(t *testing.T) {
	assert := assert.New(t)
	nw := NetWork{
		Name: "mini-docker",
	}
	ep := EndPoint{
		ID: "test-veth",
	}
	b := Bridge{}
	err := b.Connect(&nw, &ep)
	assert.Nil(err, "error is nil")
}

func TestBridgeDelete(t *testing.T) {
	assert := assert.New(t)
	nw := NetWork{
		Name: "mini-docker",
	}
	b := Bridge{}
	err := b.Delete(nw)
	assert.Nil(err, "error is nil")
}