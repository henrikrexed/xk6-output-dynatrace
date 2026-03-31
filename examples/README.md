# Examples

Sample k6 scripts demonstrating how to use the xk6-output-dynatrace extension.

## Prerequisites

1. Build a k6 binary with the extension:

   ```bash
   xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest
   ```

2. Set the required environment variables:

   ```bash
   export K6_DYNATRACE_URL=https://<environmentid>.live.dynatrace.com
   export K6_DYNATRACE_APITOKEN=<token-with-metrics-ingest-v2-scope>
   ```

   The API token needs the **Ingest metrics** (`metrics.ingest`) scope.

## Scripts

| Script | Description |
|--------|-------------|
| [basic.js](basic.js) | Simple HTTP test sending metrics to Dynatrace via MINT protocol |
| [custom-tags.js](custom-tags.js) | Using k6 tags to add custom metric dimensions in Dynatrace |
| [otlp-mode.js](otlp-mode.js) | Sending metrics via OTLP/HTTP instead of MINT |

## Running

```bash
# MINT mode (default)
./k6 run examples/basic.js -o output-dynatrace

# OTLP mode
export K6_DYNATRACE_OUTPUT_MODE=otlp
./k6 run examples/otlp-mode.js -o output-dynatrace
```

## MINT vs OTLP

- **MINT** (default): Uses the Dynatrace Metrics Ingest (MINT) protocol via `/api/v2/metrics/ingest`. Lightweight and purpose-built for Dynatrace.
- **OTLP**: Uses OpenTelemetry Protocol over HTTP. Set `K6_DYNATRACE_OUTPUT_MODE=otlp` to enable. Useful when you want to align with an OpenTelemetry-based observability pipeline.

Both modes use the same `-o output-dynatrace` flag and the same `K6_DYNATRACE_URL` / `K6_DYNATRACE_APITOKEN` environment variables.
