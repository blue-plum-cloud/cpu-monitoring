# Introduction
CPU Monitoring functions in Go using the syscall and Intel PMU Sensor Server

# How to Use
NOTE: `perf-api` requires root permissions as it uses the `perf_event_open` system call.
TODO

# Future Work
Make the perf syscall a Go implementation instead of cGo to reduce overhead.
