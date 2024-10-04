package perf

/*
#cgo CFLAGS: -I.
#include <stdio.h>
#include <stdlib.h>
#include "perf_api.h"
#include "perf_api.c"

*/
import "C"
import (
	_ "fmt"
	"log"
	"unsafe"
)

func UNUSED(x ...interface{}) {}

type PerfType C.uint64_t

const (
	PERF_TYPE_HARDWARE PerfType = C.PERF_TYPE_HARDWARE
	PERF_TYPE_HW_CACHE PerfType = C.PERF_TYPE_HW_CACHE
	PERF_TYPE_RAW      PerfType = C.PERF_TYPE_RAW
)

// PerfEventConfig holds the configuration for performance events
type PerfEventConfig struct {
	pe      C.struct_perf_event_attr
	Fds     []C.int
	Ids     []C.uint64_t
	Types   []C.uint64_t
	Configs []C.uint64_t
}

// NewPerfEventConfig initializes a new PerfEventConfig with the specified number of events
func NewPerfEventConfig(numEvents int) PerfEventConfig {
	return PerfEventConfig{
		Fds:     make([]C.int, numEvents),
		Ids:     make([]C.uint64_t, numEvents),
		Types:   make([]C.uint64_t, numEvents),
		Configs: make([]C.uint64_t, numEvents),
	}
}

func ConfigPerf(pec *PerfEventConfig, cpu int) C.int {
	group_fd := C.config_perf_multi(
		(*C.struct_perf_event_attr)(unsafe.Pointer(&pec.pe)), // Assuming `pe` is correctly initialized
		&pec.Fds[0],
		&pec.Ids[0],
		&pec.Types[0],
		&pec.Configs[0],
		(C.int)(len(pec.Fds)),
		(C.int)(cpu),
	)
	return group_fd
}

func StartInstrumentation(group_fd C.int) {
	C.reset_and_enable_ioctl(group_fd)
}

func EndInstrumentation(pe *C.struct_perf_event_attr, group_fd C.int, cStrings []*C.char, peConfig *PerfEventConfig, numEvents int) {

	// Disable and read the event
	C.disable_ioctl(group_fd)

	C.get_perf(
		(*C.struct_perf_event_attr)(unsafe.Pointer(&pe)),
		&cStrings[0],
		&peConfig.Ids[0],
		(C.int)(numEvents),
		group_fd,
	)
}

// TODO: configure for PERF_TYPE_RAW events
func SetupPerfEvents(peConfig *PerfEventConfig,
	configs []C.uint64_t,
	types []C.uint64_t,
	hw_cache_op_ids []C.uint64_t,
	hw_cache_op_result_ids []C.uint64_t,
	numEvents int) {

	cache_id_count := 0
	var cache_hw_event C.uint64_t

	if len(hw_cache_op_ids) != len(hw_cache_op_result_ids) {
		log.Fatal("hw_cache_op_ids must have the same elements as hw_cache_op_result_ids!")
		return
	}
	// hw_cache_op_ids can be empty if PERF_TYPE_HW_CACHE is not specified
	for i := 0; i < numEvents; i++ {
		if types[i] == C.PERF_TYPE_HW_CACHE {
			//uint64_t config_cache_id(uint64_t perf_hw_cache_id, uint64_t perf_hw_cache_op_id, uint64_t perf_hw_cache_op_result_id)
			if cache_id_count > len(hw_cache_op_ids) {
				log.Fatal("not enough hw_cache_op_ids!")
				return
			}
			cache_hw_event = C.config_cache_id(configs[i],
				hw_cache_op_ids[cache_id_count],
				hw_cache_op_result_ids[cache_id_count])

			cache_id_count++

			peConfig.Configs[i] = cache_hw_event
		} else {
			peConfig.Configs[i] = configs[i]
		}
		peConfig.Fds[i] = -1
		peConfig.Types[i] = types[i]
		peConfig.Ids[i] = C.uint64_t(i)
	}
}

/*
We are concerned about:
1. Executed Instructions (0x8)
2. L1D Accesses (0x4)
3. L1D Hits/Misses
4. L2 Accesses
5. L2 Hits/Misses
*/
func main() {
	// we shall test perf here

	types := []C.uint64_t{C.PERF_TYPE_HARDWARE,
		C.PERF_TYPE_HARDWARE,
		C.PERF_TYPE_HW_CACHE,
		C.PERF_TYPE_RAW,
	}

	configs := []C.uint64_t{C.PERF_COUNT_HW_INSTRUCTIONS,
		C.PERF_COUNT_HW_CPU_CYCLES,
		C.PERF_COUNT_HW_CACHE_L1D,
		0x17,
	}

	hw_cache_op_ids := []C.uint64_t{C.PERF_COUNT_HW_CACHE_OP_READ}
	hw_cache_op_result_ids := []C.uint64_t{C.PERF_COUNT_HW_CACHE_RESULT_ACCESS}

	numEvents := len(configs)

	peConfig := NewPerfEventConfig(numEvents)

	SetupPerfEvents(&peConfig, configs, types, hw_cache_op_ids, hw_cache_op_result_ids, numEvents)

	strings := []string{"Instructions", "Cycles", "L1D Cache Read Accesses", "L2_CACHE_RD"}
	// Create a slice to hold the C strings
	cStrings := make([]*C.char, len(strings))

	for i, s := range strings {
		cStrings[i] = C.CString(s)                // Convert Go string to C string
		defer C.free(unsafe.Pointer(cStrings[i])) // Free memory when done
	}
	group_fd := ConfigPerf(&peConfig, 0)

	if int(group_fd) < 0 {
		log.Fatal("Error setting group_fd")
		return
	}

	StartInstrumentation(group_fd)

	C.something()

	EndInstrumentation(&peConfig.pe, group_fd, cStrings, &peConfig, numEvents)

}
