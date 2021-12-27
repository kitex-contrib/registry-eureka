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

import "errors"

var (
	NilInfoErr          = errors.New("registry info can't be nil")
	NilAddrErr          = errors.New("registry addr can't be nil")
	EmptyServiceNameErr = errors.New("registry service name can't be empty")
	ConvertAddrErr      = errors.New("convert addr error")
	MissIPErr           = errors.New("addr missing ip")
	MissPortErr         = errors.New("addr missing port")
)
