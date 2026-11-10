# Go Service Kit

保留 Gin、API 和管理后台双入口，以及 MySQL、Redis、后台 RBAC、日志、通用中间件、多语言和 SMTP 邮件服务。初始模块路径为 `example.com/go-service-kit`，使用 Go 1.27。

## 启动 API

复制此目录到新项目，将 `go.mod` 和代码中的模块路径替换为新项目路径。在脚手架根目录运行：

```powershell
go run ./cmd/api
```

API 默认监听 `127.0.0.1:8080`，`GET /healthz` 返回存活状态，`GET /hello` 演示语言协商。MySQL、Redis 和 SMTP 按配置启用。

## 启动管理后台

在新项目目录复制 `.env.example` 为 `.env`，填写 `MYSQL_DSN`、`REDIS_KEY_PREFIX`、`JWT_SECRET` 和初始管理员凭据。在 `config/common.json` 中配置 Redis，管理端 HTTP 和日志设置放在 `config/manage.json`。MySQL 使用已存在的目标数据库，Redis 使用应用独立的键前缀。进程环境变量优先于 `.env`。

```powershell
go run ./cmd/manage --init-db
go run ./cmd/manage
```

`--init-db` 显式建表并补齐初始超管、系统开关和权限节点。重复执行保留已有管理员密码和授权。后台默认监听 `127.0.0.1:8081`，鉴权依赖 MySQL 和 Redis。后台口令使用 Argon2id，本脚手架使用新库初始化。

- `POST /auth/login` 接收 JSON `username`、`password`。后续请求通过 `Authorization: Bearer <token>` 认证。响应中的 `Authorization` 头如有值，应更新客户端令牌。
- 管理员填写邮箱后，密码验证成功会返回 `data.challenge_id`、`data.masked_email` 和 `data.expires_in`。前端将 `challenge_id` 和六位字符串 `code` 提交给 `POST /auth/login/verify-email`，验证成功后取得 `data.token` 和 `data.info`。验证码有效期 5 分钟，每账号发信间隔至少 1 分钟，每个挑战最多允许 5 次错误尝试。启用前配置 SMTP。邮箱为空的账号仍在密码验证成功后直接取得 JWT。
- `GET /common/init` 返回当前账号的菜单、权限和时区选项。
- 业务路由加在后台的鉴权路由组，并在 `cmd/manage/internal/types/management_endpoints.go` 登记权限节点。执行 `--init-db` 后，由超管给角色分配权限。普通角色必须同时满足角色启用、节点启用和节点已授权。管理员、权限和系统设置的写操作由超管执行。启用 IP 白名单前，先添加当前管理端地址并同步。

## 启动配置

从项目根目录启动时，按 **Go 默认值 → `config/common.json` → 应用 JSON（`config/api.json` 或 `config/manage.json`）→ 环境变量** 合并。缺失的文件会跳过。本地 `.env` 只补齐进程环境中未设置的变量。旧部署显式设置 `CONFIG_FILE` 时，仍只加载该变量指定的 JSON。

- 邮箱 JSON 使用 `from`、`server_host`、`server_port`、`username`、`password`。环境变量仍使用 `EMAIL_*`，显式设置后覆盖 JSON。逗号分隔的 `CORS_ORIGINS` 和 `TRUSTED_PROXIES` 显式设为空时会清空对应列表。
- 纳入版本控制的 JSON 中，DSN、密码、JWT 密钥和初始管理员凭据保留空占位，部署时由环境注入。副本 DSN 用 `MYSQL_REPLICA_0_DSN` 等变量，命名连接 DSN 用 `MYSQL_OTHER_<名称>_DSN`。
- 单点 Redis 使用 `redis.mode: "single"` 和 `redis.address`，地址可用 `REDIS_ADDRESS` 覆盖。集群 Redis 使用 `redis.mode: "cluster"` 和非空的 `redis.addrs`，同时设置单点地址会报配置冲突。默认模式为 `single`。API 的单点地址留空时跳过 Redis 连接。
- 后台运营开关由 MySQL 管理。

## 对象存储和 RSA

- 启用 `r2.enabled` 并填写 `endpoint` 和 `bucket`，通过 `R2_ACCESS_KEY_ID`、`R2_SECRET_ACCESS_KEY` 注入凭据。两端启动后可将 `runtime.Storage` 注入业务。支持对象读写、删除、分页列举、公开地址及 GET、PUT 预签名。服务端 `Put` 按声明的大小检查 `max_upload_bytes`，直传场景由业务单独控制大小。业务保存对象键，按需生成访问地址。
- 启用 `rsa.enabled` 后，从 JSON 的 `rsa.public_key_file` 加载验签或加密公钥，从 `RSA_PRIVATE_KEY_FILE` 加载签名或解密私钥。按用途提供其中一把或两把即可。两端可将 `runtime.RSA` 注入业务，使用 RSA-OAEP 加解密短数据和 RSA-PSS 签名验签，哈希均为 SHA-256，密钥至少 2048 位。密钥文件由部署环境挂载。

## Linux 打包

从脚手架根目录构建 `linux/amd64` 或 `linux/arm64` 发布包。归档只包含两个服务程序，发布清单和 SHA-256 摘要随包输出到 `dist/`。

```powershell
go run ./cmd/package -arch amd64
```

## 基础设施

升级已有后台库时，先在维护窗口执行 `go run ./cmd/manage --init-db`，补建权限码唯一索引。若提示权限码重复，先核对涉及的节点和角色授权，再处理冲突并重新执行。权限节点的 `api_path` 使用完整的 Gin 路由模板，例如 `/admin/list`。权限码由服务端根据路径和 HTTP 方法生成，已有节点的路由身份固定。

- `db/mysql` 保留 GORM 连接池、读副本和命名连接，`db/redis` 保留单点、集群和通用客户端。启用 Redis TLS 时验证证书。
- `internal/httpserver` 为两端注册请求 ID、访问日志、恢复、正文大小限制、多语言及按来源启用的 CORS。默认只采信直接连接的客户端地址，反向代理部署可用 `TRUSTED_PROXIES` 指定可信代理地址或网段，并让代理正确设置转发头。
- `utils/logger` 提供 Zap 日志、按小时轮转、日志级别和请求上下文。访问日志只记录请求元数据。后台认证后的写操作另外落库，使用 `utils/logdata` 对 JSON 脱敏。
- `internal/language` 内嵌中、英、葡、越语言支持，通过 `Accept-Language` 协商。中文使用源文案，其他语言可编辑内嵌字典。
- `shared/services/email` 提供 SMTP 客户端。配置 `EMAIL_*` 后，两端入口都能把邮件客户端注入业务。业务调用 `Send(ctx, *email.Message)`，465 端口使用隐式 TLS，其余端口使用 STARTTLS。

```powershell
go build ./...
go vet ./...
```
