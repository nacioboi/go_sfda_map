package sfda_map

type linked_list[VT any] struct {
	head *node[VT]
	tail *node[VT]
}

type node[VT any] struct {
	value VT
	next  *node[VT]
}

func (ll *linked_list[VT]) append(value VT) {

	new_node := &node[VT]{value: value}

	if ll.head == nil {
		ll.head = new_node
		ll.tail = new_node
		return
	}

	ll.tail.next = new_node
	ll.tail = new_node
}

func (ll *linked_list[VT]) iter(f func(VT) bool) (VT, bool) {
	current_node := ll.head
	for current_node != nil {
		if f(current_node.value) {
			return current_node.value, true
		}
		current_node = current_node.next
	}
	var zero VT
	return zero, false
}
