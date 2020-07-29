# Go benchmark exporter

Prometheus exporter for Go benchmark results.

```
# Export Go stdlib testing package benchmarks
$ go test -run=_NONE_ -bench=. | ./gobench_exporter
# Export gocheck benchmarks
$ go test -check.b -check.bmem | ./gobench_exporter
```
