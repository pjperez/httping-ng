# httping-ng

Fast, flexible HTTP ping tool for measuring web latency — like `ping`, but for HTTP.

[![Go](https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/github/license/pjperez/httping-ng)](./LICENSE)

---

## Features

- Measure HTTP GET latency with precision, including total time, response status, and size
- Multithreaded: use multiple workers for concurrency and load testing
- Adjustable frequency and number of pings with `-i` and `-count`
- Structured JSON output with full summary and percentiles for automation
- Visual latency histogram with adaptive bucket sizing (`--histogram`)
- Configurable User-Agent string for visibility and compliance
- Portable binaries for Linux, macOS, and Windows (via `make build-all`)
- Static build support for use in minimal container images or CI pipelines

> **Note:** `httping-ng` uses Go's standard HTTP client and therefore relies on the system's support for HTTP/1.1 and HTTP/2. It does **not** currently support HTTP/3 or QUIC.

---

## Install & Build

### Native build (for your system):

```bash
make
```

### Cross-compile all platforms:

```bash
make build-all
```

Binaries will be in the `build/` directory.

### Create release archives:

```bash
make release
```

Output will be in the `dist/` directory.

### Clean everything:

```bash
make clean
```

---

## CLI Usage

```bash
./httping-ng -url <target> [options]
```

### Examples

Ping google.com 10 times (default interval 1s):

```bash
./httping-ng -url google.com

[2025-04-20 18:41:22.407] [INFO] [client] No scheme provided, defaulting to https://google.com
[2025-04-20 18:41:22.407] [INFO] [client] Starting HTTP GET requests to https://google.com
HTTP GET https://google.com             worker=0  seq=0   status=200 size=17970B time=283.17 ms
HTTP GET https://google.com             worker=0  seq=1   status=200 size=18004B time=144.27 ms
HTTP GET https://google.com             worker=0  seq=2   status=200 size=17997B time=160.84 ms
HTTP GET https://google.com             worker=0  seq=3   status=200 size=18000B time=156.65 ms
HTTP GET https://google.com             worker=0  seq=4   status=200 size=17940B time=151.27 ms
HTTP GET https://google.com             worker=0  seq=5   status=200 size=18028B time=158.17 ms
HTTP GET https://google.com             worker=0  seq=6   status=200 size=17936B time=151.28 ms
HTTP GET https://google.com             worker=0  seq=7   status=200 size=17958B time=142.12 ms
HTTP GET https://google.com             worker=0  seq=8   status=200 size=17990B time=148.30 ms
HTTP GET https://google.com             worker=0  seq=9   status=200 size=18048B time=152.11 ms

--- https://google.com httping statistics ---
Requests : 10 sent, 10 received, 0.0% loss
RTT      : min=142.12 ms, avg=164.82 ms, max=283.17 ms

Latency percentiles:
p50: 152.11 ms
p75: 158.17 ms
p90: 283.17 ms
p99: 283.17 ms
```

Use 50 workers, 100 requests total, 100ms interval:

```bash
./httping-ng -url example.com -n 50 -count 100 -i 100
[2025-04-20 18:42:40.882] [INFO] [client] No scheme provided, defaulting to https://example.com
[2025-04-20 18:42:40.883] [INFO] [client] Starting HTTP GET requests to https://example.com
HTTP GET https://example.com            worker=48 seq=0   status=200 size=1256 B time=736.38 ms
HTTP GET https://example.com            worker=44 seq=0   status=200 size=1256 B time=736.28 ms
HTTP GET https://example.com            worker=29 seq=0   status=200 size=1256 B time=736.84 ms
HTTP GET https://example.com            worker=1  seq=0   status=200 size=1256 B time=736.67 ms
HTTP GET https://example.com            worker=36 seq=0   status=200 size=1256 B time=735.50 ms
HTTP GET https://example.com            worker=45 seq=0   status=200 size=1256 B time=736.61 ms
HTTP GET https://example.com            worker=25 seq=0   status=200 size=1256 B time=737.26 ms
HTTP GET https://example.com            worker=3  seq=0   status=200 size=1256 B time=737.45 ms

...


--- https://example.com httping statistics ---
Requests : 100 sent, 100 received, 0.0% loss
RTT      : min=198.94 ms, avg=492.55 ms, max=741.75 ms

Latency percentiles:
p50: 406.31 ms
p75: 737.97 ms
p90: 740.64 ms
p99: 741.02 ms
```

