import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    stress_rps: {
      executor: 'ramping-arrival-rate',
      startRate: 100,                  // start: 100 req/s
      timeUnit: '1s',
      preAllocatedVUs: 50,             // start with 50 VUs ready
      maxVUs: 1000,                    // can grow to 1000 VUs if needed
      stages: [
        { duration: '1m', target: 100 },   // 100 req/s
        { duration: '1m', target: 500 },   // 500 req/s
        { duration: '1m', target: 1000 },  // 1000 req/s
        { duration: '1m', target: 2000 },  // 2000 req/s
        { duration: '30s', target: 0 },    // ramp down
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/ingredients');

  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}