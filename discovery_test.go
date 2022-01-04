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
