# Fresh UserData register/login load run

Run on 2026-10-21 (Europe/Dublin), with new `RUN_ID`s and newly generated accounts. These are the results of the latest run only.

## Conditions and commands

- k6 2.3.0 with `k6/x/tcp` 0.3.1; 5 complete lifecycle iterations per VU.
- BCrypt remained Spring `BCryptPasswordEncoder()` default strength 10.
- Prometheus scraped every 5 seconds. The test phases lasted less than 10 seconds, so stage quantiles from `rate(...[10s])` are approximate and the adjacent phase boundary has some overlap.
- UserData cgroup: `cpu.max=max 100000`; `nr_throttled=0`; `throttled_usec=0`; PID max unlimited. Docker exposed 10 CPUs.

```sh
k6 run --scenario vus_25 --env START_TIME_25=0s --env ITERATIONS=5 --env RUN_ID=fresh_20261021_25 gateway/.loadtest/benchmark.js
k6 run --scenario vus_50 --env START_TIME_50=0s --env ITERATIONS=5 --env RUN_ID=fresh_20261021_50 gateway/.loadtest/benchmark.js
```

## Client results

| VUs | Window (UTC) | Iterations | HTTP requests | HTTP RPS | HTTP P95 | HTTP P99 | HTTP errors | TCP chat P95/P99 |
|---:|---|---:|---:|---:|---:|---:|---:|---:|
| 25 | 23:27:14–23:27:19 | 125 | 875 | 191.15 | 329.25 ms | 400.97 ms | 0% | 4.79/9.27 ms |
| 50 | 23:27:19–23:27:28 | 250 | 1,750 | 200.97 | 702.93 ms | 870.91 ms | 0% | 9/16.52 ms |

All k6 checks passed: 1,000/1,000 at 25 VUs and 2,000/2,000 at 50 VUs.

## Aligned server-side P95 observations

Prometheus stage P95 values below are approximated with a 10-second rate window at the first scrape after each run. The first sample is at 23:27:20 UTC for 25 VUs and at 23:27:30 UTC for 50 VUs. These scrape windows slightly straddle the adjacent phase boundary.

| VUs | BCrypt encode P95 | BCrypt verify/matches P95 | DB find P95 | DB save P95 | UserData gRPC login P95 | UserData gRPC registration P95 |
|---:|---:|---:|---:|---:|---:|---:|
| 25 | 332 ms | 320 ms | 11 ms | 29 ms | 320 ms | 347 ms |
| 50 | 805 ms | 776 ms | 12 ms | 43 ms | 779 ms | 833 ms |

At 50 VUs, BCrypt verify P95 was about 776 ms and login gRPC P95 about 779 ms, while DB lookup P95 was about 12 ms. Registration shows the same pattern: encode about 805 ms and gRPC about 833 ms. This attributes most of the measured UserData operation latency to BCrypt; the database is not the main contributor in this run.

## CPU and queue interpretation

At the 50-VU scrape just after completion (23:27:30 UTC), Prometheus reported `process_cpu_usage=98.5%`, `system_cpu_usage=99.5%`, `system_cpu_count=10`, and 102 live JVM threads. At 25 VUs, the test finished between scrapes; the scrape at the phase handoff showed 17.9% process CPU, so there is no reliable isolated peak for that short phase. Hikari pending connections stayed at zero in the sampled period. cgroup CPU throttling was zero and there was no CPU quota.

If another Grafana CPU panel shows under 20% for this same time, check its PromQL, time range, and any normalization (for example, division by `system_cpu_count`). The correlated dashboard's direct `100 * process_cpu_usage` series reached about 98.5% at the recorded 50-VU sample. Regardless of CPU display convention, BCrypt P95 near 0.8 seconds versus DB lookup near 0.012 seconds is sufficient to identify hashing as the dominant timed segment.

## Next controlled comparison

Keep VU counts, iteration counts, generated-account pattern, service build, database state shape, and Docker resources fixed. Change one factor per comparison (such as CPU resources or a separately reviewed BCrypt setting), repeat both 25 and 50 VU scenarios, and compare k6 P95/P99, RPS, errors, CPU, BCrypt/gRPC/HTTP P95, DB P95, and queue/thread samples. Do not reduce BCrypt strength without an explicit security review.
