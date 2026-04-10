# Examples

Sample k6 scripts demonstrating common use cases with the xk6-output-dynatrace extension.

All examples assume you have already built a k6 binary with the extension and set the required environment variables:

```bash
export K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com
export K6_DYNATRACE_APITOKEN=<your-api-token>
```

## Basic HTTP Test

The simplest example: run a load test and send all metrics to Dynatrace via the default MINT protocol.

```javascript title="examples/basic.js"
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
```

Run it:

```bash
./k6 run examples/basic.js -o output-dynatrace
```

## Custom Tags

Use k6 tags to add custom dimensions to your metrics in Dynatrace. Tags added to HTTP requests appear as dimensions on the corresponding k6 metrics, enabling filtering and grouping in dashboards.

```javascript title="examples/custom-tags.js"
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 3,
  duration: '30s',
};

export default function () {
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
```

Run it:

```bash
./k6 run examples/custom-tags.js -o output-dynatrace
```

In Dynatrace, you can filter metrics like `k6.http_req_duration` by `appId` or `environment` to compare results across application components or deployment stages.

!!! tip "Tags vs. Custom Headers"
    k6 tags become **metric dimensions** in Dynatrace. The `K6_DYNATRACE_HEADER_*` environment variables set **HTTP headers** on the ingest request — they are not the same thing. Use tags for data you want to query in Dynatrace dashboards.

## OTLP Mode

Send metrics via OpenTelemetry Protocol (OTLP/HTTP) instead of MINT. The test script is identical; only the environment variable changes.

```javascript title="examples/otlp-mode.js"
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
```

Run it:

```bash
export K6_DYNATRACE_OUTPUT_MODE=otlp
./k6 run examples/otlp-mode.js -o output-dynatrace
```

The `-o output-dynatrace` flag is the same for both modes. The `K6_DYNATRACE_OUTPUT_MODE=otlp` environment variable switches the transport from MINT to OTLP/HTTP.

## Using Docker

You can run any of the above examples using Docker instead of a custom binary:

```bash
docker run --rm \
  -e K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com \
  -e K6_DYNATRACE_APITOKEN=<your-api-token> \
  -v $(pwd)/examples:/home/k6/examples \
  k6-dynatrace run /home/k6/examples/basic.js -o output-dynatrace
```

For OTLP mode with Docker:

```bash
docker run --rm \
  -e K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com \
  -e K6_DYNATRACE_APITOKEN=<your-api-token> \
  -e K6_DYNATRACE_OUTPUT_MODE=otlp \
  -v $(pwd)/examples:/home/k6/examples \
  k6-dynatrace run /home/k6/examples/otlp-mode.js -o output-dynatrace
```

## Combining with Other Outputs

k6 supports multiple outputs. You can send metrics to Dynatrace and another backend simultaneously:

```bash
xk6 build \
  --with github.com/Dynatrace/xk6-output-dynatrace@latest \
  --with github.com/grafana/xk6-dashboard@latest

./k6 run script.js -o output-dynatrace -o dashboard
```

## Real-World Load Test

For a more realistic workload example that simulates user flows (browsing, adding to cart, checkout), see the [`loadgenerator.js`](https://github.com/Dynatrace/xk6-output-dynatrace/blob/master/loadgenerator.js) script in the repository root. It demonstrates:

- Multiple endpoint patterns
- Custom error counters
- Randomized user behavior
- Longer duration tests with higher VU counts