Print JSON output only:

```bash
./httping-ng -url google.com --json
{
  "target": "https://google.com",
  "total_sent": 10,
  "total_received": 10,
  "loss_percent": 0,
  "rtt_min_ms": 139.438,
  "rtt_avg_ms": 167.33599999999998,
  "rtt_max_ms": 309.414,
  "rtt_p50_ms": 154.285,
  "rtt_p75_ms": 156.603,
  "rtt_p90_ms": 309.414,
  "rtt_p99_ms": 309.414
}
```

Show latency histogram with 5 buckets:

```bash
./httping-ng -url cloudflare.com -i 100 -n 10 -count 50 -histogram --buckets 5
[2025-04-20 18:59:21.846] [INFO] [client] No scheme provided, defaulting to https://cloudflare.com
[2025-04-20 18:59:21.846] [INFO] [client] Starting HTTP GET requests to https://cloudflare.com
HTTP GET https://cloudflare.com         worker=0  seq=0   status=200 size=467121B time=1365.66 ms
HTTP GET https://cloudflare.com         worker=3  seq=0   status=200 size=467103B time=1681.24 ms
HTTP GET https://cloudflare.com         worker=2  seq=0   status=200 size=467099B time=2076.47 ms

...

HTTP GET https://cloudflare.com         worker=2  seq=5   status=200 size=467120B time=365.06 ms
HTTP GET https://cloudflare.com         worker=1  seq=6   status=200 size=467097B time=589.87 ms

--- https://cloudflare.com httping statistics ---
Requests : 50 sent, 50 received, 0.0% loss
RTT      : min=349.21 ms, avg=1257.35 ms, max=3603.22 ms

Latency histogram (exact 5 buckets):
   349–  1000 ms | ████████████████████████████████████████  30
  1000–  1651 ms | ██                                         2
  1651–  2301 ms | █████████████                             10
  2301–  2952 ms | █████                                      4
  2952–  3603 ms | █████                                      4
```


---

## Flags

| Flag             | Description                                 | Default             |
|------------------|---------------------------------------------|---------------------|
| `-url`          | Target URL or FQDN                          | *(required)*        |
| `-count`        | Total number of requests                    | `10`                |
|  `-n`            | Number of concurrent workers                | `10`                |
| `-i`             | Interval between requests (milliseconds)   | `1000`              |
| `-json`         | Output results as JSON                      | `false`             |
| `-histogram`    | Show latency histogram instead of percentiles | `false`           |
| `-buckets`      | Number of buckets in histogram              | `10`                |
| `-user-agent`   | Custom User-Agent string                    | `"httping-ng https://github.com/pjperez/httping-ng"` |

---

## Output

### CLI

```
HTTP GET https://example.com     worker=2  seq=3  status=200  size=1024B  time=132.57 ms
Request to https://example.com   worker=4  seq=6  failed: HTTP 503
```

### JSON (`--json`)

```json
{
  "target": "https://example.com",
  "total_sent": 10,
  "total_received": 9,
  "loss_percent": 10.0,
  "rtt_min_ms": 88.21,
  "rtt_avg_ms": 101.34,
  "rtt_max_ms": 110.92,
  "rtt_p50_ms": 99.87,
  "rtt_p75_ms": 105.22,
  "rtt_p90_ms": 110.00,
  "rtt_p99_ms": 110.92
}
```

---

## License

MIT — see [LICENSE](./LICENSE)

---

## Credits

Reimagined from [`pjperez/httping`](https://github.com/pjperez/httping).

Maintained by [@pjperez](https://github.com/pjperez)

