# The performance can only get better as the bucket size increases.

- **(1024\*1024\*1024)** = n = 22.7598663942453x faster:

```text
Built-in map set time: 5m43.5325362s
Built-in map get time: 1m19.4850689s
Checksum: 576460751766552576
SFDA map set time: 24.2210675s
SFDA map get time: 3.4923346s
Checksum: 576460751766552576
```

- **(1024\*1024\*8)** = n = 13.6967649627573x faster:

```text
Built-in map set time: 1.1782559s
Built-in map get time: 385.7913ms
Checksum: 35184367894528
SFDA map set time: 164.0045ms
SFDA map get time: 28.166ms
Checksum: 35184367894528
```
