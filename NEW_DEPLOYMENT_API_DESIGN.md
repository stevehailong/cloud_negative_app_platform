# 新的部署管理API设计

## 架构概述

### 核心理念
- **以应用为维度**：每个应用在每个环境只有一条主记录
- **主记录 + 历史记录**：主记录存储当前状态，历史记录存储所有变更
- **操作更直观**：重启、扩缩容、回滚都针对主记录操作

### 数据模型

```
app_deployments (主记录表)
├── id: 1
├── app_id: 8
├── env_id: 1
├── workload_name: app-8-canary
├── current_version: v2.1.0
├── current_image: nginx:1.25-alpine
├── desired_replicas: 5
└── deployment_status: running

deployment_history (历史记录表)
├── id: 101, app_deployment_id: 1, type: create,  version: v1.0, replicas: 5
├── id: 102, app_deployment_id: 1, type: update,  version: v1.1, replicas: 5
├── id: 103, app_deployment_id: 1, type: scale,   version: v1.1, replicas: 8
├── id: 104, app_deployment_id: 1, type: update,  version: v2.0, replicas: 8
└── id: 105, app_deployment_id: 1, type: scale,   version: v2.0, replicas: 5
```

## API端点设计

### 1. 部署列表（主记录）

```http
GET /api/v1/app-deployments?page=1&pageSize=20&appId=8&envId=1

Response:
{
  "code": 0,
  "data": {
    "total": 3,
    "list": [
      {
        "id": 1,
        "appId": 8,
        "appName": "book-service",
        "envId": 1,
        "envName": "生产环境",
        "namespace": "app-8",
        "workloadName": "app-8-canary",
        "currentVersion": "v2.1.0",
        "currentImage": "nginx:1.25-alpine",
        "desiredReplicas": 5,
        "availableReplicas": 5,
        "deploymentStatus": "running",
        "lastDeployTime": "2026-05-30 10:00:00",
        "lastDeployUser": "admin"
      }
    ]
  }
}
```

### 2. 部署详情

```http
GET /api/v1/app-deployments/1

Response:
{
  "code": 0,
  "data": {
    "id": 1,
    "appId": 8,
    "envId": 1,
    "namespace": "app-8",
    "workloadName": "app-8-canary",
    
    // 当前状态
    "currentVersion": "v2.1.0",
    "currentImage": "nginx:1.25-alpine",
    "desiredReplicas": 5,
    "availableReplicas": 5,
    "deploymentStatus": "running",
    
    // K8s实时状态
    "k8sStatus": {
      "replicas": 5,
      "readyReplicas": 5,
      "updatedReplicas": 5,
      "conditions": [...]
    },
    
    // 最后部署信息
    "lastDeployId": 105,
    "lastDeployTime": "2026-05-30 10:00:00",
    "lastDeployUser": "admin",
    
    // 统计信息
    "stats": {
      "totalDeploys": 5,
      "successRate": 100,
      "avgDuration": 120
    }
  }
}
```

### 3. 部署历史记录

```http
GET /api/v1/app-deployments/1/history?page=1&pageSize=20

Response:
{
  "code": 0,
  "data": {
    "total": 5,
    "list": [
      {
        "id": 105,
        "deploymentType": "scale",
        "version": "v2.0",
        "imageUrl": "nginx:1.25-alpine",
        "replicas": 5,
        "changes": {
          "replicas": "8 → 5"
        },
        "operatorUser": "admin",
        "startTime": "2026-05-30 10:00:00",
        "endTime": "2026-05-30 10:00:10",
        "duration": 10,
        "status": "success"
      },
      {
        "id": 104,
        "deploymentType": "update",
        "version": "v2.0",
        "imageUrl": "nginx:1.25-alpine",
        "replicas": 8,
        "changes": {
          "image": "httpd:alpine → nginx:1.25-alpine"
        },
        "operatorUser": "admin",
        "startTime": "2026-05-30 09:00:00",
        "endTime": "2026-05-30 09:00:30",
        "duration": 30,
        "status": "success"
      }
    ]
  }
}
```

### 4. 重启部署

