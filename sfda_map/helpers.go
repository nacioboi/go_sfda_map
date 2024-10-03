/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

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
