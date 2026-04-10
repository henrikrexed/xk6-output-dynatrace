# Configuration

All configuration for xk6-output-dynatrace is done through environment variables. Set them before running k6.

## Required Variables

| Variable | Description |
|----------|-------------|
| `K6_DYNATRACE_URL` | Base URL of your Dynatrace environment (e.g. `https://abc12345.live.dynatrace.com`). The extension appends the correct API path automatically. |
| `K6_DYNATRACE_APITOKEN` | Dynatrace API token with the **Ingest metrics** (`metrics.ingest`) scope. |

## Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `K6_DYNATRACE_FLUSH_PERIOD` | `5s` | How often buffered metrics are flushed to Dynatrace. Accepts Go duration strings (e.g. `10s`, `1m`). |
| `K6_DYNATRACE_INSECURE_SKIP_TLS_VERIFY` | `true` | Set to `false` to enforce TLS certificate verification. |
| `K6_CA_CERT_FILE` | *(none)* | Path to a custom CA certificate file for TLS verification. |
| `K6_DYNATRACE_OUTPUT_MODE` | *(MINT)* | Set to `otlp` to send metrics via OTLP/HTTP instead of the default MINT protocol. |
| `K6_DYNATRACE_BATCH_SIZE` | `1000` | Number of metric lines per batch. Maximum is `1500`. |
| `K6_DYNATRACE_MAX_CONCURRENT_EXPORTS` | `2` | Maximum number of parallel batch exports. |
| `K6_KEEP_TAGS` | `true` | Forward k6 tags as metric dimensions in Dynatrace. |
| `K6_KEEP_NAME_TAG` | `false` | Include the `name` tag as a metric dimension. |
| `K6_KEEP_URL_TAG` | `true` | Include the `url` tag as a metric dimension. |

## Output Modes

### MINT (Default)

The default output mode uses the Dynatrace Metrics Ingest (MINT) protocol. Metrics are sent to `/api/v2/metrics/ingest` as plain-text metric lines.

MINT is lightweight, purpose-built for Dynatrace, and does not require any additional configuration beyond the two required variables.

```bash
export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
export K6_DYNATRACE_APITOKEN=<your-token>
./k6 run script.js -o output-dynatrace
```

### OTLP

To send metrics via OpenTelemetry Protocol (OTLP/HTTP), set:

```bash
export K6_DYNATRACE_OUTPUT_MODE=otlp
```

In OTLP mode, metrics are exported to `/api/v2/otlp/v1/metrics` with **delta temporality**. All other configuration variables remain the same.

OTLP mode is useful when:

- You want to standardize on OpenTelemetry across your observability stack.
- You are already using OTLP collectors or pipelines and want k6 metrics to follow the same path.

!!! note
    The `-o output-dynatrace` flag is the same for both modes. The `K6_DYNATRACE_OUTPUT_MODE` variable controls which transport is used.

## Custom HTTP Headers

You can set custom HTTP headers on metric ingest requests using environment variables with the `K6_DYNATRACE_HEADER_` prefix:

```bash
export K6_DYNATRACE_HEADER_X_CUSTOM_HEADER=myvalue
export K6_DYNATRACE_HEADER_X_ANOTHER=othervalue
```

Each variable sets one header. The part after `K6_DYNATRACE_HEADER_` becomes the header name, and the variable value becomes the header value.

!!! warning "Headers vs. Dimensions"
    These environment variables set HTTP headers on the **ingest request itself**, not metric dimensions or tags. To add custom dimensions to your metrics in Dynatrace, use [k6 tags](examples.md#custom-tags) in your test script instead.

## JSON Configuration

You can also pass configuration as JSON in a k6 options file or via the `--out` argument:

```json
{
  "url": "https://abc12345.live.dynatrace.com",
  "apitoken": "dt0c01.XXXXXXXX",
  "flushPeriod": "10s",
  "insecureSkipTLSVerify": false,
  "batchSize": 500,
  "keepTags": true
}
```

Configuration is applied in this order (later values override earlier ones):

1. Default values
2. JSON configuration
3. Environment variables
4. CLI argument values

## Tuning for High-Throughput Tests

For tests with many virtual users or custom metrics, consider adjusting:

- **`K6_DYNATRACE_FLUSH_PERIOD`** — a shorter flush period (e.g. `2s`) sends data more frequently but creates more HTTP requests.
- **`K6_DYNATRACE_BATCH_SIZE`** — increase up to `1500` to pack more metric lines per request.
- **`K6_DYNATRACE_MAX_CONCURRENT_EXPORTS`** — increase to allow more parallel batch sends if your Dynatrace environment can handle the load.

If you see warnings about remote write taking longer than the flush period, metrics may be dropped. In that case, increase the flush period or batch size.
