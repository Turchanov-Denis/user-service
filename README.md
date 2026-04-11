My Sweden vps(1 core 4 ram ) test 
utils: https://github.com/grafana/k6

```
choco install k6
```

```

import http from 'k6/http';

export default function () {
  let id = Math.floor(Math.random() * 1000) + 1;
  http.get(`http://Firo.Firo.Firo.Firo:8080/user/${id}`);
}

```

```
k6 run --vus 10 --iterations 1000 test.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: test.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 10m30s max duration (incl. graceful stop):
              * default: 1000 iterations shared among 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  █ TOTAL RESULTS

    HTTP
    http_req_duration..............: avg=141.1ms  min=83.19ms med=99.94ms max=1.29s p(90)=402.63ms p(95)=412.27ms
      { expected_response:true }...: avg=141.1ms  min=83.19ms med=99.94ms max=1.29s p(90)=402.63ms p(95)=412.27ms
    http_req_failed................: 0.00%  0 out of 1000
    http_reqs......................: 1000   69.539442/s

    EXECUTION
    iteration_duration.............: avg=142.13ms min=83.19ms med=99.97ms max=1.29s p(90)=402.63ms p(95)=412.27ms
    iterations.....................: 1000   69.539442/s
    vus............................: 10     min=10        max=10
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 836 kB 58 kB/s
    data_sent......................: 82 kB  5.7 kB/s




running (00m14.4s), 00/10 VUs, 1000 complete and 0 interrupted iterations

```

New test, preload use: (avg=141.1ms vs avg=130.3ms с preload)
```
k6 run --vus 10 --iterations 1000 test.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: test.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 10m30s max duration (incl. graceful stop):
              * default: 1000 iterations shared among 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  █ TOTAL RESULTS

    HTTP
    http_req_duration..............: avg=130.3ms  min=83.72ms med=99.45ms max=889.25ms p(90)=104.04ms p(95)=407.46ms
      { expected_response:true }...: avg=130.3ms  min=83.72ms med=99.45ms max=889.25ms p(90)=104.04ms p(95)=407.46ms
    http_req_failed................: 0.00%  0 out of 1000
    http_reqs......................: 1000   74.816154/s

    EXECUTION
    iteration_duration.............: avg=131.35ms min=83.72ms med=99.5ms  max=889.25ms p(90)=104.45ms p(95)=407.46ms
    iterations.....................: 1000   74.816154/s
    vus............................: 10     min=10        max=10
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 836 kB 63 kB/s
    data_sent......................: 82 kB  6.1 kB/s




running (00m13.4s), 00/10 VUs, 1000 complete and 0 interrupted iterations
default ✓ [======================================] 10 VUs  00m13.4s/10m0s  1000/1000 shared iters
```




SPB vps
```
k6 run --vus 10 --iterations 1000 test.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: test.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 10m30s max duration (incl. graceful stop):
              * default: 1000 iterations shared among 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  █ TOTAL RESULTS

    HTTP
    http_req_duration..............: avg=29.44ms min=27.81ms med=29.09ms max=37.28ms p(90)=30.32ms p(95)=31.81ms
      { expected_response:true }...: avg=29.44ms min=27.81ms med=29.09ms max=37.28ms p(90)=30.32ms p(95)=31.81ms
    http_req_failed................: 0.00%  0 out of 1000
    http_reqs......................: 1000   335.528225/s

    EXECUTION
    iteration_duration.............: avg=29.79ms min=27.86ms med=29.11ms max=63.98ms p(90)=30.32ms p(95)=32.84ms
    iterations.....................: 1000   335.528225/s
    vus............................: 10     min=10        max=10
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 836 kB 280 kB/s
    data_sent......................: 84 kB  28 kB/s
```