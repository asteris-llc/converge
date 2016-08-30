// Copyright © 2016 Asteris, LLC
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

package rpc

import (
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

type contentMarshaler struct {
	*runtime.JSONPb
}

func newContentMarshaler() runtime.Marshaler {
	return &contentMarshaler{new(runtime.JSONPb)}
}

func (cm *contentMarshaler) Marshal(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case *pb.ContentResponse:
		return []byte(t.Content), nil

	default:
		return cm.JSONPb.Marshal(v)
	}
}

// ContentType is always "text/plain"
func (*contentMarshaler) ContentType() string { return "text/plain" }
