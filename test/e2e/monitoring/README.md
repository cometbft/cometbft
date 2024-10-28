# Monitoring

Prometheus and Grafana server for E2E testnets.

## How to run

First, `prometheus.yml` must exist in this directory. For example, generate one by running from
`test/e2e`:
```bash
make fast
./build/runner -f networks/simple.toml setup
```

To start all monitoring services:
```bash
docker compose up -d
```

To stop all monitoring services:
```bash
docker compose down
```

## Details

This docker compose (`compose.yml`) creates a local Granafa and Prometheus server. It is useful for
local debugging and monitoring.

Prometheus will connect to the host machine's ports for data, as defined in `prometheus.yml`.

You can access the Grafana web interface at `http://localhost:3000` and the Prometheus web interface
at `http://localhost:9090`.

The default Grafana username and password is `admin`/`admin`. You will only need it if you want to
change something. The pre-loaded dashboards can be viewed without a password.

Data from Grafana and Prometheus end up in the `data-grafana` and `data-prometheus` folders on your
host machine. This allows you to stop and restart the servers without data loss. The folders are
excluded from Git.
