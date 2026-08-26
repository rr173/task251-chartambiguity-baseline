基于 Go 实现的学术图表视觉编码歧义复核 Web 项目，一款后端服务，在统一变量-通道映射上检测颜色复用、轴单位冲突与缺图例标记，并将可发布图规范冻结为不可变版本。

# BENZHI 评测说明

## 1. 项目类型
学术图表视觉编码歧义复核台：导入图层/轴/图例/变量语义，建立变量到视觉通道（颜色/形状/尺寸/线型）的映射，检测同色复用、轴单位冲突与缺图例等歧义，登记豁免后把图规范冻结为不可变版本。

## 2. 标准命令
```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/chartambiguity --smoke-test
go run ./cmd/chartambiguity --addr :8080 --db chartambiguity.db
```

## 3. 评测镜像
`Dockerfile` 与 `benzhi.Dockerfile` 内容完全一致；使用 Go 1.26.3 Bookworm builder 和 Alpine 3.20 runtime 的多阶段构建，产物为 `/app/chartambiguity`。脚本第二个参数为目标平台。镜像不声明固定端口，服务监听地址由 `--addr` 指定。

```bash
./build_benzhi_docker.sh chartambiguity linux/amd64
./build_benzhi_docker.sh chartambiguity-arm64 linux/arm64
docker run --rm chartambiguity --smoke-test
docker run --rm -P chartambiguity --addr :8080 --db ./app.db
```

## 4. 关键 API
- 创建图形：`POST /api/figures` `{"title":"..."}`
- 导入语义：`POST /api/figures/{id}/import` `{"layers":[...],"axes":[...],"variables":[...],"legends":[...]}`
- 声明编码：`POST /api/figures/{id}/encodings` `{"variable":"temp","channel":"color","token":"#1f77b4"}`
- 复核歧义：`POST /api/figures/{id}/check`
- 查看歧义：`GET /api/figures/{id}/ambiguities`
- 登记豁免：`POST /api/figures/{id}/exceptions` `{"kind":"reuse_exemption","target_channel":"color","target_token":"#1f77b4","reason":"..."}`
- 创建规范：`POST /api/figures/{id}/specs`
- 冻结发布：`POST /api/figures/{id}/specs/{sid}/publish`
- 自检：`GET /api/selfcheck`

## 5. --smoke-test 契约
不启动长驻服务；真实创建图形→导入语义→声明两个变量共用同色编码→检测 1 条颜色复用歧义→登记豁免使歧义消解→创建并冻结图规范→关闭并重开同一数据库验证图形稿/编码/规范版本持久化恢复，最后以 0 退出码结束。退出码非 0 即视为失败。

## 6. 环境与组件
- Go 1.26.3（GOTOOLCHAIN=local，CGO_ENABLED=0）
- SQLite 3.46.1（modernc.org/sqlite v1.52.0，CGO 无关）
