# 学术图表视觉编码歧义复核台 (task251-chartambiguity)

基于 Go 实现的学术图表视觉编码歧义复核 Web 项目。面向科研图形编辑，校验一张图中
颜色、形状或坐标轴是否使两个变量被错误地解释为同一含义，并支持把可发布图规范
冻结为不可变版本。

## 业务闭环

导入图层/轴/图例/变量语义 → 声明视觉编码（变量经通道以 token 呈现）→ 复核歧义
（颜色复用、形状复用、轴单位冲突、缺图例、映射矛盾）→ 登记豁免或修订 → 发布图规范版本。

## 核心不变量

- 同一视觉通道的同一 token 至多代表一个变量（否则颜色复用歧义）。
- 同一变量在坐标轴上的单位必须一致（否则轴单位冲突）。
- 每个被使用的视觉通道 token 必须有图例覆盖（否则缺图例歧义）。
- 图规范冻结后绑定编码与例外快照，不可变；新版本发布会使旧 frozen 版本置为 superseded。

## 标准命令

```bash
# 构建、静态检查、测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...

# 启动服务（SQLite 持久化，重启可恢复）
go run ./cmd/chartambiguity --addr :8080 --db chartambiguity.db

# 自检（验证持久化与重启恢复，退出码 0 表示通过）
go run ./cmd/chartambiguity --smoke-test
```

## API 入口（前缀 /api）

| 能力 | 入口 |
|---|---|
| 图形稿 | `POST /api/figures` · `GET /api/figures` · `GET /api/figures/{id}` · `GET /api/figures/{id}/status` · `GET /api/figures/{id}/summary` |
| 语义导入 | `POST /api/figures/{id}/import` · `GET /api/figures/{id}/layers` · `GET /api/figures/{id}/axes` · `GET /api/figures/{id}/legends` · `GET /api/figures/{id}/variables` |
| 编码 | `POST /api/figures/{id}/encodings` · `GET /api/figures/{id}/encodings` |
| 复核 | `POST /api/figures/{id}/check` · `GET /api/figures/{id}/ambiguities` · `GET /api/figures/{id}/mappings` |
| 豁免 | `POST /api/figures/{id}/exceptions` · `GET /api/figures/{id}/exceptions` |
| 规范 | `POST /api/figures/{id}/specs` · `POST /api/figures/{id}/specs/{sid}/publish` · `GET /api/figures/{id}/specs` · `GET /api/specs/{sid}` |
| 自检 | `GET /api/selfcheck` |

## 持久化

使用纯 Go 驱动 `modernc.org/sqlite`（CGO 无关），所有实体落盘 SQLite；`--smoke-test`
会在关闭并重开同一数据库后验证图形稿、编码与规范版本均被正确恢复。
