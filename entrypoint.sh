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

should_render_config=false
if [ -n "${MYSQL_HOST:-}" ] \
  || [ -n "${MYSQL_PORT:-}" ] \
  || [ -n "${MYSQL_DB:-}" ] \
  || [ -n "${MYSQL_USER:-}" ] \
  || [ -n "${MYSQL_PASSWORD:-}" ]; then
  should_render_config=true
fi

if [ "${should_render_config}" = "true" ]; then
  if [ ! -f "${CONFIG_TEMPLATE}" ]; then
    echo "ERROR: config template not found: ${CONFIG_TEMPLATE}"
    exit 1
  fi

  : "${MYSQL_HOST:?MYSQL_HOST is required}"
  : "${MYSQL_PORT:?MYSQL_PORT is required}"
  : "${MYSQL_DB:?MYSQL_DB is required}"
  : "${MYSQL_USER:?MYSQL_USER is required}"
  : "${MYSQL_PASSWORD:?MYSQL_PASSWORD is required}"

  if [ "${MYSQL_DB}" != "${SERVICE_NAME}" ]; then
    echo "ERROR: MYSQL_DB must equal SERVICE_NAME (${SERVICE_NAME})"
    exit 1
  fi

  mkdir -p "$(dirname "${CONFIG_FILE}")"
  envsubst < "${CONFIG_TEMPLATE}" > "${CONFIG_FILE}"
elif [ ! -f "${CONFIG_FILE}" ]; then
  echo "ERROR: config file not found: ${CONFIG_FILE}"
  exit 1
fi

exec "/app/${SERVICE_NAME}" "$@"