```http
POST /api/v1/app-deployments/1/restart

Request:
{
  "reason": "修复内存泄漏"
}

Response:
{
  "code": 0,
  "data": {
    "message": "重启成功",
    "historyId": 106
  }
}

Effect:
- 创建历史记录: type=restart
- 执行 kubectl rollout restart
- 更新主记录的 last_deploy_time
```

### 5. 扩缩容

```http
POST /api/v1/app-deployments/1/scale

Request:
{
  "replicas": 10,
  "reason": "应对流量高峰"
}

Response:
{
  "code": 0,
  "data": {
    "message": "扩缩容成功",
    "historyId": 107
  }
}

Effect:
- 创建历史记录: type=scale, replicas=10, changes={"replicas": "5→10"}
- 执行 kubectl scale
- 更新主记录的 desired_replicas
```

### 6. 回滚

```http
POST /api/v1/app-deployments/1/rollback

Request:
{
  "targetHistoryId": 104  // 回滚到历史记录ID=104的版本
}

Response:
{
  "code": 0,
  "data": {
    "message": "回滚成功",
    "historyId": 108,
    "rolledBackTo": {
      "version": "v2.0",
      "image": "nginx:1.25-alpine"
    }
  }
}

Effect:
- 创建历史记录: type=rollback
- 执行部署：使用目标历史记录的镜像
- 更新主记录的 current_version, current_image
```

### 7. 部署新版本

```http
POST /api/v1/app-deployments/1/deploy

Request:
{
  "releaseId": 12,
  "version": "v2.2.0",
  "imageUrl": "nginx:1.26-alpine",
  "replicas": 5
}

Response:
{
  "code": 0,
  "data": {
    "message": "部署成功",
    "historyId": 109
  }
}

Effect:
- 创建历史记录: type=update
- 执行 K8s 部署
- 更新主记录的 current_version, current_image
```

## 前端页面设计

### 1. 部署列表页

```
╔════════════════════════════════════════════════════════════╗
║ 部署管理                                   [+ 新建部署]    ║
╠════════════════════════════════════════════════════════════╣
║                                                             ║
║ 应用: [全部 ▼]  环境: [全部 ▼]  状态: [全部 ▼]  [搜索]   ║
║                                                             ║
║ ┌─────────────────────────────────────────────────────┐   ║
║ │ 应用名称      │ 环境 │ 版本   │ 副本  │ 状态 │ 操作 │   ║
║ ├─────────────────────────────────────────────────────┤   ║
║ │ book-service │ 生产 │ v2.1.0 │ 5/5   │ 🟢运行│[详情]│   ║
║ │   app-8      │      │        │       │      │[重启]│   ║
║ │              │      │        │       │      │[扩缩]│   ║
║ ├─────────────────────────────────────────────────────┤   ║
║ │ user-service │ 生产 │ v1.5.3 │ 3/3   │ 🟢运行│[详情]│   ║
║ │   app-5      │      │        │       │      │[重启]│   ║
║ │              │      │        │       │      │[扩缩]│   ║
║ └─────────────────────────────────────────────────────┘   ║
╚════════════════════════════════════════════════════════════╝
```

### 2. 部署详情页

