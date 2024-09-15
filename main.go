/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"

	"github.com/nacioboi/go_sfda_map/sfda_map"
	tests "github.com/nacioboi/go_sfda_map/tests"
)

// A func allowing us to see memory usage in the task manager:
// func wait_for_enter() {
// 	fmt.Println("Press enter to exit...")
// 	var s string
// 	fmt.Scanln(&s)
// }

func main() {
	//
	// Setups...
	//
	runtime.SetCPUProfileRate(8000)

	tests.Test_Consistency(1024)

	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)
	defer runtime.GC()

	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n1 := uint64(1024 * 1024 * 32)
	n2 := n1
	n3 := uint64(1024)
	n4 := n1

	// Create maps...
	b_m_1 := make(map[uint64]uint64)
	sfda_map_1 := sfda_map.New_SFDA_Map[uint64, uint64](
		n1,
		sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__1_ENTRIES_PER_BUCKET),
	)
	b_m_2 := make(map[uint64]uint64)
	sfda_map_2 := sfda_map.New_SFDA_Map[uint64, uint64](
		n2,
		sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__1_ENTRIES_PER_BUCKET),
	)
	sfda_r_m := sfda_map.New_SFDA_Resizable_Map[uint64, uint64](
		16,
		sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__1_ENTRIES_PER_BUCKET),
	)

	//
	// Linear benchmarks...
	//

	// Benchmark built-in map - linear...
	runtime.GC()
	tests.Bench_Linear_Builtin_Map_Set(b_m_1, n1, true)
	tests.Bench_Linear_Builtin_Map_Get(b_m_1, n1, true)

	// Benchmark SFDA map - linear...
	runtime.GC()
	tests.Bench_Linear_SFDA_Map_Set(sfda_map_1, n1, true)
	tests.Bench_Linear_SFDA_Map_Get(sfda_map_1, n1, true)

	//
	// Random benchmarks...
	//
	fmt.Println("")

	// Generate a random keys array...
	random_keys := tests.Generate_Random_Keys(n2)

	// Benchmark built-in map - random...
	runtime.GC()
	tests.Bench_Linear_Builtin_Map_Set(b_m_2, n2, false)
	microseconds := tests.Bench_Random_Builtin_Map_Get(b_m_2, random_keys, false)
	time_per_op := float64(microseconds) / float64(n2)
	fmt.Printf("Built-in Map Microseconds  ::: RANDOM GET PER OP ::: %f\n", time_per_op)

	// Benchmark SFDA map - random...
	runtime.GC()
	tests.Bench_Linear_SFDA_Map_Set(sfda_map_2, n2, false)
	//pprof.StartCPUProfile(f)
	microseconds = tests.Bench_Random_SFDA_Map_Get(sfda_map_2, random_keys, false)
	//pprof.StopCPUProfile()
	time_per_op = float64(microseconds) / float64(n2)
	fmt.Printf("SFDA Map Microseconds      ::: RANDOM GET PER OP ::: %f\n", time_per_op)

	//
	// Deletion benchmarks...
	//
	fmt.Println("")

	// Benchmark built-in map - delete...
	runtime.GC()
	tests.Bench_Deletion_Builtin_Map(b_m_2, random_keys)

	// Benchmark SFDA map - delete...
	runtime.GC()
	tests.Bench_Deletion_SFDA_Map(sfda_map_2, random_keys)

	//
	// Memory usage benchmarks...
	//
	fmt.Println("")

	// Benchmark built-in map - memory usage...
	runtime.GC()
	tests.Bench_Mem_Usage_Builtin_Map(n3)

	// Benchmark SFDA map - memory usage...
	runtime.GC()
	tests.Bench_Mem_Usage_SFDA_Map(n3)

	//
	// SFDA resizable map benchmarks...
	//
	fmt.Println("")

	// Benchmark SFDA resizable map...
	runtime.GC()
	pprof.StartCPUProfile(f)
	tests.Bench_Linear_SFDA_Resizable_Map_Set(sfda_r_m, n4, true)
	pprof.StopCPUProfile()
	tests.Bench_Linear_SFDA_Resizable_Map_Get(sfda_r_m, n4, true)

	//wait_for_enter()
}
