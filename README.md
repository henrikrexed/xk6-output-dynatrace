
# xk6-output-dynatrace
k6 extension for publishing test-run metrics to Dynatrace 

Invented by: github.com/henrikrexed


### Usage

To build k6 binary with the Prometheus remote write output extension use:
```
xk6 build --with github.com/Dynatrace/xk6-output-dynatrace@latest 
```

Then run new k6 binary with:
```
export K6_DYNATRACE_URL=http://<environmentid>.live.dynatrace.com 
export K6_DYNATRACE_APITOKEN=<Dynatrace API token>
the api token needs to have the scope: metric ingest v2
./k6 run script.js -o output-dynatrace
```


### Custom HTTP Headers

You can set custom HTTP headers on requests sent to the Dynatrace API using environment variables with the `K6_DYNATRACE_HEADER_` prefix:

```
export K6_DYNATRACE_HEADER_X_CUSTOM_HEADER=myvalue
```

This sets the HTTP header `X_CUSTOM_HEADER: myvalue` on metric ingest requests.

> **Note:** These environment variables set HTTP request headers, **not** metric dimensions or tags. If you want custom dimensions on your metrics in Dynatrace, use k6 tags in your test script instead:
>
> ```javascript
> import http from 'k6/http';
>
> export default function () {
>   http.get('https://example.com', {
>     tags: { appId: '12345', appName: 'My App' },
>   });
> }
> ```
>
> Tags added this way will appear as dimensions on the corresponding k6 metrics in Dynatrace.

### On sample rate

k6 processes its outputs once per second and that is also a default flush period in this extension. The number of k6 builtin metrics is 26 and they are collected at the rate of 50ms. In practice it means that there will be around 1000-1500 samples on average per each flush period in case of raw mapping. If custom metrics are configured, that estimate will have to be adjusted.


