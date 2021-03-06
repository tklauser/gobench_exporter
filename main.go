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

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/tklauser/gobench_exporter/bench"
	"github.com/tklauser/gobench_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TODO: pass base directory of Go module, invoke `go test -bench` upon trigger (HTTP), report
// metrics

type handler struct {
	repoPath  string
	collector *collector.GoBenchCollector
}

func newHandler(repoPath string, collector *collector.GoBenchCollector) *handler {
	return &handler{
		repoPath:  repoPath,
		collector: collector,
	}
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	goArgs := []string{"test", "-run=_NONE_", "-bench=."}
	gocheckArgs := []string{"test", "-check.b", "-check.bmem"}

	for _, args := range [][]string{goArgs, gocheckArgs} {
		cmd := exec.Command("go", args...)
		cmd.Dir = h.repoPath
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Failed to get command stdout: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Running go command %v", cmd)
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start command: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bs, err := bench.ParseSet(stdout)
		if err == nil {
			log.Print(bs)
			w.Write([]byte(fmt.Sprintf("%v\n", bs)))
		}

		if err := cmd.Wait(); err != nil {
			log.Printf("Failed to wait for command to exit: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics.",
		).Default(":9777").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		triggerPath = kingpin.Flag(
			"web.trigger-path",
			"Path under which to trigger benchmarks.",
		).Default("/trigger").String()
		repoPath = kingpin.Flag(
			"fs.repo-path",
			"Filesystem path of the Go package to benchmark.",
		).Default(".").String()
	)

	kingpin.Version(version.Print("gobench_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Printf("Starting gobench_exporter version %s", version.Info())
	log.Printf("Benchmarking Go packages in directory %s", *repoPath)

	c := collector.NewGoBenchCollector(os.Stdin)
	if err := prometheus.Register(c); err != nil {
		log.Fatalf("Failed to register collector: %v", err)
	}
	h := newHandler(*repoPath, c)

	http.Handle(*metricsPath, promhttp.Handler())
	http.Handle(*triggerPath, h)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Go Benchmark Exporter</title></head>
			<body>
			<h1>Go Benchmark Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<p><a href="` + *triggerPath + `">Trigger benchmarks</a></p>
			</body>
			</html>`))
	})

	log.Printf("Listening on %s", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error listening on %s: %s", *listenAddress, err)
		os.Exit(1)
	}
}
