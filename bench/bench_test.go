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

package bench_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tklauser/gobench_exporter/bench"
)

func TestParseLine(t *testing.T) {
	benchs := []struct {
		line    string
		want    *bench.Benchmark
		wantErr bool
	}{
		{
			line: "BenchmarkSortSlice-8   	   17461	     69022 ns/op",
			want: &bench.Benchmark{
				Name:     "BenchmarkSortSlice-8",
				N:        17461,
				NsPerOp:  69022,
				Measured: bench.NsPerOp,
			},
		},
		{
			line: "BenchmarkSortSlice-8   	   16572	     78922 ns/op	      64 B/op	       2 allocs/op",
			want: &bench.Benchmark{
				Name:              "BenchmarkSortSlice-8",
				N:                 16572,
				NsPerOp:           78922.0,
				AllocedBytesPerOp: 64,
				AllocsPerOp:       2,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
			},
		},
		{
			line: "PASS: main_test.go:49: MySuite.BenchmarkSortSlice	   20000	     89618 ns/op",
			want: &bench.Benchmark{
				Name:     "MySuite.BenchmarkSortSlice",
				N:        20000,
				NsPerOp:  89618.0,
				Measured: bench.NsPerOp,
			},
		},
		{
			line: "PASS: main_test.go:49: MySuite.BenchmarkSortSlice	   20000	     90444 ns/op	 64 B/op	       2 allocs/op",
			want: &bench.Benchmark{
				Name:              "MySuite.BenchmarkSortSlice",
				N:                 20000,
				NsPerOp:           90444.0,
				AllocedBytesPerOp: 64,
				AllocsPerOp:       2,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
			},
		},
		{
			line: "PASS: main_test.go:1: MySuite.BenchmarkFoobar 90000",
			want: &bench.Benchmark{
				Name: "MySuite.BenchmarkFoobar",
				N:    90000,
			},
		},
		{
			line: "\t\tPASS: main_test.go:1: MySuite.BenchmarkFoobar 90000\n",
			want: &bench.Benchmark{
				Name: "MySuite.BenchmarkFoobar",
				N:    90000,
			},
		},
		{
			line:    "Benchmark",
			wantErr: true,
		},
		{
			line:    "PASS:",
			wantErr: true,
		},
		{
			line:    "PASS: main_test.go:1: Foobar",
			wantErr: true,
		},
		{
			line:    "PASS: main_test.go:1: MySuite.BenchmarkFoobar",
			wantErr: true,
		},
		{
			line:    "foobar",
			wantErr: true,
		},
	}

	for _, b := range benchs {
		got, err := bench.ParseLine(b.line)
		if !b.wantErr && err != nil {
			t.Errorf("ParseLine(%s): %v", b.line, err)
		} else if b.wantErr && err == nil {
			t.Errorf("ParseLine(%s): want an error, got nil", b.line)
		}

		if diff := cmp.Diff(b.want, got); diff != "" {
			t.Errorf("ParseLine(%s) [-want +got]:\n%s", b.line, diff)
		}
	}
}

func TestParseSet(t *testing.T) {
	// Output involving go testing.B and gocheck benchmarks with noise inbetween. Test that
	// benchmarks with the same name (e.g. from multiple runs) have their order preserved.
	in := `
		PASS: idpool_test.go:296: IDPoolTestSuite.BenchmarkLeaseIDs	 5000000	       520 ns/op	      80 B/op	       3 allocs/op
		PASS: idpool_test.go:286: IDPoolTestSuite.BenchmarkRemoveIDs	 5000000	       394 ns/op	      80 B/op	       3 allocs/op
		PASS: idpool_test.go:307: IDPoolTestSuite.BenchmarkUseAndRelease	 1000000	      4634 ns/op	     160 B/op	       6 allocs/op
		OK: 3 passed
		PASS: idpool_test.go:296: IDPoolTestSuite.BenchmarkLeaseIDs	 5000000	       517 ns/op	      80 B/op	       3 allocs/op
		PASS: idpool_test.go:286: IDPoolTestSuite.BenchmarkRemoveIDs	 5000000	       400 ns/op	      80 B/op	       3 allocs/op
		PASS: idpool_test.go:307: IDPoolTestSuite.BenchmarkUseAndRelease	 1000000	      2844 ns/op	     160 B/op	       6 allocs/op
		OK: 3 passed
		PASS
		ok  	github.com/cilium/cilium/pkg/idpool	21.324s
		goos: linux
		goarch: amd64
		pkg: github.com/cilium/cilium/pkg/labels
		BenchmarkParseLabel-8   	 2032945	       569 ns/op
		BenchmarkParseLabel-8   	 2042311	       557 ns/op
		PASS
		ok  	github.com/cilium/cilium/pkg/labels	3.217s
	`

	want := bench.Set{
		"IDPoolTestSuite.BenchmarkLeaseIDs": []*bench.Benchmark{
			{
				Name:              "IDPoolTestSuite.BenchmarkLeaseIDs",
				N:                 5000000,
				NsPerOp:           520,
				AllocedBytesPerOp: 80,
				AllocsPerOp:       3,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
				Ord:               0,
			},
			{
				Name:              "IDPoolTestSuite.BenchmarkLeaseIDs",
				N:                 5000000,
				NsPerOp:           517,
				AllocedBytesPerOp: 80,
				AllocsPerOp:       3,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
				Ord:               3,
			},
		},
		"IDPoolTestSuite.BenchmarkRemoveIDs": []*bench.Benchmark{
			{
				Name:              "IDPoolTestSuite.BenchmarkRemoveIDs",
				N:                 5000000,
				NsPerOp:           394,
				AllocedBytesPerOp: 80,
				AllocsPerOp:       3,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
				Ord:               1,
			},
			{
				Name:              "IDPoolTestSuite.BenchmarkRemoveIDs",
				N:                 5000000,
				NsPerOp:           400,
				AllocedBytesPerOp: 80,
				AllocsPerOp:       3,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
				Ord:               4,
			},
		},
		"IDPoolTestSuite.BenchmarkUseAndRelease": []*bench.Benchmark{
			{
				Name:              "IDPoolTestSuite.BenchmarkUseAndRelease",
				N:                 1000000,
				NsPerOp:           4634,
				AllocedBytesPerOp: 160,
				AllocsPerOp:       6,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
				Ord:               2,
			},
			{
				Name:              "IDPoolTestSuite.BenchmarkUseAndRelease",
				N:                 1000000,
				NsPerOp:           2844,
				AllocedBytesPerOp: 160,
				AllocsPerOp:       6,
				Measured:          bench.NsPerOp | bench.AllocedBytesPerOp | bench.AllocsPerOp,
				Ord:               5,
			},
		},
		"BenchmarkParseLabel-8": []*bench.Benchmark{
			{
				Name:     "BenchmarkParseLabel-8",
				N:        2032945,
				NsPerOp:  569,
				Measured: bench.NsPerOp,
				Ord:      6,
			},
			{
				Name:     "BenchmarkParseLabel-8",
				N:        2042311,
				NsPerOp:  557,
				Measured: bench.NsPerOp,
				Ord:      7,
			},
		},
	}

	got, err := bench.ParseSet(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ParseSet: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseSet [-want +got]:\n%s", diff)
	}
}
