# 宠物店管理系统

全栈宠物店管理系统，支持 Web 和 App 端。

## 技术栈

- **前端**: Vue 3 + uni-app (TypeScript)
- **后端**: Go + Gin
- **数据库**: MySQL 8.0
- **ORM**: GORM

## 项目结构

```
├── server/    # Go 后端服务
├── web/       # uni-app 前端 (H5 + App)
```

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- pnpm
- Docker & Docker Compose

### 启动数据库

```bash
make docker-up
```

### 启动后端

```bash
make dev-server
```

### 启动前端

```bash
make dev-web
```

## 常用命令

```bash
make dev-server    # 启动后端开发服务
make dev-web       # 启动前端开发服务 (H5)
make build-server  # 编译后端
make build-web     # 编译前端
make docker-up     # 启动 MySQL
make docker-down   # 停止 MySQL
```
