#!/bin/sh
set -eu

if [ -z "${SERVICE_NAME:-}" ]; then
  echo "ERROR: SERVICE_NAME is required"
  exit 1
fi

if [ ! -x "/app/${SERVICE_NAME}" ]; then
  echo "ERROR: /app/${SERVICE_NAME} not found or not executable"
  exit 1
fi

if [ ! -f "${CONFIG_FILE}" ]; then
  echo "ERROR: config file not found: ${CONFIG_FILE}"
  exit 1
fi

exec "/app/${SERVICE_NAME}" "$@"
