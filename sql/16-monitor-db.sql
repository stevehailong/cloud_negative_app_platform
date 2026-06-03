-- Monitor Service Database
-- 监控告警服务数据库
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建数据库
CREATE DATABASE IF NOT EXISTS monitor_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE monitor_db;

-- 指标表
CREATE TABLE IF NOT EXISTS metrics (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '指标名称',
    type VARCHAR(50) NOT NULL COMMENT '指标类型: counter/gauge/histogram/summary',
    description TEXT COMMENT '指标描述',
    unit VARCHAR(50) COMMENT '单位',
    labels JSON COMMENT '标签',
    enabled TINYINT DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_type (type),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='指标表';

-- 告警规则表
CREATE TABLE IF NOT EXISTS alert_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '规则名称',
    metric_name VARCHAR(255) NOT NULL COMMENT '指标名称',
    `condition` VARCHAR(10) NOT NULL COMMENT '告警条件: >, <, =, >=, <=, !=',
    threshold DOUBLE NOT NULL COMMENT '阈值',
    duration INT DEFAULT 60 COMMENT '持续时间(秒)',
    severity VARCHAR(20) DEFAULT 'warning' COMMENT '严重级别: critical/warning/info',
    enabled TINYINT DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
    notify_users TEXT COMMENT '通知用户列表(逗号分隔)',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_metric_name (metric_name),
    INDEX idx_severity (severity),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警规则表';

-- 告警记录表
CREATE TABLE IF NOT EXISTS alerts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id INT UNSIGNED NOT NULL COMMENT '规则ID',
    metric_name VARCHAR(255) NOT NULL COMMENT '指标名称',
    current_value DOUBLE NOT NULL COMMENT '当前值',
    threshold DOUBLE NOT NULL COMMENT '阈值',
    severity VARCHAR(20) NOT NULL COMMENT '严重级别: critical/warning/info',
    status VARCHAR(20) DEFAULT 'firing' COMMENT '状态: firing/resolved',
    message TEXT COMMENT '告警消息',
    fired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '触发时间',
    resolved_at TIMESTAMP NULL COMMENT '解决时间',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_rule_id (rule_id),
    INDEX idx_metric_name (metric_name),
    INDEX idx_severity (severity),
    INDEX idx_status (status),
    INDEX idx_fired_at (fired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';

-- 日志查询表 (Loki集成)
CREATE TABLE IF NOT EXISTS log_queries (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '查询名称',
    query TEXT NOT NULL COMMENT 'LogQL查询语句',
    description TEXT COMMENT '查询描述',
    labels JSON COMMENT '标签过滤',
    user_id INT UNSIGNED COMMENT '创建用户ID',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志查询表';

-- 链路追踪查询表 (Jaeger集成)
CREATE TABLE IF NOT EXISTS trace_queries (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '查询名称',
    service_name VARCHAR(255) COMMENT '服务名称',
    operation VARCHAR(255) COMMENT '操作名称',
    min_duration INT COMMENT '最小持续时间(微秒)',
    max_duration INT COMMENT '最大持续时间(微秒)',
    description TEXT COMMENT '查询描述',
    user_id INT UNSIGNED COMMENT '创建用户ID',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_service_name (service_name),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='链路追踪查询表';

-- 插入示例数据

-- 示例指标
INSERT INTO metrics (name, type, description, unit, labels, enabled) VALUES
('http_requests_total', 'counter', 'HTTP请求总数', 'requests', '{"service": "api-gateway"}', 1),
('cpu_usage_percent', 'gauge', 'CPU使用率', 'percent', '{"host": "server-01"}', 1),
('memory_usage_bytes', 'gauge', '内存使用量', 'bytes', '{"host": "server-01"}', 1),
('response_time_seconds', 'histogram', '响应时间', 'seconds', '{"service": "user-service"}', 1),
('error_rate', 'gauge', '错误率', 'percent', '{"service": "order-service"}', 1);

-- 示例告警规则
INSERT INTO alert_rules (name, metric_name, `condition`, threshold, duration, severity, enabled, notify_users) VALUES
('高CPU使用率告警', 'cpu_usage_percent', '>', 80, 300, 'critical', 1, 'admin,ops'),
('内存使用率告警', 'memory_usage_bytes', '>', 8589934592, 300, 'warning', 1, 'ops'),
('高错误率告警', 'error_rate', '>', 5, 60, 'critical', 1, 'admin,dev,ops'),
('慢响应告警', 'response_time_seconds', '>', 2, 120, 'warning', 1, 'dev,ops'),
('请求量异常告警', 'http_requests_total', '<', 10, 300, 'info', 1, 'ops');

-- 示例日志查询
INSERT INTO log_queries (name, query, description, labels, user_id) VALUES
('Error日志查询', '{job="api-gateway"} |= "ERROR"', '查询API网关的错误日志', '{"level": "error"}', 1),
('慢查询日志', '{job="mysql"} |= "slow query"', '查询数据库慢查询日志', '{"type": "slow"}', 1),
('用户登录日志', '{service="auth-service"} |= "login"', '查询用户登录相关日志', '{"action": "login"}', 1);

-- 链路追踪Span表（存储实际采集的trace数据）
CREATE TABLE IF NOT EXISTS trace_spans (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    trace_id VARCHAR(64) NOT NULL COMMENT '链路ID（复用X-Request-Id）',
    span_id VARCHAR(64) NOT NULL COMMENT 'Span ID',
    parent_span_id VARCHAR(64) COMMENT '父Span ID',
    service_name VARCHAR(100) NOT NULL COMMENT '服务名',
    operation_name VARCHAR(255) NOT NULL COMMENT '操作名（HTTP路径）',
    method VARCHAR(10) COMMENT 'HTTP方法',
    duration_ms INT UNSIGNED COMMENT '耗时(毫秒)',
    start_time DATETIME(3) NOT NULL COMMENT '开始时间',
    end_time DATETIME(3) COMMENT '结束时间',
    status_code INT COMMENT 'HTTP状态码',
    tags JSON COMMENT '扩展标签',
    has_error TINYINT DEFAULT 0 COMMENT '是否有错误',
    INDEX idx_trace_id (trace_id),
    INDEX idx_service_name (service_name),
    INDEX idx_start_time (start_time),
    INDEX idx_operation (operation_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='链路追踪Span表';

-- 示例链路追踪查询
INSERT INTO trace_queries (name, service_name, operation, min_duration, max_duration, description, user_id) VALUES
('慢请求追踪', 'user-service', 'GET /api/users', 1000000, NULL, '查询用户服务慢请求', 1),
('订单服务追踪', 'order-service', NULL, NULL, NULL, '查询订单服务所有请求', 1),
('支付流程追踪', 'payment-service', 'process_payment', NULL, NULL, '查询支付处理流程', 1);
