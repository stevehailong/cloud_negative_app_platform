-- 创建通知服务数据库
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
CREATE DATABASE IF NOT EXISTS notification_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE notification_db;

-- 通知表
CREATE TABLE IF NOT EXISTS notifications (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '通知ID',
    title VARCHAR(255) NOT NULL COMMENT '通知标题',
    content TEXT NOT NULL COMMENT '通知内容',
    notify_type VARCHAR(50) NOT NULL COMMENT '通知类型: release/deploy/pipeline/system',
    channel VARCHAR(50) NOT NULL COMMENT '通知渠道: email/sms/dingtalk/slack/webhook',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '发送状态: pending/sent/failed',
    receiver_type VARCHAR(20) NOT NULL COMMENT '接收者类型: user/role/group',
    receiver_ids TEXT NOT NULL COMMENT '接收者ID列表(逗号分隔)',
    sent_at DATETIME COMMENT '发送时间',
    template_id INT UNSIGNED COMMENT '关联模板ID',
    params TEXT COMMENT '模板参数(JSON格式)',
    error_msg TEXT COMMENT '错误信息',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_notify_type (notify_type),
    INDEX idx_status (status),
    INDEX idx_receiver_type (receiver_type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知记录表';

-- 通知模板表
CREATE TABLE IF NOT EXISTS notification_templates (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '模板ID',
    template_code VARCHAR(100) NOT NULL UNIQUE COMMENT '模板编码',
    template_name VARCHAR(255) NOT NULL COMMENT '模板名称',
    notify_type VARCHAR(50) NOT NULL COMMENT '通知类型',
    channel VARCHAR(50) NOT NULL COMMENT '通知渠道',
    title VARCHAR(255) COMMENT '标题模板',
    content TEXT NOT NULL COMMENT '内容模板',
    variables TEXT COMMENT '模板变量(JSON格式)',
    enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_template_code (template_code),
    INDEX idx_notify_type (notify_type),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知模板表';

-- 通知渠道配置表
CREATE TABLE IF NOT EXISTS notification_channels (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '渠道ID',
    channel_code VARCHAR(100) NOT NULL UNIQUE COMMENT '渠道编码',
    channel_name VARCHAR(255) NOT NULL COMMENT '渠道名称',
    channel_type VARCHAR(50) NOT NULL COMMENT '渠道类型: email/sms/dingtalk/slack/webhook',
    config TEXT NOT NULL COMMENT '渠道配置(JSON格式)',
    enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_channel_code (channel_code),
    INDEX idx_channel_type (channel_type),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知渠道配置表';

-- 插入默认通知模板
INSERT INTO notification_templates (template_code, template_name, notify_type, channel, title, content, variables, enabled) VALUES
('RELEASE_SUCCESS', '发布成功通知', 'release', 'dingtalk', '【发布成功】{{projectName}}', '项目: {{projectName}}\n版本: {{version}}\n环境: {{environment}}\n发布人: {{operator}}\n发布时间: {{releaseTime}}\n\n发布成功！', '["projectName","version","environment","operator","releaseTime"]', 1),
('RELEASE_FAILED', '发布失败通知', 'release', 'dingtalk', '【发布失败】{{projectName}}', '项目: {{projectName}}\n版本: {{version}}\n环境: {{environment}}\n发布人: {{operator}}\n失败原因: {{errorMsg}}\n\n请及时处理！', '["projectName","version","environment","operator","errorMsg"]', 1),
('PIPELINE_SUCCESS', '流水线成功通知', 'pipeline', 'dingtalk', '【流水线成功】{{pipelineName}}', '流水线: {{pipelineName}}\n项目: {{projectName}}\n分支: {{branch}}\n触发人: {{operator}}\n执行时间: {{duration}}\n\n执行成功！', '["pipelineName","projectName","branch","operator","duration"]', 1),
('PIPELINE_FAILED', '流水线失败通知', 'pipeline', 'dingtalk', '【流水线失败】{{pipelineName}}', '流水线: {{pipelineName}}\n项目: {{projectName}}\n分支: {{branch}}\n触发人: {{operator}}\n失败阶段: {{failedStage}}\n失败原因: {{errorMsg}}\n\n请及时处理！', '["pipelineName","projectName","branch","operator","failedStage","errorMsg"]', 1),
('DEPLOY_SUCCESS', '部署成功通知', 'deploy', 'email', '【部署成功】{{serviceName}}', '服务: {{serviceName}}\n环境: {{environment}}\n实例数: {{replicas}}\n部署人: {{operator}}\n部署时间: {{deployTime}}\n\n部署成功！', '["serviceName","environment","replicas","operator","deployTime"]', 1),
('DEPLOY_FAILED', '部署失败通知', 'deploy', 'email', '【部署失败】{{serviceName}}', '服务: {{serviceName}}\n环境: {{environment}}\n部署人: {{operator}}\n失败原因: {{errorMsg}}\n\n请及时处理！', '["serviceName","environment","operator","errorMsg"]', 1);

-- 插入默认通知渠道配置
INSERT INTO notification_channels (channel_code, channel_name, channel_type, config, enabled) VALUES
('DINGTALK_DEFAULT', '钉钉默认渠道', 'dingtalk', '{"webhook":"https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN","secret":"YOUR_SECRET"}', 1),
('EMAIL_DEFAULT', '邮件默认渠道', 'email', '{"smtp_host":"smtp.example.com","smtp_port":"465","smtp_user":"noreply@example.com","smtp_pass":"YOUR_PASSWORD","from":"noreply@example.com"}', 1),
('SLACK_DEFAULT', 'Slack默认渠道', 'slack', '{"webhook":"https://hooks.slack.com/services/YOUR/WEBHOOK/URL"}', 0),
('WEBHOOK_DEFAULT', 'Webhook默认渠道', 'webhook', '{"url":"https://your-webhook-endpoint.com/notifications","method":"POST","headers":{"Content-Type":"application/json"}}', 0);
