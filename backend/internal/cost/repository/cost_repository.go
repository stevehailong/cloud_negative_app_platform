package repository

import (
	"fmt"
	"log"
	"time"

	"my-cloud/internal/common/model"
	"my-cloud/pkg/prometheus"

	"my-cloud/pkg/kubecost"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CostRepository struct {
	db *gorm.DB
}

func NewCostRepository(db *gorm.DB) *CostRepository {
	return &CostRepository{db: db}
}

// Create 创建成本记录
func (r *CostRepository) Create(record *model.CostRecord) error {
	return r.db.Create(record).Error
}

// GetByID 根据ID查询成本记录
func (r *CostRepository) GetByID(id uint) (*model.CostRecord, error) {
	var record model.CostRecord
	err := r.db.Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Delete 删除成本记录
func (r *CostRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.CostRecord{}).Error
}

// List 分页查询成本记录列表
func (r *CostRepository) List(offset, limit int, clusterID uint, projectID, appID *uint, namespace, startDate, endDate string) ([]model.CostRecord, int64, error) {
	var records []model.CostRecord
	var total int64

	query := r.db.Model(&model.CostRecord{})

	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if appID != nil {
		query = query.Where("app_id = ?", *appID)
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if startDate != "" {
		query = query.Where("cost_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("cost_date <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("cost_date DESC, id DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	// 批量解析应用名称和项目名称
	r.batchResolveNames(records)

	return records, total, nil
}

// batchResolveNames 批量解析 app_id → app_name, project_name
func (r *CostRepository) batchResolveNames(records []model.CostRecord) {
	if len(records) == 0 || r.db == nil {
		return
	}
	// 收集唯一的 app_id
	appIDs := make(map[uint]bool)
	for _, rec := range records {
		if rec.AppID != nil {
			appIDs[*rec.AppID] = true
		}
	}
	if len(appIDs) == 0 {
		return
	}
	ids := make([]uint, 0, len(appIDs))
	for id := range appIDs {
		ids = append(ids, id)
	}

	type AppInfo struct {
		ID        uint   `gorm:"column:id"`
		Name      string `gorm:"column:name"`
		ProjectID uint   `gorm:"column:project_id"`
	}
	var apps []AppInfo
	// 跨库查询 app_db.applications
	_ = r.db.Raw("SELECT id, name, project_id FROM app_db.applications WHERE id IN ?", ids).Scan(&apps).Error

	appMap := make(map[uint]string)
	projectIDs := make(map[uint]bool)
	appProjectMap := make(map[uint]uint) // app_id → project_id
	for _, a := range apps {
		appMap[a.ID] = a.Name
		projectIDs[a.ProjectID] = true
		appProjectMap[a.ID] = a.ProjectID
	}
	// 解析 project names
	projectMap := make(map[uint]string)
	if len(projectIDs) > 0 {
		pidList := make([]uint, 0, len(projectIDs))
		for pid := range projectIDs {
			pidList = append(pidList, pid)
		}
		type ProjInfo struct {
			ID   uint   `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		var projs []ProjInfo
		_ = r.db.Raw("SELECT id, project_name as name FROM org_db.projects WHERE id IN ?", pidList).Scan(&projs).Error
		for _, p := range projs {
			projectMap[p.ID] = p.Name
		}
	}

	// 填充
	for i := range records {
		if records[i].AppID != nil {
			records[i].AppName = appMap[*records[i].AppID]
			if pid, ok := appProjectMap[*records[i].AppID]; ok {
				records[i].ProjectName = projectMap[pid]
			}
		}
	}
}

// CostOverview 成本概览聚合结果
type CostOverview struct {
	TotalCPUCost     float64                `json:"totalCpuCost"`
	TotalMemoryCost  float64                `json:"totalMemoryCost"`
	TotalStorageCost float64                `json:"totalStorageCost"`
	TotalNetworkCost float64                `json:"totalNetworkCost"`
	TotalCost        float64                `json:"totalCost"`
	ByProject        []ProjectCostSummary   `json:"byProject"`
	ByApp            []AppCostSummary       `json:"byApp"`
	RecordCount      int64                  `json:"recordCount"`
}

// ProjectCostSummary 项目成本汇总
type ProjectCostSummary struct {
	ProjectID    uint    `json:"projectId"`
	TotalCost    float64 `json:"totalCost"`
	CPUCost      float64 `json:"cpuCost"`
	MemoryCost   float64 `json:"memoryCost"`
	StorageCost  float64 `json:"storageCost"`
	NetworkCost  float64 `json:"networkCost"`
	RecordCount  int64   `json:"recordCount"`
}

// AppCostSummary 应用成本汇总
type AppCostSummary struct {
	AppID        uint    `json:"appId"`
	ProjectID    uint    `json:"projectId"`
	TotalCost    float64 `json:"totalCost"`
	CPUCost      float64 `json:"cpuCost"`
	MemoryCost   float64 `json:"memoryCost"`
	StorageCost  float64 `json:"storageCost"`
	NetworkCost  float64 `json:"networkCost"`
	RecordCount  int64   `json:"recordCount"`
}

// GetOverview 获取成本概览，按项目和应用汇总
func (r *CostRepository) GetOverview(clusterID uint, startDate, endDate string) (*CostOverview, error) {
	overview := &CostOverview{}

	query := r.db.Model(&model.CostRecord{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if startDate != "" {
		query = query.Where("cost_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("cost_date <= ?", endDate)
	}

	// 总计数和总成本汇总
	type AggResult struct {
		TotalCPUCost     float64
		TotalMemoryCost  float64
		TotalStorageCost float64
		TotalNetworkCost float64
		TotalCost        float64
		RecordCount      int64
	}
	var agg AggResult
	err := query.Select(
		"COALESCE(SUM(cpu_cost), 0) as total_cpu_cost",
		"COALESCE(SUM(memory_cost), 0) as total_memory_cost",
		"COALESCE(SUM(storage_cost), 0) as total_storage_cost",
		"COALESCE(SUM(network_cost), 0) as total_network_cost",
		"COALESCE(SUM(total_cost), 0) as total_cost",
		"COUNT(*) as record_count",
	).Scan(&agg).Error
	if err != nil {
		return nil, err
	}
	overview.TotalCPUCost = agg.TotalCPUCost
	overview.TotalMemoryCost = agg.TotalMemoryCost
	overview.TotalStorageCost = agg.TotalStorageCost
	overview.TotalNetworkCost = agg.TotalNetworkCost
	overview.TotalCost = agg.TotalCost
	overview.RecordCount = agg.RecordCount

	// 按项目汇总（project_id 可能为空，排除）
	var projectSummaries []ProjectCostSummary
	err = r.db.Model(&model.CostRecord{}).
		Where("project_id IS NOT NULL").
		Scopes(func(db *gorm.DB) *gorm.DB {
			if clusterID > 0 {
				db = db.Where("cluster_id = ?", clusterID)
			}
			if startDate != "" {
				db = db.Where("cost_date >= ?", startDate)
			}
			if endDate != "" {
				db = db.Where("cost_date <= ?", endDate)
			}
			return db
		}).
		Select(
			"project_id",
			"COALESCE(SUM(total_cost), 0) as total_cost",
			"COALESCE(SUM(cpu_cost), 0) as cpu_cost",
			"COALESCE(SUM(memory_cost), 0) as memory_cost",
			"COALESCE(SUM(storage_cost), 0) as storage_cost",
			"COALESCE(SUM(network_cost), 0) as network_cost",
			"COUNT(*) as record_count",
		).
		Group("project_id").
		Order("total_cost DESC").
		Scan(&projectSummaries).Error
	if err != nil {
		return nil, err
	}
	overview.ByProject = projectSummaries

	// 按应用汇总（app_id 可能为空，排除）
	var appSummaries []AppCostSummary
	err = r.db.Model(&model.CostRecord{}).
		Where("app_id IS NOT NULL").
		Scopes(func(db *gorm.DB) *gorm.DB {
			if clusterID > 0 {
				db = db.Where("cluster_id = ?", clusterID)
			}
			if startDate != "" {
				db = db.Where("cost_date >= ?", startDate)
			}
			if endDate != "" {
				db = db.Where("cost_date <= ?", endDate)
			}
			return db
		}).
		Select(
			"app_id",
			"COALESCE(MAX(project_id), 0) as project_id",
			"COALESCE(SUM(total_cost), 0) as total_cost",
			"COALESCE(SUM(cpu_cost), 0) as cpu_cost",
			"COALESCE(SUM(memory_cost), 0) as memory_cost",
			"COALESCE(SUM(storage_cost), 0) as storage_cost",
			"COALESCE(SUM(network_cost), 0) as network_cost",
			"COUNT(*) as record_count",
		).
		Group("app_id").
		Order("total_cost DESC").
		Scan(&appSummaries).Error
	if err != nil {
		return nil, err
	}
	overview.ByApp = appSummaries

	return overview, nil
}

// GetCostByProject 查询指定项目的成本记录
func (r *CostRepository) GetCostByProject(projectID uint, startDate, endDate string) ([]model.CostRecord, error) {
	var records []model.CostRecord
	query := r.db.Where("project_id = ?", projectID)
	if startDate != "" {
		query = query.Where("cost_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("cost_date <= ?", endDate)
	}
	err := query.Order("cost_date DESC").Find(&records).Error
	return records, err
}

// GetCostByApp 查询指定应用的成本记录
func (r *CostRepository) GetCostByApp(appID uint, startDate, endDate string) ([]model.CostRecord, error) {
	var records []model.CostRecord
	query := r.db.Where("app_id = ?", appID)
	if startDate != "" {
		query = query.Where("cost_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("cost_date <= ?", endDate)
	}
	err := query.Order("cost_date DESC").Find(&records).Error
	return records, err
}

// SyncFromPrometheus 从 Prometheus 查询集群指标并估算成本写入 cost_records
// promURL: Prometheus 地址，如 http://prometheus:9090
// clusterID: 目标集群 ID
// costDate: 成本归属日期 (YYYY-MM-DD)
//
// 修复要点：
//  1. 先将 CPU/内存/存储/网络四个维度的 PromQL 结果在内存中按 namespace 聚合
//  2. 每个 namespace 只做一次原子 upsert，避免 SELECT-then-INSERT 竞态
//  3. OnConflict DoUpdates 用 = 赋值而非 += 累加，避免重复同步时成本翻倍
func (r *CostRepository) SyncFromPrometheus(promURL string, clusterID uint, costDate string) (int, error) {
	if promURL == "" {
		return 0, fmt.Errorf("prometheus URL is empty")
	}

	client := prometheus.NewClient(promURL)

	// 测试连通性
	if err := client.Ping(); err != nil {
		log.Printf("[CostRepo] Prometheus ping failed: %v, continuing anyway", err)
	}

	now := time.Now()

	// 单价常量
	const (
		cpuCostPerCorePerHour     = 0.03
		memCostPerGBPerHour       = 0.01
		storageCostPerGBPerMonth  = 0.10
		networkCostPerGB          = 0.01
	)

	// 第一步：收集所有维度的成本数据，按 namespace 聚合到 map
	agg := make(map[string]*model.CostRecord)

	// --- CPU ---
	cpuQuery := `sum(rate(container_cpu_usage_seconds_total{container!=""}[5m])) by (namespace)`
	if cpuResp, err := client.Query(cpuQuery, time.Time{}); err != nil {
		log.Printf("[CostRepo] CPU query failed: %v", err)
	} else {
		for _, result := range cpuResp.Data.Result {
			ns := result.Metric["namespace"]
			if ns == "" {
				continue
			}
			cpuCores := parseValue(result.Value)
			cpuCost := cpuCores * cpuCostPerCorePerHour * 24 // 按天估算
			ensureRecord(agg, clusterID, ns, costDate, now).CPUCost = cpuCost
			log.Printf("[CostRepo] CPU ns=%s cores=%.4f cost=%.4f", ns, cpuCores, cpuCost)
		}
	}

	// --- 内存 ---
	memQuery := `sum(container_memory_usage_bytes{container!=""}) by (namespace) / 1024 / 1024 / 1024`
	if memResp, err := client.Query(memQuery, time.Time{}); err != nil {
		log.Printf("[CostRepo] Memory query failed: %v", err)
	} else {
		for _, result := range memResp.Data.Result {
			ns := result.Metric["namespace"]
			if ns == "" {
				continue
			}
			memGB := parseValue(result.Value)
			memCost := memGB * memCostPerGBPerHour * 24
			ensureRecord(agg, clusterID, ns, costDate, now).MemoryCost = memCost
			log.Printf("[CostRepo] Memory ns=%s gb=%.4f cost=%.4f", ns, memGB, memCost)
		}
	}

	// --- 存储 ---
	storageQuery := `sum(kube_persistentvolumeclaim_resource_requests_storage_bytes) by (namespace) / 1024 / 1024 / 1024`
	if storageResp, err := client.Query(storageQuery, time.Time{}); err != nil {
		log.Printf("[CostRepo] Storage query failed: %v", err)
	} else {
		for _, result := range storageResp.Data.Result {
			ns := result.Metric["namespace"]
			if ns == "" {
				continue
			}
			storageGB := parseValue(result.Value)
			storageCost := storageGB * storageCostPerGBPerMonth / 30
			ensureRecord(agg, clusterID, ns, costDate, now).StorageCost = storageCost
			log.Printf("[CostRepo] Storage ns=%s gb=%.4f cost=%.4f", ns, storageGB, storageCost)
		}
	}

	// --- 网络 ---
	networkQuery := `sum(rate(container_network_transmit_bytes_total{container!=""}[5m])) by (namespace) / 1024 / 1024 / 1024`
	if networkResp, err := client.Query(networkQuery, time.Time{}); err != nil {
		log.Printf("[CostRepo] Network query failed: %v", err)
	} else {
		for _, result := range networkResp.Data.Result {
			ns := result.Metric["namespace"]
			if ns == "" {
				continue
			}
			netGBps := parseValue(result.Value)
			netGBPerDay := netGBps * 86400
			netCost := netGBPerDay * networkCostPerGB
			ensureRecord(agg, clusterID, ns, costDate, now).NetworkCost = netCost
			log.Printf("[CostRepo] Network ns=%s gbps=%.6f cost=%.4f", ns, netGBps, netCost)
		}
	}

	// 第二步：计算 totalCost 并原子 upsert 每条记录
	synced := 0
	for ns, record := range agg {
		record.TotalCost = record.CPUCost + record.MemoryCost + record.StorageCost + record.NetworkCost
		if err := r.upsertCostRecord(record); err != nil {
			log.Printf("[CostRepo] Failed to upsert cost record for ns=%s: %v", ns, err)
			continue
		}
		synced++
	}

	log.Printf("[CostRepo] Sync complete: %d namespaces synced for cluster=%d date=%s", synced, clusterID, costDate)
	return synced, nil
}

// ensureRecord 从聚合 map 中获取或创建指定 namespace 的成本记录
func ensureRecord(agg map[string]*model.CostRecord, clusterID uint, namespace, costDate string, now time.Time) *model.CostRecord {
	if r, ok := agg[namespace]; ok {
		return r
	}
	appID, projectID := resolveAppFromNamespace(namespace)
	r := &model.CostRecord{
		ClusterID:  clusterID,
		Namespace:  namespace,
		CostDate:   costDate,
		Source:     "prometheus",
		CreateTime: now,
		AppID:      appID,
		ProjectID:  projectID,
	}
	agg[namespace] = r
	return r
}

// resolveAppFromNamespace 从 namespace 格式 app-{appID}-{envNs} 中提取 appID 和 projectID
func resolveAppFromNamespace(namespace string) (*uint, *uint) {
	// namespace 格式: app-{appID}-{envNamespace}
	var appID uint
	n, _ := fmt.Sscanf(namespace, "app-%d-", &appID)
	if n == 1 && appID > 0 {
		aid := new(uint)
		*aid = appID
		// projectID 需要查 applications 表，在 batch resolve 时填充
		return aid, nil
	}
	return nil, nil
}

// upsertCostRecord 原子 upsert：存在则替换成本值，不存在则插入
// 依赖数据库唯一索引 idx_unique_cost (cluster_id, namespace, cost_date, source)
func (r *CostRepository) upsertCostRecord(record *model.CostRecord) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "cluster_id"},
			{Name: "namespace"},
			{Name: "cost_date"},
			{Name: "source"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"cpu_cost",
			"memory_cost",
			"storage_cost",
			"network_cost",
			"total_cost",
			"create_time",
			"project_id",
			"app_id",
		}),
	}).Create(record).Error
}

// SyncFromKubecost 从 Kubecost API 获取成本数据（优先使用）
func (r *CostRepository) SyncFromKubecost(kubecostURL string, clusterID uint, costDate string) (int, error) {
	if kubecostURL == "" {
		return 0, fmt.Errorf("kubecost URL is empty")
	}

	client := kubecost.NewClient(kubecostURL)
	if err := client.Ping(); err != nil {
		return 0, fmt.Errorf("kubecost ping failed: %w", err)
	}

	costs, err := client.GetAllocation()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	synced := 0

	for _, ns := range costs {
		if ns.Namespace == "" {
			continue
		}
		appID, projectID := resolveAppFromNamespace(ns.Namespace)
		record := &model.CostRecord{
			ClusterID:   clusterID,
			Namespace:   ns.Namespace,
			CostDate:    costDate,
			CPUCost:     ns.CPUCost,
			MemoryCost:  ns.MemoryCost,
			StorageCost: ns.StorageCost,
			NetworkCost: ns.NetworkCost,
			TotalCost:   ns.TotalCost,
			Source:      "kubecost",
			CreateTime:  now,
			AppID:       appID,
			ProjectID:   projectID,
		}
		if err := r.upsertCostRecord(record); err != nil {
			log.Printf("[CostRepo] Failed to upsert kubecost record for ns=%s: %v", ns.Namespace, err)
			continue
		}
		synced++
	}

	log.Printf("[CostRepo] Kubecost sync complete: %d namespaces for cluster=%d date=%s", synced, clusterID, costDate)
	return synced, nil
}

// parseValue 从 Prometheus 查询结果的 value 元组中提取 float64
func parseValue(v []interface{}) float64 {
	if len(v) < 2 {
		return 0
	}
	switch s := v[1].(type) {
	case string:
		var f float64
		if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
			return 0
		}
		return f
	case float64:
		return s
	}
	return 0
}
