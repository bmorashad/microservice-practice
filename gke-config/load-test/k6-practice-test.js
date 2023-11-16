import http from 'k6/http'

export const options = {
  discardResponseBodies: true,

  scenarios: {
  /* 100 iterations will be shared across 10vus
   * lasting for 5 seconds after     
   * completion before terminating gracefully.*/
    // shared_iter_scenario: {
    //   executor: "shared-iterations",
    //   vus: 10,
    //   iterations: 100,
    //   startTime: "0s",
    //   gracefulStop: '5s',
    //   // maxDuration: '10s',
    // },
  /* 10 iterations will be executed by each 10vus */
    // per_vu_scenario: {
    //   executor: "per-vu-iterations",
    //   vus: 10,
    //   iterations: 10,
    //   // startTime: "10s",
    // },
  /* 100 iterations will be shared across 10vus
   * lasting for 5 seconds after     
   * completion before terminating gracefully.*/
    // ramping_vus: {
    //   executor: "ramping-vus",
    //   startvus: 0,
    //   stages: [
    //     { duration: "20s", target: 10 },
    //     { duration: "10s", target: 0 },
    //   ],
    //   gracefulRampDown: "1s",
    //   startTime: "10s",
    // },
    // constant_vus: {
    //   executor: 'constant-vus',
    //   vus: 10,
    //   duration: '30s',
    // },
    ramping_arrival_rate: {
      executor: 'ramping-arrival-rate',
      startRate: 300,
      timeUnit: '1m',
      preAllocatedVUs: 50,
      stages: [
        { target: 100, duration: '1m' },
        { target: 150, duration: '2m' },
        { target: 200, duration: '2m' },
        { target: 60, duration: '2m' },
        { target: 0, duration: '1m' },
      ],
      startTime: "20s",
    }
  },
};


export default function() {
  let baseUrl = 'http://google.com';
  let response = http.get(baseUrl);
}
