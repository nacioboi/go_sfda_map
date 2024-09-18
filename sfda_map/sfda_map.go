/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

import (
	"runtime"
	"time"
)

type i_bucket_entries[KT I_Positive_Integer, VT any] interface {
	set(key KT, value VT, m *SFDA_Map[KT, VT], b_ptr *bucket[KT, VT])
	get(key KT) T_Get_Result[VT]
}
type t_bucket_entry[KT I_Positive_Integer, VT any] struct {
	key   KT
	value VT
}
type bucket[KT I_Positive_Integer, VT any] struct {
	inner i_bucket_entries[KT, VT]
}

// const cluster_size = 4
// const num_clusters = 128 / cluster_size

// type cluster[KT I_Positive_Integer, VT any] struct {
// 	entries []t_bucket_entry[KT, VT]
// }
// type sorted_bucket[KT I_Positive_Integer, VT any] struct {
// 	clusters [num_clusters]cluster[KT, VT]
// }

// func (b *sorted_bucket[KT, VT]) find_idx(cluster_idx int) int {
// 	cpy := b.clusters[cluster_idx]

// 	// Find first available slot in the cluster...
// 	for i := 0; i < len(cpy.entries); i++ {
// 		if cpy.entries[i].key == 0 {
// 			return i
// 		}
// 	}
// 	// If we reach here, the cluster is full, append to the end...
// 	b.clusters[cluster_idx].entries = append(b.clusters[cluster_idx].entries, t_bucket_entry[KT, VT]{})
// 	return len(b.clusters[cluster_idx].entries) - 1
// }

// func (b *sorted_bucket[KT, VT]) set(key KT, value VT, m *SFDA_Map[KT, VT], b_ptr *bucket[KT, VT]) {
// 	cluster_idx := int((key - 1) % num_clusters)
// 	x := (int(key-1) / num_clusters)

// 	if x >= len(b.clusters[cluster_idx].entries) {
// 		base_add := x - len(b.clusters[cluster_idx].entries) + 1
// 		more_entries := make([]t_bucket_entry[KT, VT], base_add+256)
// 		b.clusters[cluster_idx].entries = append(b.clusters[cluster_idx].entries, more_entries...)
// 	}

// 	b.clusters[cluster_idx].entries[x] = t_bucket_entry[KT, VT]{key: key, value: value}
// }

// func (b *sorted_bucket[KT, VT]) get(key KT) T_Get_Result[VT] {
// 	cluster_idx := int((key - 1) % num_clusters)
// 	x := (int(key-1) / num_clusters)

// 	if b.clusters[cluster_idx].entries[x].key == key {
// 		return T_Get_Result[VT]{Value: b.clusters[cluster_idx].entries[x].value, Did_Find: true}
// 	}

// 	var zero VT
// 	return T_Get_Result[VT]{Value: zero, Did_Find: false}
// }

type linear_bucket[KT I_Positive_Integer, VT any] struct {
	entries []t_bucket_entry[KT, VT]
}

func (b *linear_bucket[KT, VT]) set(key KT, value VT, m *SFDA_Map[KT, VT], b_ptr *bucket[KT, VT]) {
	for i := 0; i < len(b.entries); i++ {
		if b.entries[i].key == key {
			// We must ensure that we don't already have the same key contained within the same bucket...
			m.setter__lazy_safety_check_queue <- t__setter__safety_check_params[KT, VT]{buck: b_ptr, key: key}
			b.entries[i].value = value
			return
		}
	}

	b.entries = append(b.entries, t_bucket_entry[KT, VT]{key: key, value: value})
}

func (b *linear_bucket[KT, VT]) get(key KT) T_Get_Result[VT] {
	var e t_bucket_entry[KT, VT]
	for i := 0; i < len(b.entries); i++ {
		e = b.entries[i]
		if e.key == key {
			return T_Get_Result[VT]{
				Value:    e.value,
				Did_Find: true,
			}
		}
	}
	var zero VT
	return T_Get_Result[VT]{
		Value:    zero,
		Did_Find: false,
	}
}

