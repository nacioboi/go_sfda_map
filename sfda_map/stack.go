/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package sfda_map

// import (
// 	"time"
// )

// type stack[T any] struct {
// 	_data *[]T
// }

// func new_stack[T any](size int) *stack[T] {
// 	data := make([]T, 0, size)
// 	return &stack[T]{
// 		_data: &data,
// 	}
// }

// func (s *stack[T]) push(v T) {
// 	*s._data = append(*s._data, v)
// }

// func (s *stack[T]) pop() T {
// 	for {
// 		if len(*s._data) > 0 {
// 			break
// 		}
// 		time.Sleep(1 * time.Millisecond)
// 	}
// 	v := (*s._data)[len(*s._data)-1]
// 	*s._data = (*s._data)[:len(*s._data)-1]
// 	return v
// }

// func (s *stack[T]) len() int {
// 	return len(*s._data)
// }
