/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	BG1_TARGET_NUM_BUCKETS = 1024

	// Preallocate how many buckets...
	PREALLOC_PREALLOCATED_BUCKETS = 16
	// Preallocate size of channels...
	PREALLOC_BG1_NUM_BUCKETS_CHAN = 256
	// Preallocate size of channels...
	PREALLOC_BG2_NUM_ENTRIES_CHAN = 256
	// Preallocate size of channels...
	PREALLOC_BG2_EXIT_CHAN = 1

	// On multi-core systems, this should be small.
	BG1_SLEEP_DEF_CASE_MILLISECONDS = 1
	BG1_SLEEP_ALWAYS_MILLISECONDS   = 10
	// On multi-core systems, this should be small.
	BG2_SLEEP_MILLISECONDS = 150
)

func (m *SFDA_Resizable_Map[KT, VT]) bg1_bucket_allocator() {
	local_num_buckets_target := uint16(10240)
	for {
		select {
		case num_buckets_target := <-m.bg_num_buckets_target_chan:
			local_num_buckets_target = num_buckets_target
		case <-m.bg_bucket_allocator_exit_chan:
			return
		default:
			time.Sleep(BG1_SLEEP_DEF_CASE_MILLISECONDS * time.Millisecond)
		}
		time.Sleep(BG1_SLEEP_ALWAYS_MILLISECONDS * time.Millisecond)

		for i := uint16(m.preallocated_buckets.len()); i < local_num_buckets_target; i++ {
			b := bucket[KT, VT]{
				entries: make([]t_bucket_entry[KT, VT], 0, m.target_i_per_b),
			}
			m.preallocated_buckets.push(b)
		}
	}
}

