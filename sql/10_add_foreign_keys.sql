-- 添加外键约束确保数据一致性
SET NAMES utf8mb4;
USE org_db;

-- 为projects表添加外键约束
-- 注意：如果存在孤立数据(projects中的tenant_id在tenants中不存在),需要先清理
ALTER TABLE projects 
ADD CONSTRAINT fk_projects_tenant 
FOREIGN KEY (tenant_id) REFERENCES tenants(id) 
ON DELETE RESTRICT 
ON UPDATE CASCADE;

-- 为organizations表添加外键约束
ALTER TABLE organizations 
ADD CONSTRAINT fk_organizations_tenant 
FOREIGN KEY (tenant_id) REFERENCES tenants(id) 
ON DELETE RESTRICT 
ON UPDATE CASCADE;

-- 为project_members表添加外键约束
ALTER TABLE project_members 
ADD CONSTRAINT fk_project_members_project 
FOREIGN KEY (project_id) REFERENCES projects(id) 
ON DELETE CASCADE 
ON UPDATE CASCADE;
