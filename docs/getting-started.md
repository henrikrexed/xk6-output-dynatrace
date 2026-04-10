# Getting Started

This guide walks you through installing and running the xk6-output-dynatrace extension for the first time.

## Prerequisites

Before you begin, make sure you have:

- A **Dynatrace environment** (SaaS or Managed).
- A **Dynatrace API token** with the **Ingest metrics** (`metrics.ingest`) scope. You can create one in **Dynatrace > Access tokens**.
- **Go 1.25 or later** — only needed if you are building a custom k6 binary (not needed for Docker).

## Step 1 — Get a k6 Binary with the Extension

Choose one of the three methods below.

### Option A: Build with xk6

[xk6](https://github.com/grafana/xk6) is the official tool for building k6 with extensions.

1. Install xk6:

    ```bash
    go install go.k6.io/xk6/cmd/xk6@latest
    ```

2. Build k6 with the Dynatrace output extension:

    ```bash
    xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest
    ```

    !!! warning "Case-sensitive module path"
        Use a capital `D` in `Dynatrace`. Go modules are case-sensitive — `github.com/dynatrace/...` (lowercase) will fail to resolve.

    This produces a `k6` binary (or `k6.exe` on Windows) in the current directory.

### Option B: Use the Prebuilt Docker Image

A `Dockerfile` is included in the repository. Build and tag it:

```bash
git clone https://github.com/Dynatrace/xk6-output-dynatrace.git
cd xk6-output-dynatrace
docker build -t k6-dynatrace .
```

The image supports `linux/amd64` and `linux/arm64` via Docker BuildKit.

### Option C: Add to an Existing xk6 Build

If you already build k6 with other extensions, add this one with an extra `--with` flag:

```bash
xk6 build \
  --with github.com/Dynatrace/xk6-output-dynatrace@latest \
  --with github.com/grafana/xk6-dashboard@latest
```

## Step 2 — Configure Environment Variables

Two environment variables are required:

```bash
export K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com
export K6_DYNATRACE_APITOKEN=<your-api-token>
```

| Variable | Description |
|----------|-------------|
| `K6_DYNATRACE_URL` | Base URL of your Dynatrace environment. Do **not** include a trailing path — the extension appends the correct API endpoint automatically. |
| `K6_DYNATRACE_APITOKEN` | API token with the **Ingest metrics** scope. |

See [Configuration](configuration.md) for the full list of optional variables.

## Step 3 — Write a Test Script

Create a file called `script.js`:

```javascript
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

## Step 4 — Run the Test

=== "Custom Binary"

    ```bash
    ./k6 run script.js -o output-dynatrace
    ```

=== "Docker"

    ```bash
    docker run --rm \
      -e K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com \
      -e K6_DYNATRACE_APITOKEN=<your-api-token> \
      -v $(pwd)/script.js:/home/k6/script.js \
      k6-dynatrace run /home/k6/script.js -o output-dynatrace
    ```

The `-o output-dynatrace` flag tells k6 to use this extension as an output backend. Metrics are sent to Dynatrace every 5 seconds by default.

## Step 5 — View Metrics in Dynatrace

Once the test is running, open your Dynatrace environment and navigate to **Explore data** (or the Data Explorer). Search for metrics with the `k6.` prefix. You will see all 26 built-in k6 metrics:

- `k6.http_req_duration`
- `k6.http_reqs`
- `k6.vus`
- `k6.iterations`
- ... and more.

You can create dashboards, set alerts, and correlate k6 results with your application telemetry.

## Next Steps

- [Configuration](configuration.md) — tune flush period, batch size, TLS settings, and OTLP mode.
- [Examples](examples.md) — sample scripts for common use cases.
