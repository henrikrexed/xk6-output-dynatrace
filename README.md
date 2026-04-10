# xk6-output-dynatrace

A [k6](https://k6.io/) extension that sends test-run metrics to [Dynatrace](https://www.dynatrace.com/) in real time. Supports both the Dynatrace Metrics Ingest (MINT) protocol and OpenTelemetry Protocol (OTLP/HTTP).

## Prerequisites

- A Dynatrace environment (SaaS or Managed)
- A Dynatrace API token with the **Ingest metrics** (`metrics.ingest`) scope
- [Go](https://go.dev/dl/) 1.25+ (only if building from source)

## Getting Started

There are three ways to use this extension: build a custom k6 binary with `xk6`, use the prebuilt Docker image, or add the extension to an existing k6 setup. Choose the approach that fits your workflow.

### Option 1 — Build a Custom k6 Binary with xk6

[xk6](https://github.com/grafana/xk6) is the official k6 extension bundler. It compiles k6 from source with one or more extensions baked in.

1. **Install xk6:**

   ```bash
   go install go.k6.io/xk6/cmd/xk6@latest
   ```

2. **Build k6 with the Dynatrace extension:**

   ```bash
   xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest
   ```

   > **Important:** Use capital `D` in `Dynatrace` — Go modules are case-sensitive. Using `github.com/dynatrace/...` (lowercase) will cause the build to fail.

   This produces a `k6` (or `k6.exe` on Windows) binary in the current directory.

3. **Set the required environment variables:**

   ```bash
   export K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com
   export K6_DYNATRACE_APITOKEN=<your-api-token>
   ```

4. **Run a test:**

   ```bash
   ./k6 run script.js -o output-dynatrace
   ```

### Option 2 — Use the Prebuilt Docker Image

A multi-architecture Docker image is provided so you can run tests without installing Go or xk6 locally.

1. **Build the image** (from the repository root):

   ```bash
   docker build -t k6-dynatrace .
   ```

2. **Run a test** with environment variables and your script mounted:

   ```bash
   docker run --rm \
     -e K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com \
     -e K6_DYNATRACE_APITOKEN=<your-api-token> \
     -v $(pwd)/script.js:/home/k6/script.js \
     k6-dynatrace run /home/k6/script.js -o output-dynatrace
   ```

   Mount additional scripts or data files as needed with extra `-v` flags.

### Option 3 — Add the Extension to an Existing k6 Build

If you already build k6 with other extensions, add this one alongside them:

```bash
xk6 build \
  --with github.com/Dynatrace/xk6-output-dynatrace@latest \
  --with github.com/grafana/xk6-dashboard@latest
```

Then run k6 with multiple outputs:

```bash
./k6 run script.js -o output-dynatrace -o dashboard
```

## Configuration

All configuration is done through environment variables. Set them before running k6.

### Required Variables

| Variable | Description |
|----------|-------------|
| `K6_DYNATRACE_URL` | Base URL of your Dynatrace environment (e.g. `https://abc12345.live.dynatrace.com`) |
| `K6_DYNATRACE_APITOKEN` | Dynatrace API token with the **Ingest metrics** (`metrics.ingest`) scope |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `K6_DYNATRACE_FLUSH_PERIOD` | `5s` | How often buffered metrics are sent to Dynatrace |
| `K6_DYNATRACE_INSECURE_SKIP_TLS_VERIFY` | `true` | Skip TLS certificate verification |
| `K6_CA_CERT_FILE` | *(none)* | Path to a custom CA certificate file |
| `K6_DYNATRACE_OUTPUT_MODE` | *(MINT)* | Set to `otlp` to send metrics via OTLP/HTTP instead of MINT |
| `K6_DYNATRACE_BATCH_SIZE` | `1000` | Number of metric lines per batch (max 1500) |
| `K6_DYNATRACE_MAX_CONCURRENT_EXPORTS` | `2` | Maximum parallel batch exports |
| `K6_KEEP_TAGS` | `true` | Forward k6 tags as metric dimensions |
| `K6_KEEP_NAME_TAG` | `false` | Include the `name` tag as a dimension |
| `K6_KEEP_URL_TAG` | `true` | Include the `url` tag as a dimension |

### Custom HTTP Headers

Set custom HTTP headers on metric ingest requests using the `K6_DYNATRACE_HEADER_` prefix:

```bash
export K6_DYNATRACE_HEADER_X_CUSTOM_HEADER=myvalue
```

This sets the HTTP header `X_CUSTOM_HEADER: myvalue` on all metric ingest requests.

> **Note:** These environment variables set HTTP request headers, **not** metric dimensions or tags. To add custom dimensions to your metrics in Dynatrace, use k6 tags in your test script instead (see [Custom Tags example](examples/custom-tags.js)).

## Output Modes

### MINT (Default)

The default mode uses the Dynatrace Metrics Ingest (MINT) protocol via `/api/v2/metrics/ingest`. It is lightweight and purpose-built for Dynatrace.

```bash
export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
export K6_DYNATRACE_APITOKEN=<your-token>
./k6 run script.js -o output-dynatrace
```

### OTLP

To send metrics via OpenTelemetry Protocol (OTLP/HTTP), set the output mode environment variable:

```bash
export K6_DYNATRACE_OUTPUT_MODE=otlp
./k6 run script.js -o output-dynatrace
```

The `-o` flag stays the same; only the transport changes. OTLP mode is useful when you want to align with an OpenTelemetry-based observability pipeline. Metrics are sent to `/api/v2/otlp/v1/metrics` with delta temporality.

## Examples

See the [examples/](examples/) directory for sample k6 scripts:

| Script | Description |
|--------|-------------|
| [basic.js](examples/basic.js) | Simple HTTP test sending metrics via MINT |
| [custom-tags.js](examples/custom-tags.js) | Using k6 tags to add custom metric dimensions |
| [otlp-mode.js](examples/otlp-mode.js) | Sending metrics via OTLP/HTTP |

### Quick Example

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

Run it:

```bash
export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
export K6_DYNATRACE_APITOKEN=<your-token>
./k6 run script.js -o output-dynatrace
```

## Sample Rate

k6 processes its outputs once per second. The default flush period for this extension is 5 seconds. The 26 built-in k6 metrics are collected at a 50ms rate, which means roughly 1000-1500 samples per flush period with the default (MINT) mapping. Custom metrics increase this estimate.

## License

[Apache-2.0](LICENSE)
