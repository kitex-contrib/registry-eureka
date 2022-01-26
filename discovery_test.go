// Copyright 2021 CloudWeGo authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/registry-eureka/resolver"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/stretchr/testify/assert"

	registry2 "github.com/kitex-contrib/registry-eureka/registry"
)

func TestEurekaDiscovery(t *testing.T) {
	var err error
	r := registry2.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)
	tags := map[string]string{"idc": "hl"}
	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}
	info := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         tags,
		Addr:         addr,
	}
	err = r.Register(info)
	assert.Nil(t, err)

	res := resolver.NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})

	target := res.Target(context.Background(), rpcinfo.NewEndpointInfo("test", "", nil, nil))

	// prevent perception delay to affect test case
	time.Sleep(30 * time.Second)

	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(result.Instances))

	instance := result.Instances[0]
	assert.Equal(t, addr.String(), instance.Address().String())
	assert.Equal(t, info.Weight, instance.Weight())

	for k, v := range info.Tags {
		v1, exist := instance.Tag(k)
		assert.Equal(t, true, exist)
		assert.Equal(t, v, v1)
	}

	// deregister
	err = r.Deregister(info)
	assert.Nil(t, err)

	// prevent perception delay to affect test case
	time.Sleep(30 * time.Second)

	// resolve again
	result, _ = res.Resolve(context.Background(), target)
	assert.Equal(t, 0, len(result.Instances))
}

func TestEurekaDiscoveryWithMultipleInstance(t *testing.T) {
	r := registry2.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)
	info1 := &registry.Info{
		ServiceName:  "test",
		Weight:       11,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1},
	}
	info2 := &registry.Info{
		ServiceName:  "test",
		Weight:       12,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2},
	}
	info3 := &registry.Info{
		ServiceName:  "test",
		Weight:       13,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3},
	}
	addrMap := map[string]int{
		info1.Addr.String(): info1.Weight,
		info2.Addr.String(): info2.Weight,
		info3.Addr.String(): info3.Weight,
	}

	assert.Nil(t, r.Register(info1))
	assert.Nil(t, r.Register(info2))
	assert.Nil(t, r.Register(info3))

	res := resolver.NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})

	target := res.Target(context.Background(), rpcinfo.NewEndpointInfo("test", "", nil, nil))

	// prevent perception delay to affect test case
	time.Sleep(30 * time.Second)

	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Len(t, result.Instances, 3)
	instances := result.Instances
	for _, instance := range instances {
		addr := instance.Address().String()
		weight, ok := addrMap[addr]
		assert.Equal(t, ok, true)
		assert.Equal(t, weight, instance.Weight())
		v1, exist := instance.Tag("idc")
		assert.Equal(t, true, exist)
		assert.Equal(t, "hl", v1)
	}

	assert.Nil(t, r.Deregister(info1))
	assert.Nil(t, r.Deregister(info2))

	// prevent perception delay to affect test case
	time.Sleep(30 * time.Second)

	result, err = res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Equal(t, len(result.Instances), 1)
	instance := result.Instances[0]
	assert.Equal(t, instance.Weight(), info3.Weight)
	assert.Equal(t, instance.Address().String(), info3.Addr.String())

	assert.Nil(t, r.Deregister(info3))

	// prevent perception delay to affect test case
	time.Sleep(30 * time.Second)

	result, err = res.Resolve(context.Background(), target)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, len(result.Instances), 0)
}

func TestEurekaDiscoveryWithInvalidInstanceInfo(t *testing.T) {
	r := registry2.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)

	assert.Equal(t, registry2.ErrNilInfo, r.Register(nil))

	info1 := &registry.Info{
		ServiceName:  "",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1},
	}
	assert.Equal(t, registry2.ErrEmptyServiceName, r.Register(info1))

	info2 := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         nil,
	}
	assert.Equal(t, registry2.ErrNilAddr, r.Register(info2))

	info3 := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{Port: 1},
	}
	assert.Equal(t, registry2.ErrMissIP, r.Register(info3))

	info4 := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)},
	}
	assert.Equal(t, registry2.ErrMissPort, r.Register(info4))
}
