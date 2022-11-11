#!/bin/sh

if [ $# -lt 1 ]; then
    echo "Missing port"
    exit 1
fi

exec ${OPENVSCODE_SERVER_ROOT}/bin/openvscode-server \
    --host 0.0.0.0 \
    --port $1 \
    --connection-token $2