```
╔════════════════════════════════════════════════════════════╗
║ book-service (app-8)                  [返回列表]           ║
╠════════════════════════════════════════════════════════════╣
║                                                             ║
║ ┌─ 基本信息 ─────────────────────────────────────────┐   ║
║ │ 命名空间: app-8                环境: 生产环境        │   ║
║ │ 当前版本: v2.1.0              镜像: nginx:1.25      │   ║
║ │ 副本数: 5/5                   状态: 🟢 运行中        │   ║
║ │ 最后部署: 2026-05-30 10:00    操作人: admin        │   ║
║ └────────────────────────────────────────────────────┘   ║
║                                                             ║
║ ┌─ 快速操作 ─────────────────────────────────────────┐   ║
║ │ [🔄 重启] [📈 扩缩容] [⏮ 回滚] [🚀 部署新版本]      │   ║
║ └────────────────────────────────────────────────────┘   ║
║                                                             ║
║ ┌─ Pod列表 ──────────────────────────────────────────┐   ║
║ │ app-8-canary-xxx-1  Running  10.244.1.10  node-1   │   ║
║ │ app-8-canary-xxx-2  Running  10.244.2.11  node-2   │   ║
║ │ app-8-canary-xxx-3  Running  10.244.3.12  node-3   │   ║
║ └────────────────────────────────────────────────────┘   ║
║                                                             ║
║ ┌─ 部署历史 ─────────────────────────────────────────┐   ║
║ │ 时间              │ 类型   │ 版本   │ 变更      │操作│   ║
║ ├─────────────────────────────────────────────────────┤   ║
║ │ 05-30 10:00:00  │ 扩缩容 │ v2.1.0 │ 8→5副本   │[-] │   ║
║ │ 05-30 09:00:00  │ 更新   │ v2.1.0 │ 新镜像    │[回]│   ║
║ │ 05-29 15:00:00  │ 扩缩容 │ v2.0.0 │ 5→8副本   │[-] │   ║
║ │ 05-29 10:00:00  │ 更新   │ v2.0.0 │ v1.1→2.0 │[回]│   ║
║ │ 05-28 14:00:00  │ 重启   │ v1.1.0 │ -         │[-] │   ║
║ └────────────────────────────────────────────────────┘   ║
╚════════════════════════════════════════════════════════════╝

[回] = 回滚到此版本
```

## 业务流程

### 场景1：CI自动部署

```
1. CI构建完成
   ↓
2. 调用 POST /api/v1/app-deployments/1/deploy
   Request: {
     "version": "v2.2.0",
     "imageUrl": "nginx:1.26-alpine"
   }
   ↓
3. 系统处理：
   - 查询主记录（ID=1）
   - 创建历史记录（type=update）
   - 执行K8s部署（更新镜像）
   - 更新主记录状态
   ↓
4. 返回结果
   Response: {
     "historyId": 110,
     "message": "部署成功"
   }
```

### 场景2：运维扩容

```
1. 前端点击"扩缩容"按钮
   ↓
2. 弹出对话框，输入副本数：10
   ↓
3. 调用 POST /api/v1/app-deployments/1/scale
   Request: {
     "replicas": 10,
     "reason": "流量高峰"
   }
   ↓
4. 系统处理：
   - 创建历史记录（type=scale, changes={"replicas":"5→10"}）
   - 执行 kubectl scale
   - 更新主记录 desired_replicas=10
   ↓
5. 前端刷新，显示新的副本数
```

### 场景3：查看部署历史

```
1. 用户点击"详情"按钮
   ↓
2. 进入部署详情页
   ↓
3. 自动加载：
   - GET /api/v1/app-deployments/1 (主记录)
   - GET /api/v1/app-deployments/1/history (历史记录)
   ↓
4. 展示：
   - 当前状态（版本、副本数、镜像）
   - Pod列表
   - 历史记录时间线
   ↓
5. 用户可以：
   - 点击"回滚"按钮回到历史版本
   - 查看每次变更的详细信息
```

## 优势对比

### 旧架构

```
deployments 表
├── id=42: app-8-canary, v1.0, 5副本
├── id=43: app-8-canary, v1.1, 5副本
├── id=44: app-8-canary, v1.1, 8副本
└── id=47: app-8-canary, v2.0, 5副本

问题：
❌ 无法快速定位"当前部署"
❌ 重启/扩缩容操作哪条记录？
❌ 历史记录混乱
❌ 数据冗余
```

### 新架构

```
app_deployments (主记录)
└── id=1: app-8, v2.0, 5副本 [当前状态]

deployment_history (历史)
├── id=101: create, v1.0, 5副本
├── id=102: update, v1.1, 5副本
├── id=103: scale,  v1.1, 8副本
└── id=104: update, v2.0, 5副本

优势：
✅ 一个应用一条主记录，清晰明确
✅ 所有操作针对主记录
✅ 历史记录独立管理
✅ 易于查询和展示
```

## 实施步骤

1. **创建新表结构**
2. **数据迁移**（从旧表导入）
3. **修改API逻辑**
4. **前端适配**
5. **灰度上线**
6. **清理旧表**
