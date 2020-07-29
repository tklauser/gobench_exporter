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
	benchmarks         bench.Set
	benchmarkNamesDesc *prometheus.Desc
	benchmarkDescs     map[string]*prometheus.Desc
}

var qtys = []string{"N", "ns/op", "B/op", "allocs/op", "MB/s"}

func validPrometheusMetricName(r rune) rune {
	// see https://github.com/prometheus/common/blob/546f1fd8d7df61d94633b254641f9f8f48248ada/model/metric.go#L92
	switch {
	case 'a' <= r && r <= 'z',
		'A' <= r && r <= 'Z',
		'0' <= r && r <= '9',
		r == '_', r == ':':
		return r
	default:
		return '_'
	}
}

// NewGoBenchCollector
func NewGoBenchCollector() *GoBenchCollector {
	// FIXME: for now just return static benchmarks
	in := `
	BenchmarkSortSlice-8   	   16315	     68857 ns/op
	PASS: main_test.go:48: MySuite.BenchmarkSortSlice	   20000	     81293 ns/op	      64 B/op	       2 allocs/op
	`

	c := &GoBenchCollector{}
	bs, err := bench.ParseSet(strings.NewReader(in))
	if err == nil {
		c.benchmarks = bs
		c.benchmarkNamesDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "benchmarks"),
			"The set of Go benchmarks",
			[]string{"name"},
			nil,
		)
		c.benchmarkDescs = make(map[string]*prometheus.Desc, len(bs))
		for _, bb := range bs {
			for _, b := range bb {
				name := strings.Map(validPrometheusMetricName, b.Name)
				for _, qty := range qtys {
					key := b.Name + qty
					if _, ok := c.benchmarkDescs[key]; !ok {
						c.benchmarkDescs[key] = prometheus.NewDesc(
							prometheus.BuildFQName(namespace, "", name+strings.ReplaceAll(qty, "/", "_per_")),
							b.Name+" "+qty,
							nil,
							nil,
						)
					}
				}
			}
		}
	}
	return c
}

// Describe implements prometheus.Collector interface.
func (e *GoBenchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.benchmarkNamesDesc
}

// Collect implements prometheus.Collector interface and sends all metrics.
func (e *GoBenchCollector) Collect(ch chan<- prometheus.Metric) {
	for _, bb := range e.benchmarks {
		for _, b := range bb {
			ch <- prometheus.MustNewConstMetric(e.benchmarkNamesDesc, prometheus.GaugeValue, 1, b.Name)

			ch <- prometheus.MustNewConstMetric(e.benchmarkDescs[b.Name+qtys[0]], prometheus.GaugeValue, float64(b.N))
			ch <- prometheus.MustNewConstMetric(e.benchmarkDescs[b.Name+qtys[1]], prometheus.GaugeValue, b.NsPerOp)
			ch <- prometheus.MustNewConstMetric(e.benchmarkDescs[b.Name+qtys[2]], prometheus.GaugeValue, float64(b.AllocedBytesPerOp))
			ch <- prometheus.MustNewConstMetric(e.benchmarkDescs[b.Name+qtys[3]], prometheus.GaugeValue, float64(b.AllocsPerOp))
			ch <- prometheus.MustNewConstMetric(e.benchmarkDescs[b.Name+qtys[4]], prometheus.GaugeValue, b.MBPerS)
		}
	}
}
