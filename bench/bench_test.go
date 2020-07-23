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

package bench_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tklauser/gobench_exporter/bench"
	"golang.org/x/tools/benchmark/parse"
)

func TestParseLine(t *testing.T) {
	benchs := []struct {
		line    string
		want    *parse.Benchmark
		wantErr bool
	}{
		{
			line: "BenchmarkSortSlice-8   	   17461	     69022 ns/op",
			want: &parse.Benchmark{
				Name:     "BenchmarkSortSlice-8",
				N:        17461,
				NsPerOp:  69022,
				Measured: parse.NsPerOp,
			},
		},
		{
			line: "BenchmarkSortSlice-8   	   16572	     78922 ns/op	      64 B/op	       2 allocs/op",
			want: &parse.Benchmark{
				Name:              "BenchmarkSortSlice-8",
				N:                 16572,
				NsPerOp:           78922.0,
				AllocedBytesPerOp: 64,
				AllocsPerOp:       2,
				Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
			},
		},
		{
			line: "PASS: main_test.go:49: MySuite.BenchmarkSortSlice	   20000	     89618 ns/op",
			want: &parse.Benchmark{
				Name:     "MySuite.BenchmarkSortSlice",
				N:        20000,
				NsPerOp:  89618.0,
				Measured: parse.NsPerOp,
			},
		},
		{
			line: "PASS: main_test.go:49: MySuite.BenchmarkSortSlice	   20000	     90444 ns/op	 64 B/op	       2 allocs/op",
			want: &parse.Benchmark{
				Name:              "MySuite.BenchmarkSortSlice",
				N:                 20000,
				NsPerOp:           90444.0,
				AllocedBytesPerOp: 64,
				AllocsPerOp:       2,
				Measured:          parse.NsPerOp | parse.AllocedBytesPerOp | parse.AllocsPerOp,
			},
		},
		{
			line: "PASS: main_test.go:1: MySuite.BenchmarkFoobar 90000",
			want: &parse.Benchmark{
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
