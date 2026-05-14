import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-arrival-rate',
      startRate: 50,
      timeUnit: '1s',
      preAllocatedVUs: 500,   // ← changed: was 50, now 500
      maxVUs: 1000,
      stages: [
        { duration: '30s', target: 50 },    // normal
        { duration: '10s', target: 1500 },  // SPIKE! (also bigger now)
        { duration: '1m',  target: 1500 },  // stay high
        { duration: '10s', target: 50 },    // back to normal
        { duration: '30s', target: 50 },    // recovery time
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.20'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/ingredients');

  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}