# xk6-output-dynatrace

A [k6](https://k6.io/) extension that sends test-run metrics to [Dynatrace](https://www.dynatrace.com/) in real time.

## What It Does

This extension plugs into k6's output pipeline and forwards every metric sample to your Dynatrace environment as it is collected. You get live visibility into your load tests directly in Dynatrace dashboards and alerting, without needing a separate metrics pipeline.

## Key Features

- **Two output modes** — send metrics via the native Dynatrace Metrics Ingest (MINT) protocol or via OpenTelemetry Protocol (OTLP/HTTP).
- **Zero-config metric mapping** — all 26 built-in k6 metrics are forwarded automatically with a `k6.` prefix.
- **Tag forwarding** — k6 tags become metric dimensions in Dynatrace, giving you the same filtering and grouping you use in k6 Cloud.
- **Batching and concurrency** — configurable batch size and parallel exports to handle high-throughput tests without back-pressure.
- **Docker-ready** — a multi-architecture Dockerfile is included so you can run tests without installing Go.

## Quick Start

```bash
# 1. Build k6 with the extension
xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest

# 2. Configure your Dynatrace environment
export K6_DYNATRACE_URL=https://<your-environment-id>.live.dynatrace.com
export K6_DYNATRACE_APITOKEN=<your-api-token>

# 3. Run a test
./k6 run script.js -o output-dynatrace
```

Head over to [Getting Started](getting-started.md) for the full walkthrough, or jump straight to [Configuration](configuration.md) if you already have a custom k6 binary.

## License

This project is licensed under the [Apache-2.0 License](https://github.com/Dynatrace/xk6-output-dynatrace/blob/master/LICENSE).
