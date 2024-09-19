/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

type T_Option_Type uint8

const (
	OPTION_TYPE__WITH_HASH_FUNC T_Option_Type = iota
	OPTION_TYPE__WITH_PERFORMANCE_PROFILE
	OPTION_TYPE__WITH_EXPERIMENTAL_BATCHED_GETS
)

type I_SFDA_Map[KT I_Positive_Integer, VT any] interface{}

type T_Option[KT I_Positive_Integer, VT any] struct {
	t     T_Option_Type
	f     func(I_SFDA_Map[KT, VT])
	other interface{}
}

func With_Hash_Func[KT I_Positive_Integer, VT any](new_hf func(KT) uint64) T_Option[KT, VT] {
	return T_Option[KT, VT]{
		t: OPTION_TYPE__WITH_HASH_FUNC,
		f: func(m I_SFDA_Map[KT, VT]) {
			switch m.(type) {
			case *SFDA_Map[KT, VT]:
				m.(*SFDA_Map[KT, VT]).users_chosen_hash_func = new_hf
				m.(*SFDA_Map[KT, VT]).using_users_hash_func = true
			// case *SFDA_Aligned_Map[KT, VT]:
			// 	m.(*SFDA_Aligned_Map[KT, VT]).users_chosen_hash_func = new_hf
			// 	m.(*SFDA_Aligned_Map[KT, VT]).using_users_hash_func = true
			default:
				panic("Invalid map type.")
			}
		},
	}
}

type T_Performance_Profile uint8

const (
	// Normal performance profile is the default.
	//
	// It attempts to balance performance and memory usage.
	PERFORMANCE_PROFILE__NORMAL T_Performance_Profile = iota

	// Fast performance profile sacrifices memory usage for maximum performance.
	PERFORMANCE_PROFILE__FAST

	// Conserve memory performance has no special optimizations.
	//
	// Meaning, it uses an entire page for the key-value pairs.
	PERFORMANCE_PROFILE__CONSERVE_MEMORY
)

func With_Performance_Profile[KT I_Positive_Integer, VT any](p T_Performance_Profile) T_Option[KT, VT] {
	return T_Option[KT, VT]{
		t: OPTION_TYPE__WITH_PERFORMANCE_PROFILE,
		f: func(m I_SFDA_Map[KT, VT]) {
			switch m.(type) {
			case *SFDA_Map[KT, VT]:
				m.(*SFDA_Map[KT, VT]).performance_profile = p
			// case *SFDA_Aligned_Map[KT, VT]:
			// 	m.(*SFDA_Aligned_Map[KT, VT]).performance_profile = p
			default:
				panic("Invalid map type.")
			}
		},
		other: p,
	}
}
