# I TRIED

On windows at least, and since golang doesn't give as much power as something like C:

I have realized that trying to page align anything in this project is near impossible as it introduces overhead.

Trying to page align will also do the following:

- Ruins readability,
- Makes the code more complex,
- Makes the code less maintainable,
- Makes the code less dynamic as you'll need a concrete page array (not slice, a constant size contiguous array) to store the page aligned data. This array is the page aligned data, and this means you'll need to dereference more which go really doesn't like.

## No.1 take away

Computers are stupid fast, and i mean ridiculously fast.

The overhead of dereferencing a pointer (cache miss) is about 1000x slower than a cache hit.

Dereferencing only is fast when we get cache just perfect, which is difficult to do, especially when there are thousands of CPUs out there.
