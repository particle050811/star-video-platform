### hz 脚手架命令

1. **初始化项目**（首次使用）
   ```bash
   hz new --module video-platform --idl idl/user.proto --proto_path=.
   ```

2. **更新/新增模块**（添加新 proto 文件时）
   ```bash
   hz update --idl idl/user.proto --proto_path=.
   hz update --idl idl/video.proto --proto_path=.
   hz update --idl idl/interaction.proto --proto_path=.
   hz update --idl idl/relation.proto --proto_path=.
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
