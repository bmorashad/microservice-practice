export const k6SMLoadOptions = {
  discardResponseBodies: true,
  scenarios: {
    shared_iter_scenario: {
      executor: "shared-iterations",
      vus: 10,
      iterations: 100,
      startTime: "0s",
      gracefulStop: '5s',
    },
    per_vu_scenario: {
      executor: "per-vu-iterations",
      vus: 10,
      iterations: 10,
      // startTime: "10s",
    },
    ramping_vus: {
      executor: "ramping-vus",
      startvus: 0,
      stages: [
        { duration: "20s", target: 10 },
        { duration: "10s", target: 0 },
      ],
      gracefulRampDown: "1s",
      startTime: "10s",
    },
    constant_vus: {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
    },
    ramping_arrival_rate: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1m',
      preAllocatedVUs: 10,
      stages: [
        { target: 100, duration: '1m' },
        { target: 150, duration: '1.5m' },
        { target: 150, duration: '1.5m' },
        { target: 10, duration: '1m' },
        { target: 0, duration: '1m' },
      ],
      startTime: "20s",
    },
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s', // 1000 iterations per second, i.e. 1000 RPS
      duration: '30s',
      preAllocatedVUs: 100, // how large the initial pool of VUs would be
      maxVUs: 200, // if the preAllocatedVUs are not enough, we can initialize more
      startTime: "1m",
    },
  },
};
