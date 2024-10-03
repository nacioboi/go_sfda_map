/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

import (
	"runtime"
	"time"
)

type T_Primitive uint8

const (
	PRIMITIVE_TYPE_UINT64 T_Primitive = iota
)

type I_Positive_Integer interface {
	uint8 | uint16 | uint32 | uint64
}

type t_bucket_entry[KT I_Positive_Integer, VT any] struct {
	key   KT
	value VT
}

type bucket[KT I_Positive_Integer, VT any] struct {
	entries []t_bucket_entry[KT, VT]
}

func (b *bucket[KT, VT]) already_inside(key KT) bool {
	for i := 0; i < len(b.entries); i++ {
		if b.entries[i].key == key {
			return true
		}
	}
	return false
}

type t__setter__safety_check_params[KT I_Positive_Integer, VT any] struct {
	buck *bucket[KT, VT]
	key  KT
}

// Super-Fast Direct-Access Map.
type SFDA_Map[KT I_Positive_Integer, VT any] struct {
	buckets     []bucket[KT, VT]
	num_buckets KT

	users_chosen_hash_func func(KT) uint64
	using_users_hash_func  bool

	exit_chan chan struct{}

	setter__lazy_safety_check_queue chan t__setter__safety_check_params[KT, VT]
}

func New[KT I_Positive_Integer, VT any](
	expected_num_inputs KT,
	options ...T_Option[KT, VT],
) *SFDA_Map[KT, VT] {
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
	default:
		panic("Invalid performance profile.")
	}

	if num_buckets%2 != 0 {
		panic("numBuckets should be a multiple of 2.")
	}

	// Allocate buckets...
	num_buckets_runtime := any(num_buckets).(uint64)
	buckets := make([]bucket[KT, VT], num_buckets_runtime)
	estimated_num_entries_per_bucket := expected_num_inputs / num_buckets
	for i := uint64(0); i < num_buckets_runtime; i++ {
		b := bucket[KT, VT]{
			entries: make([]t_bucket_entry[KT, VT], 0, estimated_num_entries_per_bucket),
		}
		buckets[i] = b
	}

	// Instantiate...
	inst := SFDA_Map[KT, VT]{
		buckets:                         buckets,
		num_buckets:                     num_buckets,
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

	// Start background goroutines...
	go inst.setter__bg__lazy_safety_check_handler()

	return &inst
}

func (m *SFDA_Map[KT, VT]) Enquire_Number_Of_Buckets() KT {
	return m.num_buckets
}

func (m *SFDA_Map[KT, VT]) hash_func(key KT) KT {
	return key & (m.num_buckets - 1)
}

func (m *SFDA_Map[KT, VT]) _inner__setter__bg__lazy_safety_check_handler(params t__setter__safety_check_params[KT, VT]) {
	if params.buck.already_inside(params.key) {
		panic("Key already exists in the map.")
	}
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

	index := m.hash_func(key)
	buck := &m.buckets[index]

	for i := 0; i < len(buck.entries); i++ {
		if buck.entries[i].key == key {
			// We must ensure that we don't already have the same key contained within the same bucket...
			m.setter__lazy_safety_check_queue <- t__setter__safety_check_params[KT, VT]{buck: buck, key: key}
			buck.entries[i].value = value
			return
		}
	}

	buck.entries = append(buck.entries, t_bucket_entry[KT, VT]{key: key, value: value})
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
	index := m.hash_func(key)
	// NOTE: Keeping value type here improves performance since we do not modify the value.
	buck := m.buckets[index]

	var e t_bucket_entry[KT, VT]
	for i := 0; i < len(buck.entries); i++ {
		e = buck.entries[i]
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

// Delete an entry from the map and return a boolean indicating whether the entry was found.
//
// - WARNING: This function is NOT thread-safe.
//
// - NOTE: Remember that keys cannot be 0.
//
// - NOTE: This function will not check if the key is 0.
func (m *SFDA_Map[KT, VT]) Delete(key KT) bool {
	index := m.hash_func(key)
	buck := &m.buckets[index]

	loc := -1

	for i := 0; i < len(buck.entries); i++ {
		if buck.entries[i].key == key {
			loc = i
			break
		}
	}

	// Rearrange the entire slice...
	if loc == -1 {
		return false
	}

	buck.entries = append(buck.entries[:loc], buck.entries[loc+1:]...)
	return true
}
