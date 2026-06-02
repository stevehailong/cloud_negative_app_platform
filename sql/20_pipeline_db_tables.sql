-- Pipeline Service Tables
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

USE pipeline_db;

-- 流水线表
CREATE TABLE IF NOT EXISTS pipelines (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pipeline_code VARCHAR(64) NOT NULL UNIQUE,
    app_id INT UNSIGNED NOT NULL,
    pipeline_name VARCHAR(128) NOT NULL,
    pipeline_type VARCHAR(32) NOT NULL COMMENT 'ci/cd/full',
    ci_tool VARCHAR(32) NOT NULL DEFAULT 'jenkins',
    config_json JSON,
    enabled INT DEFAULT 1,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_id (app_id),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线表';

-- 流水线执行记录表
CREATE TABLE IF NOT EXISTS pipeline_runs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pipeline_id INT UNSIGNED NOT NULL,
    run_no VARCHAR(64) NOT NULL UNIQUE,
    trigger_type VARCHAR(32) NOT NULL COMMENT 'manual/webhook/mr/schedule',
    git_commit VARCHAR(64),
    git_branch VARCHAR(64),
    status VARCHAR(32) NOT NULL COMMENT 'pending/running/success/failed/cancelled',
    start_time DATETIME,
    end_time DATETIME,
    duration_seconds INT DEFAULT 0,
    operator_user_id INT UNSIGNED,
    log_url VARCHAR(255),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pipeline_id (pipeline_id),
    INDEX idx_git_branch (git_branch),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线执行记录表';

-- 制品表
CREATE TABLE IF NOT EXISTS artifacts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pipeline_run_id INT UNSIGNED NOT NULL,
    artifact_type VARCHAR(32) NOT NULL COMMENT 'image/chart/package/sbom/report',
    artifact_name VARCHAR(128) NOT NULL,
    artifact_version VARCHAR(64),
    repo_url VARCHAR(255),
    digest VARCHAR(255),
    metadata_json JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pipeline_run_id (pipeline_run_id),
    INDEX idx_artifact_type (artifact_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='制品表';
