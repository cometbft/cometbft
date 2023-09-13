GRAFANA_HOST="http://localhost:3000"

# API token (see https://grafana.com/docs/grafana/latest/administration/service-accounts/#to-add-a-token-to-a-service-account)
TOKEN=""


### Auxiliary functions

curl_get() {
    KEY=$1
    URL=$2

    curl -s -k -XGET \
        -H "Content-Type: application/json" -H "Accept: application/json" -H "Authorization: Bearer $KEY" \
        $URL
}

curl_delete() {
    KEY=$1
    URL=$2

    curl -s -k -XDELETE \
        -H "Content-Type: application/json" -H "Accept: application/json" -H "Authorization: Bearer $KEY" \
        $URL
}

curl_post() {
    KEY=$1
    URL=$2
    DATA=$3

    curl -k -s -XPOST \
        -H "Content-Type: application/json" -H "Accept: application/json" -H "Authorization: Bearer $KEY" \
        --data-binary $DATA \
        $URL
}

curl_post_file() {
    KEY=$1
    URL=$2
    FILE=$3

    curl -k -s -XPOST \
        -H "Content-Type: application/json" -H "Accept: application/json" -H "Authorization: Bearer $KEY" \
        --data-binary @$FILE \
        $URL
}