func (m *SFDA_Resizable_Map[KT, VT]) Resize(
	new_num_buckets KT,
	reusable_buckets *[]bucket[KT, VT],
	options ...T_Option[KT, VT],
) {
	if m.map_.num_buckets_runtime == uint64(new_num_buckets) {
		return
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	if new_num_buckets%2 != 0 {
		panic("numBuckets should be a multiple of 2.")
	}

	num_buckets := KT(new_num_buckets)
	num_buckets_runtime := uint64(new_num_buckets)

	old_num_buckets := m.map_.num_buckets_runtime

	reusable_buckets_deref := *reusable_buckets

	inst := SFDA_Map[KT, VT]{
		// First reuse the already allocated buckets...
		buckets:                         reusable_buckets,
		num_buckets:                     num_buckets,
		num_buckets_runtime:             num_buckets_runtime,
		exit_chan:                       make(chan struct{}),
		setter__lazy_safety_check_queue: make(chan t__setter__safety_check_params[KT, VT], 4096),
	}

	// Then allocate new buckets...
	for i := uint64(len(reusable_buckets_deref)); i < num_buckets_runtime; i++ {
		reusable_buckets_deref = append(reusable_buckets_deref, m.preallocated_buckets.pop())
	}

	var items_to_reassign []t_bucket_entry[KT, VT]

	// Re-assign values from the reusable buckets...
	var b *bucket[KT, VT]
	for i := uint64(0); i < num_buckets_runtime; i++ {
		b = &reusable_buckets_deref[i]
		for j := 0; j < len(b.entries); j++ {
			if b.entries[j].key != 0 {
				items_to_reassign = append(items_to_reassign, b.entries[j])
			}
		}
	}

	var e t_bucket_entry[KT, VT]
	var old_index KT
	var new_index KT
	var loc = -1
	for i := 0; i < len(items_to_reassign); i++ {
		e = items_to_reassign[i]

		old_index = inst.hash_func(e.key, KT(old_num_buckets))
		b = &reusable_buckets_deref[old_index]

		for i := 0; i < len(b.entries); i++ {
			if b.entries[i].key == e.key {
				loc = i
				break
			}
		}

		if loc == -1 {
			panic("Key not found in the old map.")
		}
		b.entries = append(b.entries[:loc], b.entries[loc+1:]...)

		new_index = inst.hash_func(e.key, KT(num_buckets_runtime))
		b = &reusable_buckets_deref[new_index]

		b.entries = append(b.entries, e)
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

	// Remove the pointer to the reusable buckets from the old map...
	m.map_.buckets = nil

	// Assign the new map to the resizable map...
	m.map_ = &inst
	m.map_.buckets = &reusable_buckets_deref

	// Start background goroutines...
	go inst.setter__bg__lazy_safety_check_handler()
}

type SFDA_Resizable_Map[KT I_Positive_Integer, VT any] struct {
	target_i_per_b uint8

	map_                *SFDA_Map[KT, VT]
	mut                 *sync.Mutex
	bg_num_entries_chan chan int16
	bg_exit_chan        chan bool
	bg_difference_tally int16

	preallocated_buckets          *stack[bucket[KT, VT]]
	bg_num_buckets_target_chan    chan uint16
	bg_bucket_allocator_exit_chan chan bool
}

func New_SFDA_Resizable_Map[KT I_Positive_Integer, VT any](
	initial_expected_num_inputs KT,
	options ...T_Option[KT, VT],
) *SFDA_Resizable_Map[KT, VT] {
	num_buckets_runtime := determine_num_buckets(initial_expected_num_inputs, options)
	num_buckets := KT(num_buckets_runtime)

	estimated_num_entries_per_bucket := uint8(initial_expected_num_inputs / num_buckets)
	inst := SFDA_Resizable_Map[KT, VT]{
		target_i_per_b: estimated_num_entries_per_bucket,

		map_:                New_SFDA_Map[KT, VT](initial_expected_num_inputs, options...),
		bg_num_entries_chan: make(chan int16, PREALLOC_BG2_NUM_ENTRIES_CHAN),
		bg_exit_chan:        make(chan bool, PREALLOC_BG2_EXIT_CHAN),
		mut:                 &sync.Mutex{},

		preallocated_buckets:          new_stack[bucket[KT, VT]](PREALLOC_PREALLOCATED_BUCKETS),
		bg_num_buckets_target_chan:    make(chan uint16, PREALLOC_BG1_NUM_BUCKETS_CHAN),
		bg_bucket_allocator_exit_chan: make(chan bool, 1),
	}

	// Apply options...
	for _, opt := range options {
		if opt.t != OPTION_TYPE__WITH_PERFORMANCE_PROFILE {
			panic("only `With_Performance_Profile` option is supported with the `SFDA_Resizable_Map`")
		}
	}

	// Start the background goroutines...
	go inst.bg2()
	time.Sleep(100 * time.Millisecond)
	go inst.bg1_bucket_allocator()

	// TODO: Change this to automatically change the target.
	//inst.bg_num_buckets_target_chan <- uint16(estimated_num_entries_per_bucket)
	runtime.SetFinalizer(&inst, func(m *SFDA_Resizable_Map[KT, VT]) {
		m.bg_exit_chan <- true
		m.bg_bucket_allocator_exit_chan <- true
		close(m.bg_num_entries_chan)
	})

	return &inst
}

func (m *SFDA_Resizable_Map[KT, VT]) Get(key KT) T_Get_Result[VT] {
	return m.map_.Get(key)
}

const BG_TALLY_THRESHOLD = 512

func (m *SFDA_Resizable_Map[KT, VT]) Set(key KT, value VT) {
	m.mut.Lock()
	m.map_.Set(key, value)
	m.mut.Unlock()
	m.bg_difference_tally += +(1)
	if m.bg_difference_tally > int16(BG_TALLY_THRESHOLD) || m.bg_difference_tally < -int16(BG_TALLY_THRESHOLD) {
		m.send_tally()
	}
}

func (m *SFDA_Resizable_Map[KT, VT]) Delete(key KT) {
	m.map_.Delete(key)
	m.bg_difference_tally += -(1)
	if m.bg_difference_tally > int16(BG_TALLY_THRESHOLD) || m.bg_difference_tally < -int16(BG_TALLY_THRESHOLD) {
		m.send_tally()
	}
}

func (m *SFDA_Resizable_Map[KT, VT]) send_tally() {
	m.bg_num_entries_chan <- m.bg_difference_tally
	m.bg_difference_tally = 0
}

func (m *SFDA_Resizable_Map[KT, VT]) do_update(num_entries uint64, num_buckets uint64) uint64 {
	const MAX_TRIES = 50
	x := num_buckets
	for i := 0; i < MAX_TRIES; i++ {
		current_i_per_b := float32(num_entries) / float32(x)
		if current_i_per_b < float32(m.target_i_per_b) {
			x = previous_power_of_two(x - 1)
		} else if current_i_per_b > float32(m.target_i_per_b) {
			x = next_power_of_two(x + 1)
		}
		if current_i_per_b > float32(m.target_i_per_b*2) {

		}
	}
	return x
}

func (m *SFDA_Resizable_Map[KT, VT]) bg2() {
	last_time_cmd_issued := time.Now()
	local_num_entries := uint64(0)
	var previous_resize_target uint64 = 0
	for {
		select {
		case current_num_entries := <-m.bg_num_entries_chan:
			if current_num_entries > 0 {
				local_num_entries += uint64(current_num_entries)
			} else {
				local_num_entries -= uint64(-current_num_entries)
			}
		case <-m.bg_exit_chan:
			return
		}
		if time.Since(last_time_cmd_issued) > BG2_SLEEP_MILLISECONDS*time.Millisecond {
			cmd := m.do_update(local_num_entries, m.map_.num_buckets_runtime)
			if cmd != 0 && cmd != previous_resize_target {
				fmt.Printf("new command = [%d]\n", cmd)
				m.Resize(KT(cmd), m.map_.buckets)
			}
			previous_resize_target = cmd
			last_time_cmd_issued = time.Now()
		}
	}
}
