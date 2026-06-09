# jxqy-go-test

一个最小可运行的 Go HTTP JSON API 示例项目，启动时会强制连接 MySQL。

## 规则

二进制文件名就是服务名，必须连接的 MySQL 库名是服务名后面加 `db`。

例如：

```bash
go build -o api .
./api server -c settings.dev.yml
```

上面的 `settings.dev.yml` 里 `settings.application.name` 必须是 `api`，`settings.database.source` 和 `settings.gen.dbname` 必须使用 `apidb`，否则进程会直接退出。`admin` 同理必须连接 `admindb` 库，`house_admin` 必须连接 `house_admindb` 库。

## 配置

项目根目录提供了模板文件：

```bash
settings.dev.yml.tpl
```

渲染后的配置格式：

```yaml
settings:
  application:
    name: api
    port: 8989
  database:
    driver: mysql
    source: root:secret@tcp(mysql:3306)/apidb?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms
  gen:
    dbname: apidb
```

## 本地运行

先准备 MySQL，并创建对应库：

```sql
CREATE DATABASE apidb DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE admindb DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

构建并启动 `api`：

```bash
go build -o api .
./api server -c settings.dev.yml
```

默认监听 `8989` 端口。

## Docker

构建 `api` 镜像：

```bash
docker build \
  --build-arg SERVICE_NAME=api \
  --build-arg MYSQL_HOST=mysql \
  --build-arg MYSQL_USER=root \
  --build-arg MYSQL_PASSWORD=secret \
  --build-arg JWT_SECRET=change-me \
  -t jxqy-api .
```

构建时会根据 `settings.dev.yml.tpl` 生成 `/app/config/settings.dev.yml` 并打包进镜像。`DB_NAME` 默认等于 `SERVICE_NAME + db`，所以 `SERVICE_NAME=api` 会生成 `DB_NAME=apidb`。
`MYSQL_PORT` 可以不传，默认是 `3306`。

镜像固定 `EXPOSE 8989`，配置里的 `settings.application.port` 也是 `8989`。
最终镜像只包含生成后的 `/app/config/settings.dev.yml`，不再依赖运行时环境变量重新生成配置。
`entrypoint.sh` 只负责检查服务二进制和配置文件是否存在，然后启动服务。

运行：

```bash
docker run --rm -p 8989:8989 jxqy-api
```

构建 `admin` 时把 `SERVICE_NAME` 改成 `admin`，镜像里的 `DB_NAME` 会自动生成成 `admindb`。如果显式传 `DB_NAME`，它也必须等于 `SERVICE_NAME + db`。

注意：把 `settings.dev.yml` 打进镜像会把数据库密码也固化进去，这种方式更适合开发或测试镜像。

## 接口

```bash
curl http://localhost:8989/
curl http://localhost:8989/healthz
curl http://localhost:8989/api/v1/mysql
curl http://localhost:8989/api/v1/users
curl "http://localhost:8989/api/v1/hello?name=Tom"
```

`GET /api/v1/users` 会查询当前服务所连接 MySQL 库里的 `user` 表：

```sql
SELECT username FROM `user` ORDER BY username;
```

用户查询示例响应：

```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "service": "api",
    "mysql_db": "apidb",
    "usernames": ["alice", "bob"]
  },
  "timestamp": 1780911600
}
```

Hello 接口示例响应：

```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "name": "Tom",
    "message": "Hello, Tom",
    "service": "api",
    "mysql_db": "apidb"
  },
  "timestamp": 1780911600
}
```

## 测试

```bash
go test ./...
```
