#!/usr/bin/env bash

GRAFANA_HOST="http://localhost:3000"
DATA_DIR="data"

### Parameters

DATASOURCE_NAME=${1:-"PrometheusLocal"}
DATASOURCE_URL=${2:-"http://localhost:9090"}
DASHBOARD_TITLE=${3:-"CometBFT"}

## Auxiliary functions

curl_get() {
    URL=$1
    curl -s -k -XGET -H @api_headers $URL
}

curl_post() {
    URL=$1
    DATA=$2
    curl -s -k -XPOST -H @api_headers --data-binary $DATA $URL
}

### Check connection

MESSAGE=$(curl_get "$GRAFANA_HOST/api/" | jq -r .message)
if [[ "$MESSAGE" = "Unauthorized" || "$MESSAGE" = "Invalid API key" ]]; then
    echo "Authentication error: $MESSAGE"
    exit 1
fi

### Data source set up

DATASOURCE_UID=$(curl_get "$GRAFANA_HOST/api/datasources/name/$DATASOURCE_NAME" | jq -r .uid)
if [[ -z "$DATASOURCE_UID" || "$DATASOURCE_UID" == "null" ]]; then
    DATASOURCE_TEMPLATE_PATH="$DATA_DIR/datasource_prometheus_template.json"
    DATASOURCE_JSON=$(jq -c \
        --arg name "$DATASOURCE_NAME" \
        --arg url "$DATASOURCE_URL" \
        '.name = $name | .url = $url' \
        $DATASOURCE_TEMPLATE_PATH)
    DATASOURCE_UID=$(curl_post "$GRAFANA_HOST/api/datasources" "$DATASOURCE_JSON" | jq -r .datasource.uid)
    echo "✅ Created data source with uid $DATASOURCE_UID"
else
    echo "⚠️  Data source $DATASOURCE_NAME already exists with uid $DATASOURCE_UID"
    echo "Data sources:"
    curl_get "$GRAFANA_HOST/api/datasources" | jq  -r '["UID","NAME","URL"], (.[] | [.uid,.name,.url]) | @tsv' | column -ts$'\t'
fi

### Dashboard set up

DASHBOARD_UID=$(curl_get "$GRAFANA_HOST/api/search" | jq -r --arg name "$DASHBOARD_TITLE" '.[] | select(.title==$name) | .uid')
if [ -z "$DASHBOARD_UID" ]; then
    DASHBOARD_JSON_PATH=$(mktemp)
    jq -c \
        --arg title "$DASHBOARD_TITLE" \
        --arg sourcename "$DATASOURCE_NAME" \
        --arg sourceuid "$DATASOURCE_UID" \
        '.dashboard.title = $title | .dashboard.templating.list[0].current |= (.text = $sourcename | .value = $sourceuid)' \
        "$DATA_DIR/dashboard_comet_template.json" > $DASHBOARD_JSON_PATH
    DASHBOARD_UID=$(curl_post "$GRAFANA_HOST/api/dashboards/db" "@$DASHBOARD_JSON_PATH" | jq -r .uid) 
    echo "✅ Created dashboard with uid $DASHBOARD_UID, URL: $GRAFANA_HOST/d/$DASHBOARD_UID/$DASHBOARD_TITLE"
else
    echo "⚠️  Dashboard $DASHBOARD_TITLE already exists"
    echo "Dashboards:"
    curl_get "$GRAFANA_HOST/api/search?query=&" | jq -r "if type==\"array\" then .[] else . end| .uri"
fi
