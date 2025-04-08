package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	mrand "math/rand"
	"os"
	"os/signal"
	"runtime"
	"time"
)

const (
	defaultAllocSizeBytes            = 1024 * 1024 // 1 MiB
	defaultAllocIntervalMilliseconds = 250
	defaultMaxAllocBytes             = 1024 * 1024 * 1024 // 1 GiB
)

func main() {

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(os.Stdout, nil),
		),
	)

	allocSizeBytes := flag.Int(
		"allocSizeBytes",
		defaultAllocSizeBytes,
		fmt.Sprintf(
			"The allocation size in bytes (default: %d).",
			defaultAllocSizeBytes),
	)

	allocIntervalMilliseconds := flag.Int(
		"allocIntervalMilliseconds",
		defaultAllocIntervalMilliseconds,
		fmt.Sprintf(
			"The delay between each allocation/log cycle in seconds (default: %d).",
			defaultAllocIntervalMilliseconds,
		),
	)

	maxAllocBytes := flag.Int64(
		"maxAllocBytes",
		defaultMaxAllocBytes,
		fmt.Sprintf(
			"The maximum number of bytes to allocate before exiting the program (default: %d).",
			defaultMaxAllocBytes,
		),
	)

	flag.Parse()

	os.Exit(run(options{
		allocSizeBytes: *allocSizeBytes,
		allocInterval:  time.Duration(*allocIntervalMilliseconds) * time.Millisecond,
		maxAllocBytes:  *maxAllocBytes,
	}))
}

type options struct {
	allocSizeBytes int
	allocInterval  time.Duration
	maxAllocBytes  int64
}

func run(opts options) int {

	slog.Info("Starting with options.",
		"AllocSizeInBytes", opts.allocSizeBytes,
		"AllocInterval", opts.allocInterval,
		"MaxAllocBytes", opts.maxAllocBytes)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	maxAllocs := int(opts.maxAllocBytes / int64(opts.allocSizeBytes))
	arr := make([][]byte, 0, maxAllocs)
	var totalAllocsBytes int64

	// If we don't write anything to the allocated memory then Docker / K8s don't seem to count it
	// towards utilization in the output of `docker stats` and `kubectl top pods`. Writing at
	// allocation time has the desired effect when there are no memory constraints but running in
	// k8s with memory limits something (paging?) keeps the pod within the defined limits. Writing
	// frequently and randomly results in the memory metrics rising as expected.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond):
				// fallthrough
			}
			val := byte(mrand.Intn(256))
			buf := arr[mrand.Intn(len(arr))]
			buf[mrand.Intn(len(buf))] = val
		}
	}()

	for {

		allocSizeBytes, lastAlloc := func() (int, bool) {
			if totalAllocsBytes+int64(opts.allocSizeBytes) >= opts.maxAllocBytes {
				return int(opts.maxAllocBytes - totalAllocsBytes), true
			}
			return opts.allocSizeBytes, false
		}()

		arr = append(arr, make([]byte, allocSizeBytes))
		totalAllocsBytes += int64(allocSizeBytes)

		memStats := func() runtime.MemStats {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return m
		}()
		slog.Info("Memory stats.",
			"MostRecentAllocBytes", allocSizeBytes,
			"TotalAllocBytes", totalAllocsBytes,
			"IsLastAlloc", lastAlloc,
			"MemStats.Alloc", memStats.Alloc,
			"MemStats.TotalAlloc", memStats.TotalAlloc,
			"MemStats.Sys", memStats.Sys,
			"MemStats.Mallocs", memStats.Mallocs,
		)

		if lastAlloc {
			break
		}

		select {
		case sig := <-sigs:
			slog.Warn("Exit signal received.", "signal", sig)
			os.Exit(1)
		case <-time.After(opts.allocInterval):
			// continue after delay
		}
	}

	slog.Info("Max total allocations reached.", "max_alloc_bytes", opts.maxAllocBytes)
	return 0
}
