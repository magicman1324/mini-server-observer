# Hydra — 分布式服务器状态监控与秒级告警系统

基于 Pull 模型的分布式服务器监控平台，支持多区域联邦、Prometheus 兼容指标采集、ClickHouse 时序存储与秒级告警。

## 架构

```
Agent (:9100 /metrics) ←──HTTP scrape──→ Regional Collector ──gRPC stream──→ Central Aggregator
                                                                                  │
                                                                     ┌────────────┼────────────┐
                                                                     ▼            ▼            ▼
                                                                 ClickHouse   Alert Engine  REST API (:8080)
                                                                 (metrics,     (state        (Gin, JWT)
                                                                  rules,...)   machine)          │
                                                                                     ┌─────────▼─────────┐
                                                                                     │  Dashboard (React) │
                                                                                     │  :80 (Nginx)       │
                                                                                     └───────────────────┘
```

## 与哨兵（Project1）的核心差异

| 维度 | 哨兵 | Hydra |
|------|------|-------|
| 采集模型 | Push (Agent → Kafka) | **Pull** (Regional Collector 抓取) |
| 传输层 | Kafka broker | **gRPC 双向流** |
| 流计算 | Java Flink | **纯 Go pipeline** |
| 告警引擎 | Prometheus Alertmanager | **自研 state machine** |
| 时序存储 | TiDB | **ClickHouse** |
| 指标格式 | 自定义 JSON | **Prometheus exposition format** |
| 架构 | 单区域 | **多区域联邦** |

## 技术栈

| 组件 | 选型 |
|------|------|
| Agent | Go 1.25，纯 procfs/cgroup，<5MB 二进制 |
| Regional Collector | Go 1.25 + gRPC stream client |
| Central Aggregator | Go 1.25 + Gin (REST) + gRPC server |
| Alert Engine | Go state machine + channel pipeline |
| 存储 | ClickHouse MergeTree (列式，分区 + TTL) |
| 前端 | React 19 + TypeScript + Vite 6 + Recharts + Tailwind CSS 3.4 |
| 部署 | Docker Compose + Kubernetes |

## 项目结构

```
├── agent/                 # 轻量级采集 Agent (Prometheus /metrics)
├── backend/               # Go 后端 (Regional + Central)
│   ├── cmd/regional/      # Regional Collector 入口
│   ├── cmd/central/       # Central Aggregator 入口
│   └── internal/
│       ├── proto/         # gRPC 服务定义
│       ├── scraper/       # HTTP 指标抓取器
│       ├── ingest/        # gRPC 流摄取
│       ├── alert/         # 自研告警引擎
│       ├── repository/    # ClickHouse 数据访问
│       └── api/rest/      # REST API
├── frontend/              # React Dashboard
├── schema/                # ClickHouse DDL
├── deploy/                # Docker Compose + K8s
└── configs/               # 环境配置
```

## 快速开始

```bash
# 本地开发 (全栈)
cd deploy && docker compose up -d
# 浏览器访问 http://localhost

# 仅运行 Agent
cd agent && go run ./cmd/agent/
# curl http://localhost:9100/metrics

# 运行测试
make test
```

## License

MIT
