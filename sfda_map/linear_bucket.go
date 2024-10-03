package sfda_map

const c_NUM_ENTRIES = 1

func linear_bucket__set[KT I_Positive_Integer, VT any](
	keys *[c_NUM_ENTRIES]KT,
	values *[c_NUM_ENTRIES]VT,
	key KT,
	value *VT,
) {
	for i := 0; i < c_NUM_ENTRIES; i++ {
		if keys[i] == key {
			// We must ensure that we don't already have the same key contained within the same bucket...
			//m.setter__lazy_safety_check_queue <- t__setter__safety_check_params[KT, VT]{buck: b_ptr, key: key}
			(*values)[i] = *value
			return
		}
	}

	for i := 0; i < c_NUM_ENTRIES; i++ {
		if keys[i] == 0 {
			(*keys)[i] = key
			(*values)[i] = *value
			return
		}
	}

	panic("Bucket is full.")
}

func linear_bucket__get[KT I_Positive_Integer, VT any](
	keys [c_NUM_ENTRIES]KT,
	values *[c_NUM_ENTRIES]VT,
	key KT,
) T_Get_Result[VT] {
	// If number of buckets is chosen correctly, most of the time we are using the first entry...
	if keys[0] == key {
		return T_Get_Result[VT]{
			Value:       values[0],
			Did_Succeed: true,
		}
	}

	for i := 1; i < c_NUM_ENTRIES; i++ {
		if keys[i] == key {
			return T_Get_Result[VT]{
				Value:       values[i],
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
