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

package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hudl/fargo"
	"github.com/kitex-contrib/registry-eureka/constants"
	"github.com/kitex-contrib/registry-eureka/entity"
)

type eurekaRegistry struct {
	eurekaConn       *fargo.EurekaConnection
	heatBeatInterval time.Duration
	ctx              context.Context
	cancelFunc       context.CancelFunc
}

// NewEurekaRegistry creates a eureka registry.
func NewEurekaRegistry(servers []string, heatBeatInterval time.Duration) registry.Registry {
	conn := fargo.NewConn(servers...)
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &eurekaRegistry{
		eurekaConn:       &conn,
		heatBeatInterval: heatBeatInterval,
		ctx:              ctx,
		cancelFunc:       cancelFunc,
	}
}

// Register register a server with given registry info.
func (e *eurekaRegistry) Register(info *registry.Info) error {
	instance, err := e.eurekaInstance(info)
	if err != nil {
		return err
	}

	if err = e.eurekaConn.RegisterInstance(instance); err != nil {
		return err
	}

	go e.heartBeat(instance)

	return nil
}

// Deregister deregister a server with given registry info.
func (e *eurekaRegistry) Deregister(info *registry.Info) error {
	instance, err := e.eurekaInstance(info)
	if err != nil {
		return err
	}

	e.cancelFunc()

	if err = e.eurekaConn.DeregisterInstance(instance); err != nil {
		return err
	}

	return nil
}

func (e *eurekaRegistry) eurekaInstance(info *registry.Info) (*fargo.Instance, error) {
	if info == nil {
		return nil, ErrNilInfo
	}

	if info.Addr == nil {
		return nil, ErrNilAddr
	}

	if len(info.ServiceName) == 0 {
		return nil, ErrEmptyServiceName
	}

	addr, ok := info.Addr.(*net.TCPAddr)
	if !ok {
		return nil, ErrConvertAddr
	}

	if addr.IP.String() == "" || addr.IP.String() == "::" {

		return nil, ErrMissIP
	}

	if addr.Port == 0 {
		return nil, ErrMissPort
	}

	if info.Weight == 0 {
		info.Weight = discovery.DefaultWeight
	}

	meta, err := json.Marshal(&entity.RegistryEntity{Weight: info.Weight, Tags: info.Tags})
	if err != nil {
		return nil, err
	}
	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())
	instance := &fargo.Instance{
		HostName:       instanceKey,
		InstanceId:     instanceKey,
		App:            info.ServiceName,
		IPAddr:         addr.IP.String(),
		Port:           addr.Port,
		Status:         fargo.UP,
		DataCenterInfo: fargo.DataCenterInfo{Name: fargo.MyOwn},
	}

	instance.SetMetadataString(constants.Meta, string(meta))
	return instance, nil
}

func (e *eurekaRegistry) heartBeat(ins *fargo.Instance) {
	ticker := time.NewTicker(e.heatBeatInterval)

	for {
		select {

		case <-e.ctx.Done():
			ticker.Stop()
			return

		case <-ticker.C:
			if err := e.eurekaConn.HeartBeatInstance(ins); err != nil {
				klog.Errorf("heartBeat error,err=%+v", err)
			}
		}
	}
}
