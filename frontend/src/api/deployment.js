import request from '@/utils/request'

/**
 * 查询应用部署列表
 * @param {Object} params - 查询参数
 * @param {number} params.app_id - 应用ID（可选）
 * @param {number} params.env_id - 环境ID（可选）
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 */
export function getAppDeployments(params) {
  return request({
    url: '/app-deployments',
    method: 'get',
    params
  })
}

/**
 * 获取应用部署详情
 * @param {number} id - 部署ID
 */
export function getAppDeploymentDetail(id) {
  return request({
    url: `/app-deployments/${id}`,
    method: 'get'
  })
}

/**
 * 获取部署历史记录
 * @param {number} id - 部署ID
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 */
export function getDeploymentHistory(id, params) {
  return request({
    url: `/app-deployments/${id}/history`,
    method: 'get',
    params
  })
}

/**
 * 重启部署
 * @param {number} id - 部署ID
 * @param {Object} data - 请求数据
 * @param {number} data.user_id - 用户ID
 */
export function restartDeployment(id, data) {
  return request({
    url: `/app-deployments/${id}/restart`,
    method: 'post',
    data
  })
}

/**
 * 扩缩容
 * @param {number} id - 部署ID
 * @param {Object} data - 请求数据
 * @param {number} data.replicas - 目标副本数
 * @param {number} data.user_id - 用户ID
 */
export function scaleDeployment(id, data) {
  return request({
    url: `/app-deployments/${id}/scale`,
    method: 'post',
    data
  })
}

/**
 * 回滚到历史版本
 * @param {number} id - 部署ID
 * @param {Object} data - 请求数据
 * @param {number} data.history_id - 历史记录ID
 * @param {number} data.user_id - 用户ID
 */
export function rollbackDeployment(id, data) {
  return request({
    url: `/app-deployments/${id}/rollback`,
    method: 'post',
    data
  })
}

/**
 * 部署新版本
 * @param {number} id - 部署ID
 * @param {Object} data - 请求数据
 * @param {string} data.version - 版本号
 * @param {string} data.image_url - 镜像地址
 * @param {number} data.user_id - 用户ID
 */
export function deployNewVersion(id, data) {
  return request({
    url: `/app-deployments/${id}/deploy`,
    method: 'post',
    data
  })
}

/**
 * 获取应用部署的Pod列表
 * @param {number} id - 部署ID
 */
export function getAppDeploymentPods(id) {
  return request({
    url: `/app-deployments/${id}/pods`,
    method: 'get'
  })
}

/**
 * 获取应用部署的事件列表
 * @param {number} id - 部署ID
 */
export function getAppDeploymentEvents(id) {
  return request({
    url: `/app-deployments/${id}/events`,
    method: 'get'
  })
}

/**
 * 调整金丝雀流量权重
 * @param {number} id - 部署ID
 * @param {number} weight - 权重 0-100
 */
export function adjustCanaryWeight(id, weight) {
  return request({
    url: `/app-deployments/${id}/canary/adjust-weight`,
    method: 'post',
    data: { weight }
  })
}

/**
 * 删除应用部署
 * @param {number} id - 部署ID
 */
export function deleteAppDeployment(id) {
  return request({
    url: `/app-deployments/${id}`,
    method: 'delete'
  })
}

// ========== 旧版API（保持兼容）==========

/**
 * 获取部署列表（旧版）
 */
export function getDeployments(params) {
  return request({
    url: '/deployments',
    method: 'get',
    params
  })
}

/**
 * 获取部署详情（旧版）
 */
export function getDeploymentById(id) {
  return request({
    url: `/deployments/${id}`,
    method: 'get'
  })
}

/**
 * 获取Pod列表
 */
export function getDeploymentPods(id) {
  return request({
    url: `/deployments/${id}/pods`,
    method: 'get'
  })
}

/**
 * 获取事件列表
 */
export function getDeploymentEvents(id) {
  return request({
    url: `/deployments/${id}/events`,
    method: 'get'
  })
}

/**
 * 删除部署
 */
export function deleteDeployment(id) {
  return request({
    url: `/deployments/${id}`,
    method: 'delete'
  })
}

/**
 * 删除Pod
 */
export function deletePod(podName) {
  return request({
    url: `/deployments/pods/${podName}`,
    method: 'delete'
  })
}
