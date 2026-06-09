settings:
  application:
    # dev开发环境 test测试环境 prod线上环境
    mode: test
    # 服务器ip，默认使用 0.0.0.0
    host: 0.0.0.0
    # 服务名称
    name: ${SERVICE_NAME}
    # 端口号
    port: 8989
    readtimeout: 1
    writertimeout: 2
    enabledp: false

  logger:
    path: temp/logs
    stdout: ''
    level: trace
    enableddb: false

  jwt:
    # token 密钥
    secret: ${JWT_SECRET}
    timeout: 86400

  database:
    driver: mysql
    source: ${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${DB_NAME}?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms

  gen:
    dbname: ${DB_NAME}
    frontpath: ../go-admin-ui/src

  extend:
    demo:
      name: data

  cache:
    memory: ''

  queue:
    memory:
      poolSize: 100

  locker:
    redis:
