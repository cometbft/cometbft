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
        - On Linux, follow your Linux distro installation instructions and to start the server
          see https://grafana.com/docs/grafana/latest/setup-grafana/start-restart-grafana/
    - The web interface is typically located in `http://localhost:3000/`.
    - The default username is 'admin' and password is 'admin'.
        - If the default credentials don't work, open a terminal and use the command below to reset
        the admin password
        `sudo grafana-cli admin reset-admin-password`

## Set up API authentication

Go to the Grafana web interface and [create an API token](https://grafana.com/docs/grafana/latest/administration/service-accounts/#to-add-a-token-to-a-service-account).

Briefly:
- Go to Administration > Service Accounts.
- Click on your service account (or create one if there is none).
- Ensure that the `Admin` role is selected for the account (you will need that to create a data source below)
- Click on the "Add token" button for the service account
    - After you click on `Generate token`, please ensure you copy it to the clipboard (you will need it on the next step)
- Edit the `config.sh` file
- Set the variables below:
    - GRAFANA_HOST="http://localhost:3000"
    - TOKEN=[the value of the token you copied to the clipboard]

>NOTE: All these steps are needed in order to run the following scripts that interact with Grafana via its HTTP API.

## Create a data source

To create a Grafana data source connected to your Prometheus server, run:

    ./datasource.sh `[source-name]` `[source-url]`

where `[source-name]` is a unique name for the data source and `[source-url]` is
the URL of your Prometheus server.

For example:

    ./datasource.sh PrometheusLocal http://localhost:9090

You can also [create a data source via the web interface](https://grafana.com/docs/grafana/latest/administration/data-source-management/).

## Import predefined dashboard

A predefined dashboard is located in `data/dashboards/comet_template.json`.

To import it, run:

    ./dashboard.sh [title] [datasource-name] [datasource-uid]

where `[title]` is a unique name for the dashboard, and `[datasource-name]` and
`[datasource-uid]` are the name and uid of the data source created in the
previous step.
