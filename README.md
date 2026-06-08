# jxqy-go-test

一个最小可运行的 Go HTTP JSON API 示例项目，启动时会强制连接 MySQL。

## 规则

二进制文件名就是服务名，也是必须连接的 MySQL 库名。

例如：

```bash
go build -o api .
./api server -c settings.dev.yml
```

上面的 `settings.dev.yml` 里必须配置 `mysql.db: api`，否则进程会直接退出。`admin` 同理必须连接 `admin` 库。

## 配置

项目根目录提供了模板文件：

```bash
settings.dev.yml.tpl
```

渲染后的配置格式：

```yaml
mysql:
  host: "127.0.0.1"
  port: 3306
  db: "api"
  user: "root"
  password: "secret"
```

## 本地运行

先准备 MySQL，并创建对应库：

```sql
CREATE DATABASE api DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE admin DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

构建并启动 `api`：

```bash
go build -o api .
./api server -c settings.dev.yml
```

默认监听 `8989` 端口，也可以通过环境变量指定端口：

```bash
PORT=9000 ./api server -c settings.dev.yml
```

## Docker

构建 `api` 镜像：

```bash
docker build \
  --build-arg SERVICE_NAME=api \
  --build-arg CMD_PATH=. \
  -t jxqy-api .
```

运行：

```bash
docker run --rm -p 8989:8989 \
  -e MYSQL_HOST=host.docker.internal \
  -e MYSQL_PORT=3306 \
  -e MYSQL_DB=api \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=secret \
  jxqy-api
```

构建 `admin` 时把 `SERVICE_NAME` 和 `MYSQL_DB` 都改成 `admin`。

## 接口

```bash
curl http://localhost:8989/
curl http://localhost:8989/healthz
curl http://localhost:8989/api/v1/mysql
curl "http://localhost:8989/api/v1/hello?name=Tom"
```

示例响应：

```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "name": "Tom",
    "message": "Hello, Tom",
    "service": "api",
    "mysql_db": "api"
  },
  "timestamp": 1780911600
}
```

## 测试

```bash
go test ./...
```
