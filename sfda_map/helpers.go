/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

type T_Primitive uint8

const (
	PRIMITIVE_TYPE_UINT64 T_Primitive = iota
)

type I_Positive_Integer interface {
	uint8 | uint16 | uint32 | uint64
}

func _inner__next_power_of_two__uint64(n uint64) uint64 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}

func _inner__next_power_of_two__uint32(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

func _inner__next_power_of_two__uint16(n uint16) uint16 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n++
	return n
}

func _inner__next_power_of_two__uint8(n uint8) uint8 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n++
	return n
}

func next_power_of_two[KT I_Positive_Integer](n KT) KT {
	switch any(n).(type) {
	case uint64:
		return any(_inner__next_power_of_two__uint64(any(n).(uint64))).(KT)
	case uint32:
		return any(_inner__next_power_of_two__uint32(any(n).(uint32))).(KT)
	case uint16:
		return any(_inner__next_power_of_two__uint16(any(n).(uint16))).(KT)
	case uint8:
		return any(_inner__next_power_of_two__uint8(any(n).(uint8))).(KT)
	default:
		panic("Unsupported type.")
	}
}

func _inner__previous_power_of_two__uint64(n uint64) uint64 {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n - (n >> 1)
}

func _inner__previous_power_of_two__uint32(n uint32) uint32 {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n - (n >> 1)
}

func _inner__previous_power_of_two__uint16(n uint16) uint16 {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	return n - (n >> 1)
}

func _inner__previous_power_of_two__uint8(n uint8) uint8 {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	return n - (n >> 1)
}

func previous_power_of_two[KT I_Positive_Integer](n KT) KT {
	switch any(n).(type) {
	case uint64:
		return any(_inner__previous_power_of_two__uint64(any(n).(uint64))).(KT)
	case uint32:
		return any(_inner__previous_power_of_two__uint32(any(n).(uint32))).(KT)
	case uint16:
		return any(_inner__previous_power_of_two__uint16(any(n).(uint16))).(KT)
	case uint8:
		return any(_inner__previous_power_of_two__uint8(any(n).(uint8))).(KT)
	default:
		panic("Unsupported type.")
	}
}

func parse_profile[KT I_Positive_Integer, VT any](
	expected_num_inputs_raw KT,
	options []T_Option[KT, VT],
) (uint64, T_Performance_Profile) {
	expected_num_inputs_raw = next_power_of_two(expected_num_inputs_raw)
	expected_num_inputs := uint64(expected_num_inputs_raw)

	profile := PERFORMANCE_PROFILE__NORMAL
	for _, opt := range options {
		if opt.t == OPTION_TYPE__WITH_PERFORMANCE_PROFILE {
			profile = opt.other.(T_Performance_Profile)
		}
	}

	var num_buckets uint64
	switch profile {
	case PERFORMANCE_PROFILE__FAST:
		num_buckets = expected_num_inputs / 64
	case PERFORMANCE_PROFILE__NORMAL:
		num_buckets = expected_num_inputs / 1
	case PERFORMANCE_PROFILE__CONSERVE_MEMORY:
		num_buckets = expected_num_inputs / 256
	default:
		panic("Invalid performance profile.")
	}

	if num_buckets%2 != 0 {
		panic("numBuckets should be a multiple of 2.")
	}

	return num_buckets, profile
}
