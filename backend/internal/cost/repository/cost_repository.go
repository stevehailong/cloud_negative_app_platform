package repository

import (
	"fmt"
	"log"
	"time"

	"my-cloud/internal/common/model"
	"my-cloud/pkg/prometheus"

	"gorm.io/gorm"
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

	return records, total, nil
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
	synced := 0

	// ---------- CPU 成本估算 ----------
	// PromQL: 平均每核每小时成本 ≈ $0.03，按命名空间汇总
	cpuCostPerCorePerHour := 0.03
	cpuQuery := `sum(rate(container_cpu_usage_seconds_total{container!=""}[5m])) by (namespace)`
	cpuResp, err := client.Query(cpuQuery, time.Time{})
	if err != nil {
		log.Printf("[CostRepo] CPU query failed: %v", err)
	} else {
		for _, result := range cpuResp.Data.Result {
			namespace := result.Metric["namespace"]
			if namespace == "" {
				continue
			}
			cpuCores := parseValue(result.Value)
			cpuCost := cpuCores * cpuCostPerCorePerHour * 24 // 按天估算

			record := &model.CostRecord{
				ClusterID:   clusterID,
				Namespace:   namespace,
				CostDate:    costDate,
				CPUCost:     cpuCost,
				TotalCost:   cpuCost,
				Source:      "prometheus",
				CreateTime:  now,
			}
			if err := r.upsertRecord(record); err != nil {
				log.Printf("[CostRepo] Failed to upsert CPU record for ns=%s: %v", namespace, err)
			} else {
				synced++
			}
		}
	}

	// ---------- 内存成本估算 ----------
	// PromQL: 平均每 GB 每小时成本 ≈ $0.01
	memCostPerGBPerHour := 0.01
	memQuery := `sum(container_memory_usage_bytes{container!=""}) by (namespace) / 1024 / 1024 / 1024`
	memResp, err := client.Query(memQuery, time.Time{})
	if err != nil {
		log.Printf("[CostRepo] Memory query failed: %v", err)
	} else {
		for _, result := range memResp.Data.Result {
			namespace := result.Metric["namespace"]
			if namespace == "" {
				continue
			}
			memGB := parseValue(result.Value)
			memCost := memGB * memCostPerGBPerHour * 24

			record := &model.CostRecord{
				ClusterID:   clusterID,
				Namespace:   namespace,
				CostDate:    costDate,
				MemoryCost:  memCost,
				TotalCost:   memCost,
				Source:      "prometheus",
				CreateTime:  now,
			}
			if err := r.upsertRecord(record); err != nil {
				log.Printf("[CostRepo] Failed to upsert Memory record for ns=%s: %v", namespace, err)
			} else {
				synced++
			}
		}
	}

	// ---------- 存储成本估算 ----------
	// PromQL: PVC 总容量，按 namespace 汇总，每 GB 每月 ≈ $0.10，折算每日
	storageCostPerGBPerMonth := 0.10
	storageQuery := `sum(kube_persistentvolumeclaim_resource_requests_storage_bytes) by (namespace) / 1024 / 1024 / 1024`
	storageResp, err := client.Query(storageQuery, time.Time{})
	if err != nil {
		log.Printf("[CostRepo] Storage query failed: %v", err)
	} else {
		for _, result := range storageResp.Data.Result {
			namespace := result.Metric["namespace"]
			if namespace == "" {
				continue
			}
			storageGB := parseValue(result.Value)
			storageCost := storageGB * storageCostPerGBPerMonth / 30

			record := &model.CostRecord{
				ClusterID:   clusterID,
				Namespace:   namespace,
				CostDate:    costDate,
				StorageCost: storageCost,
				TotalCost:   storageCost,
				Source:      "prometheus",
				CreateTime:  now,
			}
			if err := r.upsertRecord(record); err != nil {
				log.Printf("[CostRepo] Failed to upsert Storage record for ns=%s: %v", namespace, err)
			} else {
				synced++
			}
		}
	}

	// ---------- 网络成本估算 ----------
	// PromQL: 网络流量总计，按 namespace 汇总，每 GB ≈ $0.01
	networkCostPerGB := 0.01
	networkQuery := `sum(rate(container_network_transmit_bytes_total{container!=""}[5m])) by (namespace) / 1024 / 1024 / 1024`
	networkResp, err := client.Query(networkQuery, time.Time{})
	if err != nil {
		log.Printf("[CostRepo] Network query failed: %v", err)
	} else {
		for _, result := range networkResp.Data.Result {
			namespace := result.Metric["namespace"]
			if namespace == "" {
				continue
			}
			netGBps := parseValue(result.Value)
			netGBPerDay := netGBps * 86400
			netCost := netGBPerDay * networkCostPerGB

			record := &model.CostRecord{
				ClusterID:   clusterID,
				Namespace:   namespace,
				CostDate:    costDate,
				NetworkCost: netCost,
				TotalCost:   netCost,
				Source:      "prometheus",
				CreateTime:  now,
			}
			if err := r.upsertRecord(record); err != nil {
				log.Printf("[CostRepo] Failed to upsert Network record for ns=%s: %v", namespace, err)
			} else {
				synced++
			}
		}
	}

	return synced, nil
}

// upsertRecord 按 cluster_id + namespace + cost_date + source 查找并更新或插入
func (r *CostRepository) upsertRecord(record *model.CostRecord) error {
	var existing model.CostRecord
	err := r.db.Where(
		"cluster_id = ? AND namespace = ? AND cost_date = ? AND source = ?",
		record.ClusterID, record.Namespace, record.CostDate, record.Source,
	).First(&existing).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(record).Error
	}

	// 合并成本：已有成本和新增成本累加
	existing.CPUCost += record.CPUCost
	existing.MemoryCost += record.MemoryCost
	existing.StorageCost += record.StorageCost
	existing.NetworkCost += record.NetworkCost
	existing.TotalCost = existing.CPUCost + existing.MemoryCost + existing.StorageCost + existing.NetworkCost

	return r.db.Save(&existing).Error
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
