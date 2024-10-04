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
	"strings"

	"github.com/nacioboi/go_sfda_map/sfda_map"
	tests "github.com/nacioboi/go_sfda_map/tests"
)

func format_Number_With_Commas(n int64) string {
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		s = s[1:]
	}
	var result []string
	for len(s) > 3 {
		result = append(result, s[len(s)-3:])
		s = s[:len(s)-3]
	}
	result = append(result, s)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	formattedNumber := strings.Join(result, ",")
	if n < 0 {
		formattedNumber = "-" + formattedNumber
	}
	return formattedNumber
}

func main() {
	tests.Test_Consistency(1024)

	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)
	defer runtime.GC()

	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n_normal := uint64(1024 * 1024 * 32)
	n_memory := uint64(1024)

	// Create maps...
	bm := make(map[uint64]uint64)
	sfda_64 := sfda_map.New[uint64, uint64](
		n_normal,
		sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__64_ENTRIES_PER_BUCKET),
	)
	sfda_32 := sfda_map.New[uint64, uint64](
		n_normal,
		sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__32_ENTRIES_PER_BUCKET),
	)
	sfda_1 := sfda_map.New[uint64, uint64](
		n_normal,
		sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__1_ENTRIES_PER_BUCKET),
	)

	bm_m_f := func() map[uint64]uint64 {
		return make(map[uint64]uint64)
	}
	sfda_64_m_f := func() *sfda_map.SFDA_Map[uint64, uint64] {
		return sfda_map.New[uint64, uint64](
			n_memory,
			sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__64_ENTRIES_PER_BUCKET),
		)
	}
	sfda_1_m_f := func() *sfda_map.SFDA_Map[uint64, uint64] {
		return sfda_map.New[uint64, uint64](
			n_memory,
			sfda_map.With_Performance_Profile[uint64, uint64](sfda_map.PERFORMANCE_PROFILE__1_ENTRIES_PER_BUCKET),
		)
	}

	var res tests.Test_Result

	// Benchmark Linear Set...
	res = tests.Bench_Linear_Builtin_Map_Set(bm, n_normal)
	fmt.Printf("\nBM       :: LINEAR SET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("BM       :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	res = tests.Bench_Linear_SFDA_Map_Set(sfda_64, n_normal)
	fmt.Printf("SFDA 64  :: LINEAR SET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA 64  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	res = tests.Bench_Linear_SFDA_Map_Set(sfda_32, n_normal)
	fmt.Printf("SFDA 32  :: LINEAR SET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA 32  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	res = tests.Bench_Linear_SFDA_Map_Set(sfda_1, n_normal)
	fmt.Printf("SFDA  1  :: LINEAR SET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA  1  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))

	// Benchmark Linear Get...
	res = tests.Bench_Linear_Builtin_Map_Get(bm, n_normal)
	fmt.Printf("\nBM       :: LINEAR GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("BM       :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	bm_checksum := res.Checksum
	res = tests.Bench_Linear_SFDA_Map_Get(sfda_64, n_normal)
	fmt.Printf("SFDA 64  :: LINEAR GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA 64  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	sfda_64_checksum := res.Checksum
	res = tests.Bench_Linear_SFDA_Map_Get(sfda_32, n_normal)
	fmt.Printf("SFDA 32  :: LINEAR GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA 32  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	sfda_32_checksum := res.Checksum
	res = tests.Bench_Linear_SFDA_Map_Get(sfda_1, n_normal)
	fmt.Printf("SFDA  1  :: LINEAR GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA  1  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	sfda_1_checksum := res.Checksum

	// Benchmark Random get...
	data := tests.Generate_Random_Keys(n_normal)
	res = tests.Bench_Random_Builtin_Map_Get(bm, data)
	fmt.Printf("\nBM       :: RANDOM GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("BM       :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	pprof.StartCPUProfile(f)
	res = tests.Bench_Random_SFDA_Map_Get(sfda_64, data)
	pprof.StopCPUProfile()
	fmt.Printf("SFDA 64  :: RANDOM GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA 64  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	res = tests.Bench_Random_SFDA_Map_Get(sfda_32, data)
	fmt.Printf("SFDA 32  :: RANDOM GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA 32  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))
	res = tests.Bench_Random_SFDA_Map_Get(sfda_1, data)
	fmt.Printf("SFDA  1  :: RANDOM GET            :: %d\n", res.Elapsed_Time)
	fmt.Printf("SFDA  1  :: MICROSECONDS PER OP   :: %f\n", float64(res.Elapsed_Time)/float64(n_normal))

	// Benchmark memory usage...
	res = tests.Bench_Mem_Usage_Builtin_Map(bm_m_f, n_memory)
	fmt.Printf("\nBM       :: MEMORY USAGE           :: %s\n", format_Number_With_Commas(int64(res.Memory_Usage)))
	res = tests.Bench_Mem_Usage_SFDA_Map(sfda_64_m_f, n_memory)
	fmt.Printf("SFDA 64  :: MEMORY USAGE           :: %s\n", format_Number_With_Commas(int64(res.Memory_Usage)))
	res = tests.Bench_Mem_Usage_SFDA_Map(sfda_1_m_f, n_memory)
	fmt.Printf("SFDA  1  :: MEMORY USAGE           :: %s\n", format_Number_With_Commas(int64(res.Memory_Usage)))

	// Print checksums...
	fmt.Printf("\nBM       :: CHECKSUM               :: %d\n", bm_checksum)
	fmt.Printf("SFDA 64  :: CHECKSUM               :: %d\n", sfda_64_checksum)
	fmt.Printf("SFDA 32  :: CHECKSUM               :: %d\n", sfda_32_checksum)
	fmt.Printf("SFDA  1  :: CHECKSUM               :: %d\n", sfda_1_checksum)

	// Assert checksums...
	for _, checksum := range []uint64{sfda_64_checksum, sfda_32_checksum, sfda_1_checksum} {
		if checksum != bm_checksum {
			log.Fatalf("Checksums do not match!")
		}
	}
	fmt.Printf("\nChecksums match!\n")
}
