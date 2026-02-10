# patpet-server

简单登录后端：Go + Gin + GORM + PostgreSQL + JWT。

## 本地运行

```bash
# 1. 启动数据库
docker-compose up -d

# 2. 安装依赖
go mod tidy

# 3. 启动服务
go run .
```

## API

- `POST /api/v1/register` — 注册（body: email, password, nickname）
- `POST /api/v1/login` — 登录（body: email, password）
- `GET /api/v1/profile` — 获取当前用户（Header: Authorization: Bearer <token>）

## 环境变量

- `DATABASE_URL` — PostgreSQL 连接串（默认本地）
- `JWT_SECRET` — JWT 密钥
- `PORT` — 服务端口（默认 8080）
