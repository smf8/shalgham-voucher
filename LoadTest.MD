# Load Test

The test was done on Microsoft Azure with a 2 node kubernetes setup and pre-configured Redis and PostgreSQL.
System Specifications:
- Database (PostreSQL)
  - 4 vCPU, 16GB Memory, Same Region(Network)
  - 100 connection / pod
- Redis
  - Single Instance setup (not sentinel)
  - Disk persist enabled
  - 2 vCPU, 6GB Memory, Same Region(Network)
  - 100 connection / pod
- Wallet
  - (250m, 500Mi) (500m, 1Gi) request and limit resource respectively
  - 3 pods
  - autoscale disabled
- Voucher
  - (250m, 500Mi) (500m, 1Gi) request and limit resource respectively
  - 5 pods
  - autoscale disabled

### Create profile

To test wallet service and database stress in profile creation scneario a k6 load test with following parameters:
```javascript
create_profiles: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      preAllocatedVUs: 250,
      maxVUs: 1200,
      stages: [
        {target: 10, duration: '30s'},
        {target: 60, duration: '30s'},
        {target: 200, duration: '20s'},
        {target: 1000, duration: '60s'},
      ],
      exec: 'create_profile',
      startTime: '0s'
    }
```
Result:
```text

running (2m20.1s), 0000/0250 VUs, 39946 complete and 0 interrupted iterations
create_profiles ✓ [======================================] 0000/0250 VUs  2m20s  0999.94 iters/s

     ✓ is status 200

     checks.........................: 100.00% ✓ 39946      ✗ 0
     data_received..................: 6.6 MB  47 kB/s
     data_sent......................: 7.6 MB  54 kB/s
     http_req_blocked...............: avg=336.89µs min=1.3µs   med=2.5µs   max=81.44ms  p(90)=5.3µs   p(95)=12.3µs
     http_req_connecting............: avg=332.78µs min=0s      med=0s      max=81.32ms  p(90)=0s      p(95)=0s
     http_req_duration..............: avg=58.9ms   min=52.23ms med=58.29ms max=368.21ms p(90)=63.61ms p(95)=65.16ms
       { expected_response:true }...: avg=58.9ms   min=52.23ms med=58.29ms max=368.21ms p(90)=63.61ms p(95)=65.16ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 39946
     http_req_receiving.............: avg=31.77µs  min=11.2µs  med=21.4µs  max=5.02ms   p(90)=50.1µs  p(95)=86µs
     http_req_sending...............: avg=16.52µs  min=6.1µs   med=11.2µs  max=6.52ms   p(90)=21.4µs  p(95)=28.9µs
     http_req_tls_handshaking.......: avg=0s       min=0s      med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=58.85ms  min=52.2ms  med=58.25ms max=368.18ms p(90)=63.56ms p(95)=65.11ms
     http_reqs......................: 39946   285.204936/s
     iteration_duration.............: avg=59.37ms  min=52.34ms med=58.46ms max=368.34ms p(90)=63.9ms  p(95)=65.55ms
     iterations.....................: 39946   285.204936/s
     vus............................: 250     min=250      max=250
     vus_max........................: 250     min=250      max=250

```
### Redeem Voucher
We test with 
```javascript
create_voucher: {
      executor: 'shared-iterations',
      vus: 50,
      iterations: 500,
      startTime: '0s',
      exec: 'create_voucher',
    },
    redeem_voucher: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      preAllocatedVUs: 4000,
      maxVUs: 5000,
      stages: [
        {target: 100, duration: '20s'},
        {target: 1500, duration: '10s'},
        {target: 300, duration: '10s'},
        {target: 3000, duration: '10s'},
        {target: 5000, duration: '10s'},
        {target: 100, duration: '30s'},
      ],
      exec: 'redeem_voucher',
      startTime: '30s'
    },
```

We first create 50 vouchers with 2000 limit and random amount value. the `redeem_voucher` scenario includes 2 spikes in request. 
First spike is from `100rps -> 1500rps` in 15 seconds. the second one is `300 -> 3000 -> 5000` in 2 stages in 20 seconds.

```text
running (02m04.9s), 0000/5050 VUs, 86970 complete and 0 interrupted iterations
create_voucher ✓ [======================================] 50 VUs         00m00.7s/10m0s  500/500 shared iters
redeem_voucher ✓ [======================================] 0000/5000 VUs  1m30s           0110.10 iters/s

     ✓ is status 201
     ✗ is status 200
      ↳  99% — ✓ 86428 / ✗ 42

     checks.........................: 99.95% ✓ 86928     ✗ 42
     data_received..................: 14 MB  115 kB/s
     data_sent......................: 18 MB  145 kB/s
     dropped_iterations.............: 65523  524.73395/s
     http_req_blocked...............: avg=3.16ms  min=700ns   med=1.7µs    max=1.05s   p(90)=4.4µs   p(95)=49.88ms
     http_req_connecting............: avg=3.16ms  min=0s      med=0s       max=1.05s   p(90)=0s      p(95)=49.83ms
     http_req_duration..............: avg=2.5s    min=54.01ms med=908.3ms  max=26.05s  p(90)=7.25s   p(95)=9.56s
       { expected_response:true }...: avg=2.49s   min=54.01ms med=907.73ms max=26.05s  p(90)=7.24s   p(95)=9.55s
     http_req_failed................: 0.04%  ✓ 42        ✗ 86928
     http_req_receiving.............: avg=32.88µs min=9.4µs   med=22.8µs   max=2.33ms  p(90)=59.19µs p(95)=77.29µs
     http_req_sending...............: avg=23.34µs min=4.7µs   med=12.99µs  max=18.78ms p(90)=36.9µs  p(95)=58.99µs
     http_req_tls_handshaking.......: avg=0s      min=0s      med=0s       max=0s      p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=2.5s    min=53.97ms med=908.25ms max=26.05s  p(90)=7.25s   p(95)=9.56s
     http_reqs......................: 86970  696.48996/s
     iteration_duration.............: avg=2.5s    min=54.09ms med=909.15ms max=26.05s  p(90)=7.25s   p(95)=9.56s
     iterations.....................: 86970  696.48996/s
     vus............................: 5000   min=0       max=5000
     vus_max........................: 5050   min=2033    max=5050
```

Result notes:
- Considering less than 2 cores for each service. high latency is expected in peak loads. 
- Among 42 failed requests. 17 requests failed logically (duplicate entry) and others **timed out connecting** to wallet service (TCP connection).
- The system kept consistency for all 200 requests.