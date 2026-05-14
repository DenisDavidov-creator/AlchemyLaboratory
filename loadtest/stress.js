import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 20 },   // warm up: 20 users
    { duration: '2m', target: 50 },   // grow to 50 users
    { duration: '2m', target: 100 },  // grow to 100 users
    { duration: '2m', target: 200 },  // grow to 200 users
    { duration: '30s', target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],   // more relaxed: 500ms
    http_req_failed: ['rate<0.05'],     // allow up to 5% errors
  },
};

export default function () {
  const res = http.get('http://localhost:8080/ingredients');

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}