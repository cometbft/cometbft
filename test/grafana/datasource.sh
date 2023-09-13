#!/usr/bin/env bash

. "$(dirname "$0")/config.sh"

DATASOUCE_NAME=${1:-"PrometheusLocal"}
DATASOUCE_URL=${2:-"http://localhost:9090"}

datasource_list() {
    echo Data sources:
    curl_get $TOKEN "$GRAFANA_HOST/api/datasources" | jq -r '["UID","NAME","URL"], (.[] | [.uid,.name,.url]) | @tsv' | column -ts$'\t'
}

datasource_create() {
    DATA=$1
    echo Creating data source...
    DATASOURCE_UID=$(curl_post $TOKEN "$GRAFANA_HOST/api/datasources" "$DATA" | jq -r .datasource.uid)
    echo Created data source with uid $DATASOURCE_UID
}

datasource_delete() {
    UIDX=$1
    echo Deleting data source with uid $UIDX...
    curl_delete $TOKEN "$GRAFANA_HOST/api/datasources/uid/$UIDX" | jq
}

DATASOURCE_TEMPLATE_PATH="data/datasources/prometheus_template.json"
DATASOUCE_JSON=$(jq -c \
    --arg name "$DATASOUCE_NAME" \
    --arg url "$DATASOUCE_URL" \
    '.name = $name | .url = $url' \
    $DATASOURCE_TEMPLATE_PATH)
datasource_create "$DATASOUCE_JSON"

# datasource_list 
