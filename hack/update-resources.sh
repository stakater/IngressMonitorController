#!/bin/bash

CSV_VERSION="0.0.1"

gofmt -s -w .
operator-sdk generate crds
operator-sdk generate k8s
operator-sdk generate csv --csv-version $CSV_VERSION