/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map_tests

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/nacioboi/go_sfda_map/sfda_map"
)

var t uint64
var start time.Time

type Test_Result struct {
	Elapsed_Time int64
	Checksum     uint64
	Memory_Usage uint64
}

func Bench_Linear_Builtin_Map_Set(builtin_map map[uint64]uint64, n uint64) Test_Result {
	start = time.Now()
	for i := uint64(0); i < n; i++ {
		builtin_map[i+1] = i
	}
	since := time.Since(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
	}
}

func Bench_Linear_Builtin_Map_Get(builtin_map map[uint64]uint64, n uint64) Test_Result {
	t = 0
	start = time.Now()
	for i := uint64(0); i < n; i++ {
		x := builtin_map[i+1]
		t += x
	}
	since := time.Since(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
		Checksum:     t,
	}
}

func Bench_Linear_SFDA_Map_Set(sfda *sfda_map.SFDA_Map[uint64, uint64], n uint64) Test_Result {
	start = time.Now()
	for i := uint64(0); i < n; i++ {
		sfda.Set(i+1, i)
	}
	since := time.Since(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
	}
}

func Bench_Linear_SFDA_Map_Get(sfda *sfda_map.SFDA_Map[uint64, uint64], n uint64) Test_Result {
	t = 0
	start = time.Now()
	for i := uint64(0); i < n; i++ {
		res := sfda.Get(i + 1)
		t += res.Value
	}
	since := time.Since(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
		Checksum:     t,
	}
}

func Generate_Random_Keys(n uint64) []uint64 {
	keys := make([]uint64, n)
	for i := uint64(0); i < n; i++ {
		keys[i] = i + 1
	}
	rand.Shuffle(int(n), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})
	return keys
}

func Bench_Random_Builtin_Map_Get(builtin_map map[uint64]uint64, random_keys []uint64) Test_Result {
	start = time.Now()
	var end time.Time
	for i := 0; i < len(random_keys); i++ {
		key := random_keys[i]
		_, ok := builtin_map[key]
		if !ok {
			log.Fatalf("Key %d not found.\n", key)
		}
	}
	end = time.Now()
	since := end.Sub(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
	}
}

func Bench_Random_SFDA_Map_Get(sfda *sfda_map.SFDA_Map[uint64, uint64], random_keys []uint64) Test_Result {
	start = time.Now()
	var end time.Time
	for i := 0; i < len(random_keys); i++ {
		key := random_keys[i]
		res := sfda.Get(key)
		if res.Did_Find == false {
			log.Fatalf("Key %d not found.\n", key)
		}
	}
	end = time.Now()
	since := end.Sub(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
	}
}

func Bench_Deletion_Builtin_Map(builtin_map map[uint64]uint64, keys []uint64) Test_Result {
	start := time.Now()
	for _, key := range keys {
		delete(builtin_map, key)
	}
	since := time.Since(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
	}
}

func Bench_Deletion_SFDA_Map(sfda *sfda_map.SFDA_Map[uint64, uint64], keys []uint64) Test_Result {
	start := time.Now()
	for _, key := range keys {
		sfda.Delete(key)
	}
	since := time.Since(start)
	return Test_Result{
		Elapsed_Time: since.Microseconds(),
	}
}

func Bench_Mem_Usage_Builtin_Map(f func() map[uint64]uint64, n uint64) Test_Result {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	before := m.Alloc

	builtin_map := f()

	// Perform insertions
	for i := uint64(0); i < n; i++ {
		builtin_map[i+1] = i
	}

	runtime.ReadMemStats(&m)
	after := m.Alloc

	return Test_Result{
		Memory_Usage: after - before,
	}
}

func Bench_Mem_Usage_SFDA_Map(f func() *sfda_map.SFDA_Map[uint64, uint64], n uint64) Test_Result {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	before := m.Alloc

	sfda := f()

	// Perform insertions
	for i := uint64(0); i < n; i++ {
		sfda.Set(i+1, i)
	}

	runtime.ReadMemStats(&m)
	after := m.Alloc

	return Test_Result{
		Memory_Usage: after - before,
	}
}

// func Bench_Concurrent_Access_Builtin_Map(builtin_map map[uint64]uint64, n uint64) {
// 	var end_write, end_read time.Time

