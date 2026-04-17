### hz 脚手架命令

1. **初始化项目**（首次使用）
   ```bash
   hz new --module video-platform --idl idl/user.proto --proto_path=. --unset_omitempty
   ```

2. **更新/新增模块**（添加新 proto 文件时）
   ```bash
   hz update --idl idl/user.proto --proto_path=. --unset_omitempty
   hz update --idl idl/video.proto --proto_path=. --unset_omitempty
   hz update --idl idl/interaction.proto --proto_path=. --unset_omitempty
   hz update --idl idl/relation.proto --proto_path=. --unset_omitempty
   ```

### 当前项目目录结构

```text
.
├── AGENTS.md
├── main.go
├── router.go
├── router_gen.go
├── build.sh
├── go.mod
├── go.sum
├── .env
├── .env.example
├── biz
│   ├── dal
│   │   ├── db
│   │   ├── model
│   │   └── rdb
│   ├── handler
│   │   ├── ping.go
│   │   └── platform
│   ├── model
│   │   ├── api
│   │   └── platform
│   ├── repository
│   │   └── user_repo.go
│   ├── router
│   │   ├── platform
│   │   └── register.go
│   └── service
│       ├── errors.go
│       └── user_service.go
├── idl
│   ├── api.proto
│   ├── common.proto
│   └── user.proto
├── pkg
│   └── response
│       └── response.go
└── script
    └── bootstrap.sh
```

### Handler 编写约定

- 对已经挂载鉴权中间件的路由，例如 `middleware.JWTAuth()`，handler 内不要重复校验 `c.Get(middleware.ContextUserID)` 是否存在
- 这类 handler 统一按当前项目既有写法处理：直接读取上下文中的 `user_id`，例如 `userIDValue, _ := c.Get(middleware.ContextUserID)`，再做类型断言
- `middleware.JWTAuth()` 已经通过 `c.Set(ContextUserID, claims.UserID)` 写入用户身份；当前项目里 `claims.UserID` 的类型是 `uint`，不是 `uint64`
- 因此已挂鉴权中间件的 handler 内不要再补 `ok` 判断、未登录分支，或“防御性”类型重复校验；默认直接使用 `userIDValue.(uint)`
- 只有未挂鉴权中间件的路由，才允许在 handler 内自行处理未登录分支

### 分层约定

- 调用链路统一按 `handler -> service -> repository -> dal/db + dal/rdb`
- `biz/dal/db` 只放 MySQL 访问代码，包括 CRUD、SQL 查询、事务，不要搬到 `repository/mysql`
- `biz/dal/rdb` 只放 Redis 访问代码，包括缓存读写、key 操作、分布式锁等
- `biz/repository` 负责组合 MySQL 和 Redis，处理缓存命中、回源、写回、失效，以及跨存储的数据组装
- `biz/service` 负责业务规则编排，原则上不要直接调用 `dal/db` 或 `dal/rdb`；需要存储访问时通过 `biz/repository`
- 后续引入缓存时优先修改 `biz/repository`，不要让 service 感知 Redis 细节

### Git 提交规范

- 使用 Conventional Commits 风格前缀，并使用中文提交信息
- 提交标题格式：`<type>: <中文说明>`
- 标题简洁明确，直接说明本次改动

常用前缀：

- `feat`: 新功能
- `fix`: 修复问题
- `refactor`: 重构，功能不变
- `docs`: 文档改动
- `test`: 测试相关
- `ci`: 持续集成、流水线、自动化构建发布
- `chore`: 杂项、工程维护

使用要求：

- 功能新增使用 `feat`
- 缺陷修复使用 `fix`
- 结构调整但不改功能使用 `refactor`
- CI/CD 或工作流改动使用 `ci`
- 不使用英文提交说明，统一使用中文
