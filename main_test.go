// Copyright 2020 Tobias Klauser

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

package main

import (
	"sort"
	"testing"

	"gopkg.in/check.v1"
)

// BenchmarkSortSlice is an example benchmarking sort.Slice using the testing package.
func BenchmarkSortSlice(b *testing.B) {
	data := make([]int, 1<<10)
	b.ResetTimer() // ignore big allocation
	for i := 0; i < b.N; i++ {
		for i := range data {
			data[i] = i ^ 0x2cc
		}
		b.StartTimer()
		sort.Slice(data, func(i, j int) bool {
			return data[i] < data[j]
		})
		b.StopTimer()
	}

}

func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

// BenchmarkSortSlice is an examplke benchmarking sort.Slice using the gocheck package.
func (s *MySuite) BenchmarkSortSlice(c *check.C) {
	data := make([]int, 1<<10)
	c.ResetTimer() // ignore big allocation
	for i := 0; i < c.N; i++ {
		for i := range data {
			data[i] = i ^ 0x2cc
		}
		c.StartTimer()
		sort.Slice(data, func(i, j int) bool {
			return data[i] < data[j]
		})
		c.StopTimer()
	}
}
