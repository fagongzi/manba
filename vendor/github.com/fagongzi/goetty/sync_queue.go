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

package goetty

import (
	"fmt"
	"sync"
)

// OffsetQueue is a queue for sync.
type OffsetQueue struct {
	sync.Mutex

	start, end uint64
	items      []interface{}
}

func newOffsetQueue() *OffsetQueue {
	return &OffsetQueue{}
}

// Add add a item to the queue
func (q *OffsetQueue) Add(item interface{}) uint64 {
	q.Lock()
	q.end++
	q.items = append(q.items, item)
	max := q.getMaxOffset0()
	q.Unlock()
	return max
}

// Get returns all the items after the offset, and remove all items before this offset
func (q *OffsetQueue) Get(offset uint64) ([]interface{}, uint64) {
	oldOffset := offset
	if offset > 0 {
		offset = offset - 1
	}

	q.Lock()
	max := q.getMaxOffset0()
	if offset > max {
		panic(fmt.Sprintf("bug: error offset %d, end is %d", offset, q.end))
	} else if offset < q.start || (oldOffset == 0 && offset == q.start && q.start == 0) {
		value := q.items[0:]
		q.Unlock()
		return value, max
	}

	var value []interface{}
	for i := q.start; i < q.end; i++ {
		if i <= offset {
			q.items[i-q.start] = nil
		} else {
			value = append(value, q.items[i-q.start])
		}
	}

	old := q.start
	q.start = offset + 1
	if q.start < q.end {
		q.items = q.items[q.start-old:]
	} else {
		q.items = make([]interface{}, 0, 0)
	}
	q.Unlock()

	return value, max
}

// GetMaxOffset returns the max offset in the queue
func (q *OffsetQueue) GetMaxOffset() uint64 {
	q.Lock()
	v := q.getMaxOffset0()
	q.Unlock()
	return v
}

func (q *OffsetQueue) getMaxOffset0() uint64 {
	if q.end == 0 {
		return 0
	}

	return q.end
}
