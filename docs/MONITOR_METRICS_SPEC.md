# 监控指标口径说明文档

> 更新时间: 2026-06-04

## 一、数据来源

监控数据来自 **Prometheus**（配置了 Prometheus 时）或 **Kubernetes Pod 估算**（回退方案）。

### 数据源优先级

| 优先级 | 数据源 | 说明 |
|--------|--------|------|
| 1 | Prometheus - K8s 容器指标 | `container_cpu_usage_seconds_total`, `container_memory_working_set_bytes` |
| 2 | Prometheus - mycloud 自定义指标 | `go_goroutines`, `go_memstats_alloc_bytes`, `mycloud_*` |
| 3 | K8s API - Pod resources | 通过 Pod `resources.requests/limits` 估算 |
| 4 | 无数据 | 返回 0，`data_source: "none"` |

### 当前 Prometheus 采集状态

| 指标类型 | 可用性 | 说明 |
|----------|--------|------|
| mycloud 自定义指标 | ✅ 可用 | monitor-service 自身暴露的指标 |
| K8s 容器指标 (cAdvisor) | ❌ 未配置 | 需在 Prometheus 中添加 `kubernetes_sd_configs` |
| 平台服务 /metrics | ❌ 大部分 DOWN | 除 monitor-service 外其他服务未暴露 /metrics 端点 |

---

## 二、指标枚举值含义

### 1. 监控对象类型 (type)

| 值 | 含义 | 筛选依据 |
|----|------|----------|
| `app` | 应用监控 | 按 K8s 标签 `app=<应用名>` 或 Prometheus 标签 `service=<应用名>` 过滤 |
| `environment` | 环境监控 | 按 K8s 标签 `env=<环境名>` 过滤 |
| `cluster` | 集群监控 | 全部应用聚合，不额外筛选 |

### 2. 时间范围 (timeRange)

| 值 | 窗口 | Prometheus rate 窗口 |
|----|------|---------------------|
| `1h` | 最近 1 小时 | `[5m]` |
| `6h` | 最近 6 小时 | `[15m]` |
| `24h` | 最近 24 小时 | `[1h]` |
| `7d` | 最近 7 天 | `[6h]` |

### 3. 数据源标记 (data_source)

| 值 | 含义 |
|----|------|
| `prometheus` | 从 Prometheus 查询，指标按应用标签过滤 |
| `k8s` | 基于 K8s Pod 的 resources.requests/limits 估算 |
| `none` | 无可用数据源，所有指标为 0 |

---

## 三、指标计算口径

### CPU使用率

| 数据源 | 计算公式 | 说明 |
|--------|----------|------|
| **K8s 容器** (优先) | `sum(rate(container_cpu_usage_seconds_total{container_label_app="<app>"}[5m])) × 100` | 容器实际 CPU 使用率，按应用的 app 标签过滤。**当前 Prometheus 未配置 K8s 采集，此指标不可用** |
| **mycloud 自定义** (回退) | `go_goroutines{service="<app>"} × 5` | Go goroutine 数 × 5 作为负载代理。100 goroutine ≈ 100%。**仅当应用暴露 /metrics 端点且按 service 标签上报时有效** |
| **K8s API** (回退) | `(sum(requests.cpu) / sum(limits.cpu)) × 100 × 0.7` | 基于 Pod 声明的 CPU requests/limits 比率 × 衰减因子 0.7。**这不是实际使用率** |
| **无数据** | `0` | 无 Prometheus 无 K8s 客户端时 |

### 内存使用率

| 数据源 | 计算公式 | 说明 |
|--------|----------|------|
| **K8s 容器** (优先) | `sum(container_memory_working_set_bytes{container_label_app="<app>"}) / sum(kube_pod_container_resource_limits{resource="memory"}) × 100` | 工作集内存 / 内存 limit。**当前不可用** |
| **mycloud 自定义** (回退) | `mycloud_memory_usage_percent{service="<app>"}` | 应用通过 Go runtime.ReadMemStats 上报的堆内存/512MB。**仅已部署且上报指标的应用有效** |
| **K8s API** (回退) | `(sum(requests.memory) / sum(limits.memory)) × 100 × 0.6` | 基于 Pod 声明的内存 requests/limits × 衰减因子 0.6。**不是实际使用率** |
| **无数据** | `0` | 无数据源或无部署应用 |

### 请求QPS

| 数据源 | 计算公式 | 说明 |
|--------|----------|------|
| **mycloud 自定义** | `sum(rate(mycloud_http_requests_total{service="<app>"}[5m]))` | 按应用 service 标签过滤的 HTTP 请求速率。**无已部署应用时返回 0** |
| **K8s API** | `runningPods × 15` | 运行中 Pod 数 × 15 的硬编码估算。**不反映真实流量** |
| **无数据** | `0` | |

### 错误率

| 数据源 | 计算公式 | 说明 |
|--------|----------|------|
| **mycloud 自定义** | `sum(rate(mycloud_http_requests_total{service="<app>",status=~"5.."}[5m])) / sum(rate(mycloud_http_requests_total{service="<app>"}[5m])) × 100` | 应用的 5xx 错误占比，按 service 标签过滤 |
| **K8s API** | `0` | K8s 回退方案不估算错误率，恒为 0 |
| **无数据** | `0` | |

---

## 四、Grafana 仪表盘

### 嵌入展示

前端通过 iframe 嵌入 Grafana 仪表盘：

```
{grafanaUrl}/d/my-cloud-overview/my-cloud-监控概览?kiosk=tv&theme=light&from=now-{t}&to=now
```

| 参数 | 含义 |
|------|------|
| `kiosk=tv` | 隐藏 Grafana 导航栏，仅显示图表 |
| `theme=light` | 浅色主题 |
| `from/to` | 根据前端时间范围选择器映射 (`1h→now-1h`, `6h→now-6h`, `24h→now-24h`, `7d→now-7d`) |

### 筛选传递

当前 Grafana iframe **不传递**应用/环境/集群筛选参数到 Grafana 变量。Grafana 显示的是全局仪表盘默认视图。

---

## 五、已知限制

1. **无 K8s 容器指标**：Prometheus 未配置 `kubernetes_sd_configs`，无法获取 `container_cpu_usage_seconds_total` 等真实容器指标
2. **mycloud 指标仅 monitor-service 上报**：其他平台服务（gateway、auth-service 等）未暴露 `/metrics` 端点
3. **无趋势数据**：所有 `*Trend` 字段均为空字符串，趋势功能未实现
4. **Grafana 不支持应用级筛选**：iframe 仅传递时间范围，不传递应用/环境变量
5. **K8s 回退为估算值**：CPU/内存/QPS 基于 Pod resources 的硬编码估算，不反映实际使用情况
6. **无自动刷新**：指标仅在手动点击"查询"时获取，无轮询机制
