# TODO

- [x] Try align to cache slabs.
  - Test results are in:
    - Given the same entries per bucket, the `experimental_page_alignment` branch is faster than the `main` branch.
    - The `main` branch is still faster, and by a large margin, than the `experimental_page_alignment` branch when this is not taken into account.

- [x] Test for page faults.
  - It should be safe to say that there are minimal page faults in the `experimental_page_alignment` branch.
  - This is not the case for the `main` branch since we're dereferencing pointers that are all over the place.

## Focusing on fleshing out the `main` branch.

- [x] Experiment with generic types.
  - [x] also try straight up go's built-in generics.
    - Works great!

- [x] Optimize get requests.
- [x] Optimize set requests.

> Works fine.

- [ ] Experiment with a 'snapshot' mechanism.
  - This will take our pointers from all over the place, dereference them, and then collate them into a single contiguous block.
  - This (theoretically) will allow us to have much better cache performance.

- [x] Experiment with a queue mechanism such that we can keep the CPU busy when we get our turn from the go scheduler.
  - I could not get it to work as fast as i wanted it to.

- [x] Run tests for random read/write instead of just i++ read/write.
  - This will give us a better idea of how real-world performance will be.
  - TEST RESULTS: Im a god at programming.

### Optimization strategy A

- [ ] Experiment with a small array per bucket (type uint8) that will act as a mechanism to tell us if the mod2(key) is 0 or 1.

- If remainder is 0:

```go
for i := 0; i < num_entries; i += 2 {
  // Do something
}
```

- Otherwise, if remainder is 1:

```go
for i := 1; i < num_entries; i += 2 {
  // Do something
}
```

- You could also experiment with mod4, mod8, etc.

### Optimization strategy B

- Right now going through the buckets is exceedingly slow since they're not loaded into the cache most of the time.

- [ ] If strategy A doesn't work as expected, we can try using our small uint8 array to tell us roughly where the next bucket is.
  - This would rely on keeping the bucket entries sorted.

> For example:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func fast_find_starting_idx(diffs []uint32, test []*uint32, target uint32) int {
	i := 0
	x := uint32(0)
	for i < len(diffs) {
		current := *(test[i])
		x += diffs[i]
		x += current
		if x < target {
			i++
			continue
		} else if current+x == target {
			return i
		} else {
			return i * 8
		}
	}
	return -1
}

func generate_random_test(n uint32) []*uint32 {
	test := make([]*uint32, n)
	nums := make([]uint32, n)
	for i := uint32(0); i < n; i++ {
		nums[i] = i
		test[i] = &nums[i]
	}
	return test
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	test := generate_random_test(1<<32 - 16)
  // We don't need to sort the test since it's already sorted.
	//--sort.Slice(test, func(i, j int) bool {
	//--	return *(test[i]) < *(test[j])
	//--})
	fmt.Println("Test generated...")
	diffs := make([]uint32, len(test)/4)
	for i := 0; i < len(test)-1; i += 4 {
		a := *(test[i+1])
		b := *(test[i+2]) + *(test[i+3])
		diffs[i/4] = (a + b) - *(test[i])
	}
	fmt.Println("Diffs generated...")
	target := *(test[len(test)/2])

	var start time.Time

	n := 4

	// benchmark standard iteration...
	start = time.Now()
	for i := 0; i < n; i++ {
		for i := 0; i < len(test); i++ {
			if *(test[i]) == target {
				break
			}
		}
	}
	fmt.Println("Standard iteration time:", time.Since(start))

	// benchmark fast_find_starting_idx...
	start = time.Now()
	for i := 0; i < n; i++ {
		bench(diffs, test, target)
	}
	fmt.Println("Fast find starting idx time:", time.Since(start))
}
```

Given the above code, we get:

```text
Test generated...
Diffs generated...
Standard iteration time: 4.9252156s
Fast find starting idx time: 3.0098655s
```

NOTE: The overhead is only, and i mean ONLY, justified once the entries per bucket is large enough.

### Optimization strategy C

Right now a very large bottleneck occurs when the bucket size reaches some magnitude.

At 1 bucket per entry, we are going nearly as fast as we can.

> We are getting ~36x performance gain for get requests compared to the built-in map.
> This basically cuts in half every time we double the number of entries per bucket.

- [ ] To combat this, we can try loading a slice of the bucket into the cache instead of loading each individual entry.

## Test other kinds of metrics

- [x] Random read/write.
- [ ] Random read/write with a snapshot.
- [x] Test Big O Time Complexity for any random key.
  - It is safe to say that the time complexity is O(1) for the `main` branch.

## There is an issue where we can run out of RAM when using `SFDA_Resizable_Map`

If we have a large enough number of buckets, when we go to resize, because it goes to the next power of two, we can easily run out of RAM.

- [ ] Implement a secondary hash function that will work without needing a power of two size of buckets.
