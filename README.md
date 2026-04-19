# star-video-platform

## Docker 运行方案

当前项目采用“仅应用容器化”的方式：

- `app` 使用 Docker 镜像运行
- MySQL 和 Redis 默认复用宿主机现有实例
- 不额外在 Docker 内启动一套新的 MySQL / Redis，避免与本机已有数据分叉

## 背景说明

项目启动时会读取当前目录下的 `.env` 文件，并在启动阶段连接 MySQL、Redis。

在 Windows + Docker Desktop + WSL2 环境下，若让容器通过默认桥接网络访问宿主机数据库，虽然 `3306` / `6379` 端口可能探测为可达，但 MySQL 握手阶段可能出现以下错误：

```text
unexpected EOF
driver: bad connection
```

当前仓库已验证可用的方式是使用主机网络启动容器。这样容器内访问 `127.0.0.1` 时，可直接复用宿主机上的 MySQL 和 Redis。

## 配置文件

- `.env`：本地直接运行 Go 服务和当前 Docker 运行方式共用
- `.env.docker`：保留为备用配置文件，当前默认流程未使用

当前 Docker 方案下，镜像内复制 `.env` 到 `/app/.env`。

## 构建与启动

构建镜像：

```bash
docker build -t star-video-platform .
```

启动容器时建议后台运行，并将宿主机项目目录下的 `storage` 挂载到容器内 `/app/storage`，用于持久化上传文件：

```bash
mkdir -p storage

docker run -d --name star-video-platform --network host \
  -v "$(pwd)/storage:/app/storage" \
  star-video-platform
```

查看后台容器日志：

```bash
docker logs -f star-video-platform
```

如果需要在当前终端直接查看启动日志，也可以临时去掉 `-d` 前台运行。

启动成功后，服务默认监听：

```text
http://127.0.0.1:8888
```

## 进入运行中容器

先查看正在运行的容器：

```bash
docker ps
```

进入容器终端：

```bash
docker exec -it <容器名或容器ID> sh
```

当前运行镜像基于 Alpine，通常只有 `sh`，不一定包含 `bash`。

## 为什么使用 `--network host`

因为当前开发环境下：

- 应用跑在 Docker 容器中
- MySQL 和 Redis 跑在宿主机 / WSL2 现有环境中
- 需要直接复用已有数据库和缓存数据

如果改为在 Docker 内再启动一套 MySQL / Redis，会形成独立数据，不与当前 WSL2 中的实例互通。

使用：

```bash
docker run -d --name star-video-platform --network host \
  -v "$(pwd)/storage:/app/storage" \
  star-video-platform
```

可以避免额外维护一套数据库实例，也规避桥接网络下的 MySQL 握手异常。

## 镜像实现说明

当前 `Dockerfile` 使用多阶段构建：

1. 在 `golang:1.25` 中执行 `build.sh`
2. 在 `alpine:3.20` 中仅保留编译后的服务二进制

为了兼容 Alpine 运行镜像，`build.sh` 中已经使用：

```bash
CGO_ENABLED=0 go build -o output/bin/hertz_service
```

这样生成的二进制不依赖宿主系统动态库，更适合轻量镜像运行。

## GitHub Actions

仓库已补充 GitHub Actions：

- `ci`：执行 `go test ./...` 和 `bash ./build.sh`
- `lint`：执行 `golangci-lint`

对应文件：

- `.github/workflows/ci.yml`
- `.github/workflows/lint.yml`
- `.golangci.yml`

其中 `.golangci.yml` 已显式关闭未使用参数检查，避免因为函数里保留未使用的 `ctx` 参数导致 CI lint 失败。当前关闭项包括：

- `unparam`
- `revive` 的 `unused-parameter` 规则

如果你本地也使用 `golangci-lint`，直接在仓库根目录执行：

```bash
golangci-lint run
```
