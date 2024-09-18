# The performance can only get better as the bucket size increases.

- **(1024\*1024\*8)** = n = 32x faster for linear get:

> 2.10214455917395x faster for random get per op.

```text
❯ go run .\main.go
Built-in Map Microseconds  ::: LINEAR SET ::: 1364256
Built-in Map Microseconds  ::: LINEAR GET ::: 432671
Checksum: 35184367894528
SFDA Map Microseconds      ::: LINEAR SET ::: 60403
SFDA Map Microseconds      ::: LINEAR GET ::: 13415
Checksum: 35184367894528

Built-in Map Microseconds  ::: RANDOM GET PER OP ::: 0.052932
SFDA Map Microseconds      ::: RANDOM GET PER OP ::: 0.025180

Builtin Map Microseconds   ::: DELETE ::: 655616
SFDA Map Microseconds      ::: DELETE ::: 602211

Memory Used (Built-in):  41,128 bytes
Memory Used (SFDA):      86,000 bytes

SFDA Resizable Map Microseconds      ::: LINEAR SET ::: 5826063
SFDA Resizable Map Microseconds      ::: LINEAR GET ::: 29706
Checksum: 35180333515576
```

- **(1024\*1024\*768)** = n = 57x faster for linear get:

> 2.05405405405405 faster for random get per op.

```text
❯ go run .\main.go
Built-in Map Microseconds  ::: LINEAR SET ::: 171001398
Built-in Map Microseconds  ::: LINEAR GET ::: 68579949
Checksum: 324259172768022528
SFDA Map Microseconds      ::: LINEAR SET ::: 17409698
SFDA Map Microseconds      ::: LINEAR GET ::: 1195170
Checksum: 324259172768022528

Built-in Map Microseconds  ::: RANDOM GET PER OP ::: 0.103056
SFDA Map Microseconds      ::: RANDOM GET PER OP ::: 0.050172

Builtin Map Microseconds   ::: DELETE ::: 109759052
SFDA Map Microseconds      ::: DELETE ::: 107128885

Memory Used (Built-in):  41,128 bytes
Memory Used (SFDA):      85,920 bytes
```

A few things to note between the two:

- One fantastic thing to note is that the performance of the `SFDA_Map` (linear get) is much faster than the previous run.
  - This could be due to the fact that the runtime savings per run, compared to the builtin map that is; compounds over a larger data size.
- All stats are slower overall since the data size is much larger.
- The average random per op is also slower but the ratio is still the same.
- The memory stays the same since we are using a different data size and it stayed the same between both runs.
- And finally, the performance of delete has got much slower for SFDA Map, yes, even the ratio is worse.
  - This could be due to the cache performance of the `SFDA_Map` being worse than the built-in map.
