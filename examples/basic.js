// Basic example: send k6 metrics to Dynatrace via MINT protocol.
//
// Prerequisites:
//   export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
//   export K6_DYNATRACE_APITOKEN=<token-with-metrics-ingest-v2-scope>
//
// Build:
//   xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest
//
// Run:
//   ./k6 run examples/basic.js -o output-dynatrace

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 5,
  duration: '30s',
};

export default function () {
  const res = http.get('https://test-api.k6.io/public/crocodiles/');

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
