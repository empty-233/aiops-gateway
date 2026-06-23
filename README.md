# AIOps 智能告警网关

AIOps Gateway 是一个面向告警事件的智能分析网关。它接收监控系统产生的告警，聚合指标、日志等上下文，调用 LLM 生成结构化分析结果，并按策略触发通知和记录归档。

项目目标是把告警接入、上下文采集、智能分析、通知分发解耦成清晰的网关能力，便于在不同监控系统、指标服务、日志来源、LLM 服务和通知渠道之间扩展。

## 功能特性

- 统一 Webhook 告警入口
- 聚合指标、日志等上下文信息
- 使用 LLM 生成结构化告警分析
- 异步队列与后台 worker 处理
- 支持通知分发与告警记录归档
- 面向适配器的模块化设计

## 架构概览

```text
Alert Webhook
      |
      v
AIOps Gateway
      |
      +--> Context Collectors
      |      +--> Metrics
      |      +--> Logs
      |
      +--> LLM Analyzer
      |
      +--> Notification Channels
      |
      +--> Alert Repository
```

## 目录结构

```text
cmd/server/              服务入口
configs/                 配置样例
internal/handler/        HTTP 路由与 Webhook 处理
internal/service/        告警处理、队列和分析流程
internal/context/logs/   日志上下文采集
internal/prometheus/     指标服务适配
internal/llm/            LLM 调用抽象
internal/notify/         通知渠道抽象
internal/repository/     数据存储
```

## 环境要求

- Go 1.25+
- 可访问的 LLM 服务
- 可访问的 MySQL 实例，或可写入的 SQLite 数据目录
- 可访问的指标服务和日志来源

## 快速开始

复制配置文件：

```bash
cp configs/config.example.yaml configs/config.yaml
```

按实际环境修改 `configs/config.yaml`，然后启动服务：

```bash
go mod download
go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/health
```

## 配置说明

服务默认读取 `configs/config.yaml`。仓库提供 [configs/config.example.yaml](configs/config.example.yaml) 作为配置模板，建议复制后再修改本地配置。

配置也可以通过 `AIOPS_` 前缀环境变量覆盖。嵌套字段使用下划线连接，例如：

```bash
AIOPS_SERVER_PORT=8081
AIOPS_AI_API_KEY=your-api-key
```

主要配置分组：

| 分组 | 说明 |
| --- | --- |
| `server` | HTTP 端口和运行模式 |
| `ai` | LLM 通道、模型、提示词、超时、队列和 worker 参数 |
| `database` | 告警记录数据库配置 |
| `query` | 指标查询窗口和步长 |
| `logs` | 日志上下文来源和匹配规则 |
| `notify` | 通知渠道配置 |

常用字段：

| 字段 | 说明 |
| --- | --- |
| `server.port` | 服务监听端口 |
| `server.mode` | 运行模式，可用于控制日志格式和框架模式 |
| `ai.channel` | LLM 服务通道 |
| `ai.api_key` | LLM 服务密钥 |
| `ai.base_url` | LLM 服务地址 |
| `ai.model` | 分析使用的模型 |
| `ai.prompt` | 系统提示词 |
| `ai.timeout` | 单次分析超时时间 |
| `ai.queue_size` | 告警处理队列大小 |
| `ai.worker_count` | 后台 worker 数量 |
| `database.driver` | 数据库类型，支持 `mysql` 或 `sqlite` |
| `database.mysql` | MySQL 连接配置 |
| `database.sqlite.path` | SQLite 数据库文件路径 |
| `query.range_time` | 指标查询回看窗口 |
| `query.step` | 指标查询步长 |
| `logs.sources` | 日志来源列表 |
| `notify.channels` | 通知渠道列表 |

时间类字段支持 `30s`、`5m`、`1h` 这类 duration 写法。

## 接口说明

健康检查：

```http
GET /health
```

Webhook 告警接入：

```http
POST /webhook/prometheus
```

告警请求进入队列后会立即返回，实际分析、通知和记录写入在后台异步完成。

示例：

```bash
curl -X POST http://localhost:8080/webhook/prometheus \
  -H 'Content-Type: application/json' \
  -d @payload.json
```

响应示例：

```json
{
  "alertCount": 1,
  "status": "firing",
  "analysis": "",
  "alerts": []
}
```

## Docker 部署

推荐直接使用 Docker Hub 中的多架构镜像：

```bash
docker pull kongwu233/aiops-gateway:latest
```

启动容器：

```bash
docker run -d \
  --name aiops-gateway \
  -p 8080:8080 \
  -v "$PWD/configs/config.yaml:/app/configs/config.yaml:ro" \
  -v "$PWD/data:/app/data" \
  kongwu233/aiops-gateway:latest
```

使用 MySQL 时，请确保容器能访问配置中的数据库地址；使用 SQLite 时，保留 `data` 目录挂载用于持久化数据库文件。

也可以使用 GitHub Container Registry 镜像：

```bash
docker pull ghcr.io/kongwu233/aiops-gateway:latest
```

也可以从源码本地构建镜像：

```bash
docker build -t aiops-gateway .
```

如需采集宿主机日志，将日志目录挂载到容器内，并在配置中使用容器内路径。

## 开发

运行测试：

```bash
go test ./...
```

格式化代码：

```bash
gofmt -w .
```

## 许可证

本项目采用 GPL-3.0 许可证，详见 [LICENSE](LICENSE)。
