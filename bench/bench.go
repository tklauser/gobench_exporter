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

package bench

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/tools/benchmark/parse"
)

// Flags used by BenchmarkStat.Measured to indicate
// which measurements a Benchmark contains.
const (
	NsPerOp = 1 << iota
	MBPerS
	AllocedBytesPerOp
	AllocsPerOp
)

// Benchmark is one run of a single benchmark. Based on x/tools/benchmark/parse.Benchmark.
type Benchmark struct {
	Name              string  // benchmark name
	N                 int     // number of iterations
	NsPerOp           float64 // nanoseconds per iteration
	AllocedBytesPerOp uint64  // bytes allocated per iteration
	AllocsPerOp       uint64  // allocs per iteration
	MBPerS            float64 // MB processed per second
	Measured          int     // which measurements were recorded
	Ord               int     // ordinal position within a benchmark run
}

// parseGoCheckLine extracts a parse.Benchmark from a single line of benchmark output as emitted by
// gopkg.in/check.v1 (https://labix.org/gocheck).
// Based on
// https://github.com/golang/tools/blob/a7c6fd066f6dcf64c13983e28e029ce7874760ff/benchmark/parse/parse.go#L41
func parseGoCheckLine(line string) (*Benchmark, error) {
	// line format:
	// PASS: main_test.go:48: MySuite.BenchmarkSortSlice	   20000	     90444 ns/op	      64 B/op	       2 allocs/op
	fields := strings.Fields(line)

	// Four required positional fields: PASS, file/line, benchmark name, iterations
	if len(fields) < 4 {
		return nil, fmt.Errorf("four fields required, have %d", len(fields))
	}
	if fields[0] != "PASS:" {
		return nil, fmt.Errorf("gocheck benchmark did not pass")
	}
	if !strings.Contains(fields[2], "Benchmark") {
		return nil, fmt.Errorf("not a gocheck benchmark")
	}
	n, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, err
	}
	b := &Benchmark{Name: fields[2], N: n}

	// Parse any remaining pairs of fields; we've parsed one pair already.
	for i := 1; i < len(fields)/2; i++ {
		// based on
		// https://github.com/golang/tools/blob/a7c6fd066f6dcf64c13983e28e029ce7874760ff/benchmark/parse/parse.go#L64
		quant, unit := fields[i*2], fields[i*2+1]
		switch unit {
		case "ns/op":
			if f, err := strconv.ParseFloat(quant, 64); err == nil {
				b.NsPerOp = f
				b.Measured |= parse.NsPerOp
			}
		case "MB/s":
			if f, err := strconv.ParseFloat(quant, 64); err == nil {
				b.MBPerS = f
				b.Measured |= parse.MBPerS
			}
		case "B/op":
			if i, err := strconv.ParseUint(quant, 10, 64); err == nil {
				b.AllocedBytesPerOp = i
				b.Measured |= parse.AllocedBytesPerOp
			}
		case "allocs/op":
			if i, err := strconv.ParseUint(quant, 10, 64); err == nil {
				b.AllocsPerOp = i
				b.Measured |= parse.AllocsPerOp
			}
		}
	}

	return b, nil
}

// ParseLine extracts a Benchmark from a single line of testing.B or check.C benchmark output.
func ParseLine(line string) (*Benchmark, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "Benchmark") {
		// Go standard library testing format
		b, err := parse.ParseLine(line)
		if err != nil {
			return nil, err
		}
		return &Benchmark{
			Name:              b.Name,
			N:                 b.N,
			NsPerOp:           b.NsPerOp,
			AllocedBytesPerOp: b.AllocedBytesPerOp,
			AllocsPerOp:       b.AllocsPerOp,
			MBPerS:            b.MBPerS,
			Measured:          b.Measured,
			Ord:               b.Ord,
		}, err
	} else if strings.HasPrefix(line, "PASS:") && strings.Contains(line, "Benchmark") {
		return parseGoCheckLine(line)
	}
	return nil, fmt.Errorf("not a valid benchmark line")
}

// Set is a collection of benchmarks from one testing.B or gocheck.C benchmark run, keyed by name to
// facilitate comparison.
// Based on x/tools/benchmark/parse.Set.
type Set map[string][]*Benchmark

// ParseSet extracts a Set from testing.B or check.C benchmark output.
// ParseSet preserves the order of benchmarks that have identical
// names.
func ParseSet(r io.Reader) (Set, error) {
	bb := make(Set)
	scan := bufio.NewScanner(r)
	ord := 0
	for scan.Scan() {
		if b, err := ParseLine(scan.Text()); err == nil {
			b.Ord = ord
			ord++
			bb[b.Name] = append(bb[b.Name], b)
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return bb, nil
}