// func (b *bucket[KT, VT]) already_inside(key KT) bool {
// 	for i := 0; i < len(b.entries); i++ {
// 		if b.entries[i].key == key {
// 			return true
// 		}
// 	}
// 	return false
// }

type t__setter__safety_check_params[KT I_Positive_Integer, VT any] struct {
	buck *bucket[KT, VT]
	key  KT
}

// Super-Fast Direct-Access Map.
type SFDA_Map[KT I_Positive_Integer, VT any] struct {
	buckets             *[]bucket[KT, VT]
	num_buckets         KT
	num_buckets_runtime uint64

	users_chosen_hash_func func(KT) uint64
	using_users_hash_func  bool

	exit_chan chan struct{}

	setter__lazy_safety_check_queue chan t__setter__safety_check_params[KT, VT]

	performance_profile T_Performance_Profile
}

func determine_num_buckets[KT I_Positive_Integer, VT any](
	expected_num_inputs KT,
	options []T_Option[KT, VT],
) uint64 {
	expected_num_inputs = next_power_of_two(expected_num_inputs)

	profile := PERFORMANCE_PROFILE__8_ENTRIES_PER_BUCKET
	for _, opt := range options {
		if opt.t == OPTION_TYPE__WITH_PERFORMANCE_PROFILE {
			profile = opt.other.(T_Performance_Profile)
		}
	}

	var num_buckets KT
	switch profile {
	case PERFORMANCE_PROFILE__1_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs
	case PERFORMANCE_PROFILE__2_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 2
	case PERFORMANCE_PROFILE__4_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 4
	case PERFORMANCE_PROFILE__8_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 8
	case PERFORMANCE_PROFILE__16_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 16
	case PERFORMANCE_PROFILE__32_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 32
	case PERFORMANCE_PROFILE__64_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 64
	case PERFORMANCE_PROFILE__128_ENTRIES_PER_BUCKET:
		num_buckets = expected_num_inputs / 128
	default:
		panic("Invalid performance profile.")
	}

	if num_buckets%2 != 0 {
		panic("numBuckets should be a multiple of 2.")
	}

	return uint64(num_buckets)
}

func (m *SFDA_Map[KT, VT]) allocate_linear_buckets(
	num_buckets_runtime uint64,
	expected_num_inputs KT,
	num_buckets KT,
) {
	estimated_num_entries_per_bucket := expected_num_inputs / num_buckets
	for i := uint64(0); i < num_buckets_runtime; i++ {
		b := linear_bucket[KT, VT]{
			entries: make([]t_bucket_entry[KT, VT], 0, estimated_num_entries_per_bucket),
		}
		(*m.buckets)[i] = bucket[KT, VT]{
			inner: &b,
		}
	}
}

// func (m *SFDA_Map[KT, VT]) allocate_sorted_buckets(
// 	num_buckets_runtime uint64,
// ) {
// 	for i := uint64(0); i < num_buckets_runtime; i++ {
// 		var clusters [num_clusters]cluster[KT, VT]
// 		for j := 0; j < num_clusters; j++ {
// 			entries := make([]t_bucket_entry[KT, VT], cluster_size)
// 			clusters[j] = cluster[KT, VT]{
// 				entries: entries,
// 			}
// 		}
// 		b := sorted_bucket[KT, VT]{
// 			clusters: clusters,
// 		}
// 		(*m.buckets)[i] = bucket[KT, VT]{
// 			inner: &b,
// 		}
// 	}
// }

