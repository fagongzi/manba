// Copyright 2016 DeepFabric, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package protoc

import (
	"log"
	"runtime"
)

// PB pb interface
type PB interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (int, error)
	Unmarshal(data []byte) error
}

// MustUnmarshal if unmarshal failed, will panic
func MustUnmarshal(pb PB, data []byte) {
	err := pb.Unmarshal(data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("pb unmarshal failed, data=<%v> errors:\n %+v \n %s",
			data,
			err,
			buf)
	}
}

// MustMarshal if marsh failed, will panic
func MustMarshal(pb PB) []byte {
	data, err := pb.Marshal()
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("pb marshal failed, pb=<%+v> errors:\n %+v \n %s",
			pb,
			err,
			buf)
	}

	return data
}

// MustMarshalTo if marsh failed, will panic
func MustMarshalTo(pb PB, data []byte) int {
	n, err := pb.MarshalTo(data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("pb marshal failed, pb=<%v> errors:\n %+v \n %s",
			pb,
			err,
			buf)
	}

	return n
}
