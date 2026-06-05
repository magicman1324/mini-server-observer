-- ============================================================
-- Hydra 分布式监控平台 — ClickHouse Schema
-- Engine 选型针对时序监控场景优化
-- ============================================================

CREATE DATABASE IF NOT EXISTS hydra;

-- ============================================================
-- 核心时序数据表 (高写入吞吐)
-- MergeTree + 按日分区 + 90天 TTL
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.metrics (
    ts          DateTime64(3)           COMMENT '指标时间戳 (毫秒精度)',
    name        LowCardinality(String)  COMMENT '指标名: cpu_usage_percent, memory_usage_bytes...',
    value       Float64                 COMMENT '指标值',
    host        LowCardinality(String)  COMMENT '来源主机名',
    region      LowCardinality(String)  COMMENT '来源区域 (ap-shanghai-1 等)',
    labels      Map(String, String)     COMMENT '额外标签键值对'
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(ts)
ORDER BY (name, host, ts)
TTL ts + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

-- ============================================================
-- 小时级降采样聚合表 (长期保留)
-- SummingMergeTree 自动合并同主键行
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.metrics_hourly (
    ts_hour     DateTime                COMMENT '小时桶起始',
    name        LowCardinality(String),
    host        LowCardinality(String),
    region      LowCardinality(String),
    avg_val     Float64                 COMMENT '小时均值',
    max_val     Float64                 COMMENT '小时峰值',
    min_val     Float64                 COMMENT '小时谷值',
    sample_count UInt64                 COMMENT '原始采样数'
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(ts_hour)
ORDER BY (name, host, ts_hour)
TTL ts_hour + INTERVAL 365 DAY DELETE;

-- 物化视图: 自动从 metrics 聚合到 metrics_hourly
CREATE MATERIALIZED VIEW IF NOT EXISTS hydra.metrics_hourly_mv
TO hydra.metrics_hourly
AS SELECT
    toStartOfHour(ts) AS ts_hour,
    name,
    host,
    region,
    avg(value)        AS avg_val,
    max(value)        AS max_val,
    min(value)        AS min_val,
    count()           AS sample_count
FROM hydra.metrics
GROUP BY ts_hour, name, host, region;

-- ============================================================
-- 告警规则 (运维配置, 小数据量, 高频读取)
-- ReplacingMergeTree 支持 upsert
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.rules (
    id           String                  COMMENT 'UUID',
    name         String                  COMMENT '规则名称',
    description  String         DEFAULT '',
    metric       String                  COMMENT '监控指标名',
    operator     LowCardinality(String)  COMMENT '>, >=, <, <=, ==, !=',
    threshold    Float64                 COMMENT '阈值',
    duration_sec UInt32         DEFAULT 30 COMMENT '持续多久触发 (秒)',
    severity     LowCardinality(String)  COMMENT 'critical / warning / info',
    expression   String         DEFAULT '' COMMENT '复合规则表达式 (可选)',
    enabled      UInt8          DEFAULT 1,
    labels       Map(String, String)     COMMENT '附加标签',
    annotations  Map(String, String)     COMMENT 'summary, description 模板',
    created_at   DateTime       DEFAULT now(),
    updated_at   DateTime       DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY id;

-- ============================================================
-- 告警事件 (触发记录 + 恢复记录)
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.alerts (
    id           String                  COMMENT 'UUID',
    rule_id      String                  COMMENT '关联规则 ID',
    rule_name    String                  COMMENT '冗余规则名 (加速查询)',
    host         LowCardinality(String),
    region       LowCardinality(String),
    severity     LowCardinality(String),
    metric       String,
    value        Float64                 COMMENT '触发时的指标值',
    threshold    Float64                 COMMENT '触发时的阈值',
    message      String                  COMMENT '可读告警消息',
    status       LowCardinality(String)  COMMENT 'firing / resolved',
    fired_at     DateTime64(3),
    resolved_at  Nullable(DateTime64(3)),
    labels       Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(fired_at)
ORDER BY (rule_id, host, fired_at)
TTL fired_at + INTERVAL 180 DAY DELETE;

-- ============================================================
-- 通知渠道
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.alert_channels (
    id           String,
    name         String,
    type         LowCardinality(String)  COMMENT 'dingtalk / email',
    config       String                  COMMENT 'JSON 配置 (webhook URL, SMTP 等)',
    enabled      UInt8          DEFAULT 1,
    created_at   DateTime       DEFAULT now()
)
ENGINE = ReplacingMergeTree()
ORDER BY id;

-- ============================================================
-- 规则-渠道映射 (M:N)
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.rule_channels (
    rule_id      String,
    channel_id   String
)
ENGINE = MergeTree()
ORDER BY (rule_id, channel_id);

-- ============================================================
-- 告警静默 (维护窗口)
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.silences (
    id           String,
    matchers     String                  COMMENT 'JSON 匹配器数组',
    starts_at    DateTime,
    ends_at      DateTime,
    comment      String         DEFAULT '',
    created_by   String         DEFAULT '',
    created_at   DateTime       DEFAULT now()
)
ENGINE = MergeTree()
ORDER BY (starts_at, ends_at);

-- ============================================================
-- 主机注册表 (首次收到指标时自动注册)
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.hosts (
    host         String,
    region       LowCardinality(String),
    labels       Map(String, String),
    first_seen   DateTime       DEFAULT now(),
    last_seen    DateTime       DEFAULT now()
)
ENGINE = ReplacingMergeTree(last_seen)
ORDER BY host;

-- ============================================================
-- 审计日志
-- ============================================================
CREATE TABLE IF NOT EXISTS hydra.audit_logs (
    id           String,
    action       LowCardinality(String)  COMMENT 'create / update / delete',
    resource     LowCardinality(String)  COMMENT 'rules / silences / channels',
    resource_id  String,
    detail       String                  COMMENT '变更详情 JSON',
    performed_by String         DEFAULT '',
    created_at   DateTime       DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (resource, created_at)
TTL created_at + INTERVAL 365 DAY DELETE;