func New_SFDA_Map[KT I_Positive_Integer, VT any](
	expected_num_inputs KT,
	options ...T_Option[KT, VT],
) *SFDA_Map[KT, VT] {
	num_buckets_runtime := determine_num_buckets(expected_num_inputs, options)
	num_buckets := KT(num_buckets_runtime)

	// Allocate buckets...
	buckets := make([]bucket[KT, VT], num_buckets_runtime)

	// Instantiate...
	inst := SFDA_Map[KT, VT]{
		buckets:                         &buckets,
		num_buckets:                     num_buckets,
		num_buckets_runtime:             num_buckets_runtime,
		exit_chan:                       make(chan struct{}),
		setter__lazy_safety_check_queue: make(chan t__setter__safety_check_params[KT, VT], 4096),
	}

	runtime.SetFinalizer(&inst, func(m *SFDA_Map[KT, VT]) {
		close(m.exit_chan)
	})

	// Apply options...
	for _, opt := range options {
		opt.f(&inst)
	}

	switch inst.performance_profile {
	case PERFORMANCE_PROFILE__128_ENTRIES_PER_BUCKET:
		//inst.allocate_sorted_buckets(num_buckets_runtime)
		panic("not implemented")
	default:
		inst.allocate_linear_buckets(num_buckets_runtime, expected_num_inputs, num_buckets)
	}

	// Start background goroutines...
	go inst.setter__bg__lazy_safety_check_handler()

	return &inst
}

func (m *SFDA_Map[KT, VT]) Enquire_Number_Of_Buckets() KT {
	return m.num_buckets
}

func (m *SFDA_Map[KT, VT]) hash_func(key KT, n KT) KT {
	return key & (n - 1)
}

func (m *SFDA_Map[KT, VT]) _inner__setter__bg__lazy_safety_check_handler(params t__setter__safety_check_params[KT, VT]) {
	// if params.buck.already_inside(params.key) {
	// 	panic("Key already exists in the map.")
	// }
}

func (m *SFDA_Map[KT, VT]) setter__bg__lazy_safety_check_handler() {
	for {
		select {
		case params := <-m.setter__lazy_safety_check_queue:
			m._inner__setter__bg__lazy_safety_check_handler(params)
		case <-m.exit_chan:
			return
		default:
			time.Sleep(250 * time.Millisecond)
			continue
		}
	}
}

// Set a key-value pair in the map.
// Will panic if something goes wrong.
//
// - WARNING: This function is NOT thread-safe.
func (m *SFDA_Map[KT, VT]) Set(key KT, value VT) {
	if key == 0 {
		panic("Key cannot be 0.")
	}

	idx := m.hash_func(key, m.num_buckets)
	buck := &(*m.buckets)[idx]
	buck.inner.set(key, value, m, buck)
}

type T_Get_Result[VT any] struct {
	Value    VT
	Did_Find bool
}

// Returns the value and a boolean indicating whether the value was found.
//
// - WARNING: This function is NOT thread-safe.
//
// - NOTE: Remember that keys cannot be 0.
//
// - NOTE: This function will not check if the key is 0.
func (m *SFDA_Map[KT, VT]) Get(key KT) T_Get_Result[VT] {
	idx := m.hash_func(key, m.num_buckets)
	buck := &(*m.buckets)[idx]
	return buck.inner.get(key)
}

// Delete an entry from the map and return a boolean indicating whether the entry was found.
//
// - WARNING: This function is NOT thread-safe.
//
// - NOTE: Remember that keys cannot be 0.
//
// - NOTE: This function will not check if the key is 0.
func (m *SFDA_Map[KT, VT]) Delete(key KT) bool {
	// index := m.hash_func(key, m.num_buckets)
	// buck := &(*m.buckets)[index]

	// loc := -1

	// for i := 0; i < len(buck.entries); i++ {
	// 	if buck.entries[i].key == key {
	// 		loc = i
	// 		break
	// 	}
	// }

	// // Rearrange the entire slice...
	// if loc == -1 {
	// 	return false
	// }

	// buck.entries = append(buck.entries[:loc], buck.entries[loc+1:]...)
	// return true
	panic("Not implemented.")
}
