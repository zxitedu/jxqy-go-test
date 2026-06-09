# ========= 1. build stage =========
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME
ARG CMD_PATH

RUN test -n "${SERVICE_NAME}" \
    && test -n "${CMD_PATH}" \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
       go build -trimpath -ldflags="-s -w" \
       -o /out/${SERVICE_NAME} ${CMD_PATH}

# ========= 2. runtime stage =========
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.22

RUN apk add --no-cache ca-certificates tzdata gettext \
    && update-ca-certificates

WORKDIR /app

ARG SERVICE_NAME
ARG MYSQL_HOST
ARG MYSQL_PORT
ARG MYSQL_DB
ARG MYSQL_USER
ARG MYSQL_PASSWORD
ARG SERVICE_PORT=8989

ENV SERVICE_NAME=${SERVICE_NAME}
ENV CONFIG_TEMPLATE=/app/config/settings.dev.yml.tpl
ENV CONFIG_FILE=/app/config/settings.dev.yml
ENV PORT=${SERVICE_PORT}
ENV TZ=Asia/Shanghai

COPY --from=builder /out/${SERVICE_NAME} /app/${SERVICE_NAME}
COPY settings.dev.yml.tpl /app/config/settings.dev.yml.tpl
COPY entrypoint.sh /app/entrypoint.sh

RUN test -n "${SERVICE_NAME}" \
    && : "${SERVICE_PORT:?SERVICE_PORT build arg is required}" \
    && : "${MYSQL_HOST:?MYSQL_HOST build arg is required}" \
    && : "${MYSQL_PORT:?MYSQL_PORT build arg is required}" \
    && : "${MYSQL_USER:?MYSQL_USER build arg is required}" \
    && : "${MYSQL_PASSWORD:?MYSQL_PASSWORD build arg is required}" \
    && expected_mysql_db="${SERVICE_NAME}db" \
    && export MYSQL_DB="${MYSQL_DB:-${expected_mysql_db}}" \
    && if [ "${MYSQL_DB}" != "${expected_mysql_db}" ]; then \
         echo "ERROR: MYSQL_DB must equal SERVICE_NAME + db (${expected_mysql_db})"; \
         exit 1; \
       fi \
    && export MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD \
    && envsubst < "${CONFIG_TEMPLATE}" > "${CONFIG_FILE}" \
    && chmod +x /app/entrypoint.sh \
    && chmod +x /app/${SERVICE_NAME}

EXPOSE ${PORT}

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["server", "-c", "/app/config/settings.dev.yml", "-a", "true"]
