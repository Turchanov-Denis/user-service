branch(gRPC):

utils: https://github.com/grafana/k6

```
choco install k6
```

usage 

```
//$env:API_URL="your_ip:port"; k6 run --vus 10 --iterations 1000 testgRPC.js PowerShell
//set API_URL=your_ip:port && k6 run --vus 10 --iterations 1000 testgRPC.js CMD
```
My Sweden vps(1 core 4 ram ) test
```
PS C:\Users\Firo\GolandProjects\user-service\k6> k6 run --vus 10 --iterations 1000 testgRPC.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local                                                                                                                                               
        script: testgRPC.js                                                                                                                                         
        output: -                                                                                                                                                   

     scenarios: (100.00%) 1 scenario, 10 max VUs, 10m30s max duration (incl. graceful stop):                                                                        
              * default: 1000 iterations shared among 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  █ TOTAL RESULTS

    checks_total.......: 1000    62.089363/s
    checks_succeeded...: 100.00% 1000 out of 1000
    checks_failed......: 0.00%   0 out of 1000

    ✓ status is OK                                                                                                                                                  

    EXECUTION
    iteration_duration...: avg=159.5ms  min=85.22ms med=101.34ms max=4.81s p(90)=404.48ms p(95)=541.46ms                                                            
    iterations...........: 1000   62.089363/s
    vus..................: 2      min=2       max=10
    vus_max..............: 10     min=10      max=10

    NETWORK
    data_received........: 618 kB 38 kB/s
    data_sent............: 85 kB  5.3 kB/s

    GRPC
    grpc_req_duration....: avg=157.23ms min=85.22ms med=101.2ms  max=4.81s p(90)=403.67ms p(95)=536.85ms 
```

SPB vps
```
PS C:\Users\Firo\GolandProjects\user-service\k6> k6 run --vus 10 --iterations 1000 testgRPC.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: testgRPC.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 10m30s max duration (incl. graceful stop):                                                                        
              * default: 1000 iterations shared among 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  █ TOTAL RESULTS

    checks_total.......: 1000    333.267424/s
    checks_succeeded...: 100.00% 1000 out of 1000
    checks_failed......: 0.00%   0 out of 1000

    ✓ status is OK                                                                                                                                                  

    EXECUTION
    iteration_duration...: avg=29.91ms min=27.55ms med=28.99ms max=109.97ms p(90)=29.99ms p(95)=30.74ms                                                             
    iterations...........: 1000   333.267424/s
    vus..................: 4      min=4        max=10
    vus_max..............: 10     min=10       max=10

    NETWORK
    data_received........: 618 kB 206 kB/s
    data_sent............: 85 kB  28 kB/s

    GRPC
    grpc_req_duration....: avg=29.09ms min=27.37ms med=28.89ms max=37.43ms  p(90)=29.89ms p(95)=30.41ms    
```
