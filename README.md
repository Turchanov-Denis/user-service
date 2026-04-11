My vps(1 core 4 ram ) test 
utils: https://github.com/rakyll/hey
```

PS C:\Users\Firo\Downloads> .\hey.exe -n 10000 -c 100 http://Firo.Firo.Firo.Firo:8080/user/3

Summary:
  Total:        33.8167 secs
  Slowest:      2.9362 secs
  Fastest:      0.1002 secs
  Average:      0.2824 secs
  Requests/sec: 295.7121


Response time histogram:
  0.100 [1]     |
  0.384 [8751]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.667 [824]   |■■■■
  0.951 [326]   |■
  1.235 [66]    |
  1.518 [22]    |
  1.802 [7]     |
  2.085 [2]     |
  2.369 [0]     |
  2.653 [0]     |
  2.936 [1]     |


Latency distribution:
  10%% in 0.1230 secs
  25%% in 0.2133 secs
  50%% in 0.2278 secs
  75%% in 0.2953 secs
  90%% in 0.5126 secs
  95%% in 0.6576 secs
  99%% in 0.9486 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0010 secs, 0.0000 secs, 0.1189 secs
  DNS-lookup:   0.0000 secs, 0.0000 secs, 0.0000 secs
  req write:    0.0000 secs, 0.0000 secs, 0.0098 secs
  resp wait:    0.1824 secs, 0.0827 secs, 1.6154 secs
  resp read:    0.0988 secs, 0.0000 secs, 2.7928 secs

Status code distribution:
  [200] 10000 responses

```