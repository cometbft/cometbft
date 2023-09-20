# Set up a Grafana dashboard

These are the instructions to set up a Grafana server and dashboard connected to
a Prometheus server.

## Requirements

- `jq`
- `curl`
- Grafana server
    - To install Grafana, see https://grafana.com/docs/grafana/latest/setup-grafana/installation/
        - On macOS, install with `brew install grafana`, and start the server
          with `brew services start grafana`.
        - On Linux, follow your Linux distro installation instructions; to start the server,
          see https://grafana.com/docs/grafana/latest/setup-grafana/start-restart-grafana/.
    - The web interface is typically located in `http://localhost:3000/`.
    - The default username is 'admin' and password is 'admin'.
        - If the default credentials don't work, open a terminal and run 
        `sudo grafana-cli admin reset-admin-password` to reset the admin password.

## Set up API authentication

The following steps are needed in order to run the scripts below that interact
with Grafana via its HTTP API.

### Authentication

Run the following command to generate a new API token and write it directly into
the file `api_headers`, required by `curl` in the setup script. You will be
prompted for `admin`'s password.

    curl -s -u 'admin' -X POST -H "Content-Type: application/json" -d '{"name":"apikeycurl", "role": "Admin"}' http://localhost:3000/api/auth/keys | jq -r .key | xargs -I{} sed -i'' -e "s/Bearer.*/Bearer {}/" api_headers

For more info on generating tokens, see
[here](https://grafana.com/docs/grafana/latest/developers/http_api/create-api-tokens-for-org/);
or go to the Grafana web interface and [create an API token
manually](https://grafana.com/docs/grafana/latest/administration/service-accounts/#to-add-a-token-to-a-service-account).

### Other settings

If you need to change the default Grafana host (`http://localhost:3000`), edit
the `GRAFANA_HOST` variable in `setup.sh`

## Create data source and dashboard

The script `setup.sh` creates 
- a Grafana data source connected to an existing Prometheus server (if it
  doesn't exist), and
- a Grafana dashboard with the data source as a predefined parameter (if it
  doesn't exist). 

The full command is:

    ./setup.sh `[source-name]` `[source-url]` `[dashboard-title]`

where 
- `[source-name]` is a unique name for the data source (default: `PrometheusLocal`), 
- `[source-url]` is the URL of your Prometheus server (default: `http://localhost:9090`), and 
- `[title]` is a unique name for the dashboard (default: `CometBFT`).

You can also [create a data source via the web interface](https://grafana.com/docs/grafana/latest/administration/data-source-management/).
