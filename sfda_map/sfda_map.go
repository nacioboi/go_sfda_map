/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

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

// Super-Fast Direct-Access Map.
type SFDA_Map[KT I_Positive_Integer, VT any] struct {
	buckets        []bucket[KT, VT]
	num_buckets_m1 KT

	users_chosen_hash_func func(KT) uint64
	using_users_hash_func  bool

	profile T_Performance_Profile
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
		buckets:        buckets,
		num_buckets_m1: num_buckets - 1,
		profile:        profile,
	}

	// Apply options...
	for _, opt := range options {
		if opt.t != OPTION_TYPE__WITH_PERFORMANCE_PROFILE {
			opt.f(&inst)
		}
	}

	return &inst
}

func (m *SFDA_Map[KT, VT]) Enquire_Number_Of_Buckets() KT {
	return m.num_buckets_m1 + 1
}

// Set a key-value pair in the map.
// Will panic if something goes wrong.
//
// - WARNING: This function is NOT thread-safe.
//
//go:inline
func (m *SFDA_Map[KT, VT]) Set(key KT, value VT) {
	if key == 0 {
		panic("Key cannot be 0.")
	}

	index := key & m.num_buckets_m1
	buck := &m.buckets[index]

	for i := 0; i < len(buck.entries); i++ {
		if buck.entries[i].key == key {
			buck.entries[i].value = value
			return
		}
	}

	buck.entries = append(buck.entries, t_bucket_entry[KT, VT]{key: key, value: value})
}

// Returns the value and a boolean indicating whether the value was found.
//
// - WARNING: This function is NOT thread-safe.
//
// - NOTE: Remember that keys cannot be 0.
//
// - NOTE: This function will not check if the key is 0.
//
//go:inline
func (m *SFDA_Map[KT, VT]) Get(key KT) (VT, bool) {
	// NOTE: Keeping value type here improves performance since we do not modify the value.
	buck := m.buckets[key&m.num_buckets_m1]

	for i := 0; i < len(buck.entries); i++ {
		if buck.entries[i].key == key {
			return buck.entries[i].value, true
		}
	}

	var zero VT
	return zero, false
}

// Delete an entry from the map and return a boolean indicating whether the entry was found.
//
// - WARNING: This function is NOT thread-safe.
//
// - NOTE: Remember that keys cannot be 0.
//
// - NOTE: This function will not check if the key is 0.
//
//go:inline
func (m *SFDA_Map[KT, VT]) Delete(key KT) bool {
	index := key & m.num_buckets_m1
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
