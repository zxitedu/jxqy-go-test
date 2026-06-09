# ========= 1. build stage =========
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME

RUN test -n "${SERVICE_NAME}" \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
       go build -trimpath -ldflags="-s -w" \
       -o /out/${SERVICE_NAME} .

# ========= 2. config stage =========
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.22 AS config

RUN apk add --no-cache gettext

ARG SERVICE_NAME
ARG DB_NAME
ARG MYSQL_HOST
ARG MYSQL_PORT=3306
ARG MYSQL_USER
ARG MYSQL_PASSWORD
ARG JWT_SECRET

COPY settings.dev.yml.tpl /tmp/settings.dev.yml.tpl

RUN test -n "${SERVICE_NAME}" \
    && : "${MYSQL_HOST:?MYSQL_HOST build arg is required}" \
    && : "${MYSQL_PORT:?MYSQL_PORT build arg is required}" \
    && : "${MYSQL_USER:?MYSQL_USER build arg is required}" \
    && : "${MYSQL_PASSWORD:?MYSQL_PASSWORD build arg is required}" \
    && : "${JWT_SECRET:?JWT_SECRET build arg is required}" \
    && expected_db_name="${SERVICE_NAME}db" \
    && export DB_NAME="${DB_NAME:-${expected_db_name}}" \
    && if [ "${DB_NAME}" != "${expected_db_name}" ]; then \
         echo "ERROR: DB_NAME must equal SERVICE_NAME + db (${expected_db_name})"; \
         exit 1; \
       fi \
    && export SERVICE_NAME MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD JWT_SECRET \
    && mkdir -p /out \
    && envsubst < /tmp/settings.dev.yml.tpl > /out/settings.dev.yml

# ========= 3. runtime stage =========
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

WORKDIR /app

ARG SERVICE_NAME

ENV SERVICE_NAME=${SERVICE_NAME}
ENV CONFIG_FILE=/app/config/settings.dev.yml
ENV TZ=Asia/Shanghai

COPY --from=builder /out/${SERVICE_NAME} /app/${SERVICE_NAME}
COPY --from=config /out/settings.dev.yml /app/config/settings.dev.yml
COPY entrypoint.sh /app/entrypoint.sh

RUN test -n "${SERVICE_NAME}" \
    && chmod +x /app/entrypoint.sh \
    && chmod +x /app/${SERVICE_NAME}

EXPOSE 8989

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["server", "-c", "/app/config/settings.dev.yml", "-a", "true"]
