// Custom tags example: use k6 tags to add metric dimensions in Dynatrace.
//
// Tags added to HTTP requests appear as dimensions on k6 metrics in Dynatrace.
// This is the correct way to add custom dimensions — NOT K6_DYNATRACE_HEADER_*
// environment variables, which set HTTP headers on the ingest request itself.
//
// Prerequisites:
//   export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
//   export K6_DYNATRACE_APITOKEN=<token-with-metrics-ingest-v2-scope>
//
// Run:
//   ./k6 run examples/custom-tags.js -o output-dynatrace

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 3,
  duration: '30s',
};

export default function () {
  // Tags are forwarded as metric dimensions to Dynatrace.
  const res = http.get('https://test-api.k6.io/public/crocodiles/', {
    tags: {
      appId: '12345',
      environment: 'staging',
    },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
