// OTLP mode example: send k6 metrics to Dynatrace via OTLP/HTTP.
//
// Instead of the default MINT ingest protocol, this mode uses OpenTelemetry
// Protocol (OTLP) to export metrics. The output flag remains the same
// (-o output-dynatrace); the mode is selected via an environment variable.
//
// Prerequisites:
//   export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
//   export K6_DYNATRACE_APITOKEN=<token-with-metrics-ingest-v2-scope>
//   export K6_DYNATRACE_OUTPUT_MODE=otlp
//
// Build:
//   xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest
//
// Run:
//   ./k6 run examples/otlp-mode.js -o output-dynatrace
//
// The -o flag is identical to MINT mode. The K6_DYNATRACE_OUTPUT_MODE=otlp
// environment variable switches the transport from MINT to OTLP/HTTP.

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
