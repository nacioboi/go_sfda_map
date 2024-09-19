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
	for i := 0; i < len(b.entries); i++ {
		if b.entries[i].key == key {
			return T_Get_Result[VT]{
				Value:       b.entries[i].value,
				Did_Succeed: true,
			}
		}
	}
	var zero VT
	return T_Get_Result[VT]{
		Value:       zero,
		Did_Succeed: false,
	}
}

type ll_bucket[KT I_Positive_Integer, VT any] struct {
	entries []linked_list[t_bucket_entry[KT, VT]]
}

func (b *ll_bucket[KT, VT]) set(key KT, value VT, m *SFDA_Map[KT, VT], b_ptr *bucket[KT, VT]) {
	remainder := key & 7

	// We must ensure that we don't already have the same key contained within the same bucket...
	//m.setter__lazy_safety_check_queue <- t__setter__safety_check_params[KT, VT]{buck: b_ptr, key: key}
	b.entries[remainder].append(t_bucket_entry[KT, VT]{key: key, value: value})
}

func (b *ll_bucket[KT, VT]) get(key KT) T_Get_Result[VT] {
	remainder := key & 7

	val, did_find := b.entries[remainder].iter(func(entry t_bucket_entry[KT, VT]) bool {
		if entry.key == key {
			return true
		}
		return false
	})

	return T_Get_Result[VT]{
		Value:       val.value,
		Did_Succeed: did_find,
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

func New_SFDA_Map[KT I_Positive_Integer, VT any](
	expected_num_inputs KT,
	options ...T_Option[KT, VT],
) *SFDA_Map[KT, VT] {
	num_buckets_runtime, prof := parse_profile(expected_num_inputs, options)
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
		if opt.t != OPTION_TYPE__WITH_PERFORMANCE_PROFILE {
			opt.f(&inst)
		}
	}

	estimated_num_entries_per_bucket := expected_num_inputs / num_buckets
	switch prof {
	case PERFORMANCE_PROFILE__NORMAL:
		if estimated_num_entries_per_bucket > KT(c_NUM_ENTRIES_NORMAL_MODE) {
			panic("Invalid number of entries per bucket.")
		}
		for i := uint64(0); i < num_buckets_runtime; i++ {
			b := ll_bucket[KT, VT]{
				entries: make([]linked_list[t_bucket_entry[KT, VT]], 8),
			}
			(*inst.buckets)[i] = bucket[KT, VT]{
				inner: &b,
			}
		}
	default:
		for i := uint64(0); i < num_buckets_runtime; i++ {
			b := linear_bucket[KT, VT]{
				entries: make([]t_bucket_entry[KT, VT], 0, estimated_num_entries_per_bucket),
			}
			(*inst.buckets)[i] = bucket[KT, VT]{
				inner: &b,
			}
		}
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
	// TODO: interface adds some overhead.
	buck.inner.set(key, value, m, buck)
}

type T_Get_Result[VT any] struct {
	Value       VT
	Did_Succeed bool
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
