#!/bin/bash

bash ./hack/update-resources.sh

export WATCH_NAMESPACE=""
export CONFIG_SECRET_NAME="imc-config"
export LOG_LEVEL="info"
export LOG_FORMAT="text"
export OPERATOR_NAMESPACE="test"
operator-sdk run --local --verbose
#operator-sdk run --local --verbose --enable-delve