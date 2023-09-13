#!/usr/bin/env bash

. "$(dirname "$0")/config.sh"

DASHBOARD_TITLE=${1:-"CometBFT"}
DATASOUCE_NAME=$2
DATASOUCE_UID=$3

dashboard_list() {
    echo Dashboards:
    curl_get $TOKEN "$GRAFANA_HOST/api/search?query=&" | jq -r "if type==\"array\" then .[] else . end| .uri"
}

# creates or updates a dashboard
dashboard_update() {
    FILE=$1
    DASHBOARD_UID=$(curl_post_file $TOKEN "$GRAFANA_HOST/api/dashboards/db" $FILE | jq -r .uid) 
    echo Created dashboard with uid $DASHBOARD_UID
    echo Dashboard URL: $GRAFANA_HOST/d/$DASHBOARD_UID/$DASHBOARD_TITLE
}


echo Creating dashboard...
DASHBOARD_DIR="data/dashboards"
DASHBOARD_PATH="$DASHBOARD_DIR/$DASHBOARD_TITLE.json"
DASHBOARD_JSON=$(jq -c \
    --arg title "$DASHBOARD_TITLE" \
    --arg sourcename "$DATASOUCE_NAME" \
    --arg sourceuid "$DATASOUCE_UID" \
    '.dashboard.title = $title | .dashboard.templating.list[0].current |= (.text = $sourcename | .value = $sourceuid)' \
    "$DASHBOARD_DIR/comet_template.json" > $DASHBOARD_PATH)
dashboard_update $DASHBOARD_PATH

# dashboard_list 