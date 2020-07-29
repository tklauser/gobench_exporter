// Copyright 2020 Isovalent, Inc

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

package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tklauser/gobench_exporter/bench"
)

// namespace is the common namespace to be used by all metrics.
const namespace = "gobench"

// GoBenchCollector implements the prometheus.GoBenchCollector interface.
type GoBenchCollector struct {
	benchmarks     bench.Set
	benchmarksDesc *prometheus.Desc
}

// NewGoBenchCollector
func NewGoBenchCollector() *GoBenchCollector {
	// for now just return static benchmarks
	in := `
	BenchmarkSortSlice-8   	   16315	     68857 ns/op
	PASS: main_test.go:48: MySuite.BenchmarkSortSlice	   20000	     81293 ns/op	      64 B/op	       2 allocs/op
	`

	c := &GoBenchCollector{}
	b, err := bench.ParseSet(strings.NewReader(in))
	if err == nil {
		c.benchmarks = b
		c.benchmarksDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "go_benchmarks"),
			"The set of Go benchmarks",
			[]string{"name"},
			nil,
		)
	}
	return c
}

// Describe implements prometheus.Collector interface.
func (e *GoBenchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.benchmarksDesc
}

// Collect implements prometheus.Collector interface and sends all metrics.
func (e *GoBenchCollector) Collect(ch chan<- prometheus.Metric) {
	for _, bb := range e.benchmarks {
		for _, b := range bb {
			ch <- prometheus.MustNewConstMetric(e.benchmarksDesc, prometheus.GaugeValue, 1, b.Name)
		}
	}
}