// 	var wg sync.WaitGroup
// 	numGoroutines := 100
// 	operationsPerGoroutine := n / uint64(numGoroutines)

// 	// Writes
// 	wg.Add(numGoroutines)
// 	start = time.Now()
// 	for i := 0; i < numGoroutines; i++ {
// 		go func(id int) {
// 			defer wg.Done()
// 			for j := uint64(0); j < operationsPerGoroutine; j++ {
// 				key := uint64(id)*operationsPerGoroutine + j + 1
// 				builtin_map[key] = key
// 			}
// 		}(i)
// 	}
// 	wg.Wait()
// 	end_write = time.Now()

// 	// Reads
// 	wg.Add(numGoroutines)
// 	start = time.Now()
// 	for i := 0; i < numGoroutines; i++ {
// 		go func(id int) {
// 			defer wg.Done()
// 			for j := uint64(0); j < operationsPerGoroutine; j++ {
// 				key := uint64(id)*operationsPerGoroutine + j + 1
// 				_ = builtin_map[key]
// 			}
// 		}(i)
// 	}
// 	wg.Wait()
// 	end_read = time.Now()

// 	fmt.Println("Built-in map write time:", end_write.Sub(start).Microseconds())
// 	fmt.Println("Built-in map read time:", end_read.Sub(start).Microseconds())
// }

// func Bench_Concurrent_Access_SFDA_Map(sfda *sfda_map.SFDA_Map[uint64, uint64], n uint64) {
// 	var end_write, end_read time.Time

// 	var wg sync.WaitGroup
// 	numGoroutines := 100
// 	operationsPerGoroutine := n / uint64(numGoroutines)

// 	// Writes
// 	wg.Add(numGoroutines)
// 	start = time.Now()
// 	for i := 0; i < numGoroutines; i++ {
// 		go func(id int) {
// 			defer wg.Done()
// 			for j := uint64(0); j < operationsPerGoroutine; j++ {
// 				key := uint64(id)*operationsPerGoroutine + j + 1
// 				sfda.Set(key, key)
// 			}
// 		}(i)
// 	}
// 	wg.Wait()
// 	end_write = time.Now()

// 	// Reads
// 	wg.Add(numGoroutines)
// 	start = time.Now()
// 	for i := 0; i < numGoroutines; i++ {
// 		go func(id int) {
// 			defer wg.Done()
// 			for j := uint64(0); j < operationsPerGoroutine; j++ {
// 				key := uint64(id)*operationsPerGoroutine + j + 1
// 				sfda.Get(key)
// 			}
// 		}(i)
// 	}
// 	wg.Wait()
// 	end_read = time.Now()

// 	fmt.Println("SFDA map write time:", end_write.Sub(start).Microseconds())
// 	fmt.Println("SFDA map read time:", end_read.Sub(start).Microseconds())
// }

// func Bench_SFDA_Resizable_Map(f *os.File, n uint64) {
// 	start = time.Now()
// 	sfda_resizable := sfda_map.Make_SFDA_Resizable_Map(uint32(1024), 4)
// 	sfda_resizable.Start_Background_Services()
// 	defer sfda_resizable.Stop_Background_Services()

// 	for i := uint64(0); i < n; i++ {
// 		sfda_resizable.Set(i+1, i)
// 	}
// 	fmt.Println("SFDA resizable map set time:", time.Since(start))

// 	t = 0
// 	start = time.Now()
// 	//pprof.StartCPUProfile(f)
// 	for i := uint64(0); i < n; i++ {
// 		x, _ := sfda_resizable.Get(i + 1)
// 		t += x.(uint64)
// 	}
// 	//pprof.StopCPUProfile()
// 	fmt.Println("SFDA resizable map get time:", time.Since(start))
// 	fmt.Println("Sum:", t)
// }

func Test_Consistency(n uint64) {
	sfda_map := sfda_map.New[uint64, uint64](n)

	for i := uint64(0); i < n; i++ {
		sfda_map.Set(i+1, i)
	}

	for i := uint64(0); i < n; i++ {
		res := sfda_map.Get(i + 1)
		if res.Value != i {
			log.Fatalf("Expected %d, got %d\n", i, res.Value)
		}
	}
}
