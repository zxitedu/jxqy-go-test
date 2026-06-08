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

ENV SERVICE_NAME=${SERVICE_NAME}
ENV CONFIG_TEMPLATE=/app/config/settings.dev.yml.tpl
ENV CONFIG_FILE=/app/config/settings.dev.yml
ENV TZ=Asia/Shanghai

COPY --from=builder /out/${SERVICE_NAME} /app/${SERVICE_NAME}
COPY settings.dev.yml.tpl /app/config/settings.dev.yml.tpl
COPY entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh \
    && chmod +x /app/${SERVICE_NAME}

EXPOSE 8989

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["server", "-c", "/app/config/settings.dev.yml", "-a", "true"]
