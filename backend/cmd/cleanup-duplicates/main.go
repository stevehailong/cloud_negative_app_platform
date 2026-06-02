package main

import (
	"log"
	"my-cloud/internal/deploy/model"
	"my-cloud/pkg/database"
)

func main() {
	// 连接数据库
	dsn := "root:root123456@tcp(127.0.0.1:3306)/deploy_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("========================================")
	log.Println("清理重复的部署记录")
	log.Println("========================================")

	// 1. 查找重复记录
	type DuplicateRecord struct {
		Namespace    string
		WorkloadName string
		Count        int64
		IDs          string
	}

	var duplicates []DuplicateRecord
	err = db.Raw(`
		SELECT 
			namespace, 
			workload_name, 
			COUNT(*) as count,
			GROUP_CONCAT(id ORDER BY update_time DESC) as ids
		FROM app_deployments
		GROUP BY namespace, workload_name
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error

	if err != nil {
		log.Fatalf("Failed to query duplicates: %v", err)
	}

	if len(duplicates) == 0 {
		log.Println("✓ 没有发现重复记录")
		return
	}

	log.Printf("发现 %d 组重复记录:\n", len(duplicates))
	for _, dup := range duplicates {
		log.Printf("  - %s/%s: %d 条记录 (IDs: %s)\n", 
			dup.Namespace, dup.WorkloadName, dup.Count, dup.IDs)
	}

	// 2. 对每组重复记录,保留最新的,删除其他的
	var totalDeleted int
	for _, dup := range duplicates {
		var deployments []model.AppDeployment
		err := db.Where("namespace = ? AND workload_name = ?", dup.Namespace, dup.WorkloadName).
			Order("update_time DESC").
			Find(&deployments).Error
		
		if err != nil {
			log.Printf("✗ 查询失败: %v\n", err)
			continue
		}

		if len(deployments) <= 1 {
			continue
		}

		// 保留第一条(最新的),删除其他的
		keepID := deployments[0].ID
		log.Printf("\n处理 %s/%s:", dup.Namespace, dup.WorkloadName)
		log.Printf("  保留: ID=%d (更新时间: %s)", keepID, deployments[0].UpdateTime)

		for i := 1; i < len(deployments); i++ {
			deleteID := deployments[i].ID
			log.Printf("  删除: ID=%d (更新时间: %s)", deleteID, deployments[i].UpdateTime)
			
			err := db.Delete(&model.AppDeployment{}, deleteID).Error
			if err != nil {
				log.Printf("    ✗ 删除失败: %v", err)
			} else {
				log.Printf("    ✓ 删除成功")
				totalDeleted++
			}
		}
	}

	log.Printf("\n========================================")
	log.Printf("清理完成! 共删除 %d 条重复记录", totalDeleted)
	log.Println("========================================")

	// 3. 验证结果
	var remainingDuplicates []DuplicateRecord
	err = db.Raw(`
		SELECT 
			namespace, 
			workload_name, 
			COUNT(*) as count
		FROM app_deployments
		GROUP BY namespace, workload_name
		HAVING COUNT(*) > 1
	`).Scan(&remainingDuplicates).Error

	if err != nil {
		log.Printf("验证查询失败: %v", err)
		return
	}

	if len(remainingDuplicates) == 0 {
		log.Println("✓ 验证通过: 没有剩余重复记录")
	} else {
		log.Printf("✗ 警告: 仍有 %d 组重复记录", len(remainingDuplicates))
	}

	// 4. 显示最终数据
	var finalDeployments []model.AppDeployment
	err = db.Order("app_id, env_id, workload_name").Find(&finalDeployments).Error
	if err != nil {
		log.Printf("查询最终数据失败: %v", err)
		return
	}

	log.Println("\n最终部署记录:")
	log.Println("ID\tAppID\tEnvID\tNamespace\tWorkloadName\t\tStatus")
	log.Println("--\t-----\t-----\t---------\t------------\t\t------")
	for _, d := range finalDeployments {
		log.Printf("%d\t%d\t%d\t%s\t\t%s\t%s\n", 
			d.ID, d.AppID, d.EnvID, d.Namespace, d.WorkloadName, d.DeploymentStatus)
	}
}
