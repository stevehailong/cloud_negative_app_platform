import request from '@/utils/request'

/**
 * 获取流水线列表
 */
export function getPipelineList(params) {
  return request({
    url: '/pipelines',
    method: 'get',
    params
  })
}

/**
 * 获取流水线详情
 */
export function getPipelineDetail(id) {
  return request({
    url: `/pipelines/${id}`,
    method: 'get'
  })
}

/**
 * 创建流水线
 */
export function createPipeline(data) {
  return request({
    url: '/pipelines',
    method: 'post',
    data
  })
}

/**
 * 更新流水线
 */
export function updatePipeline(id, data) {
  return request({
    url: `/pipelines/${id}`,
    method: 'put',
    data
  })
}

/**
 * 删除流水线
 */
export function deletePipeline(id) {
  return request({
    url: `/pipelines/${id}`,
    method: 'delete'
  })
}

/**
 * 触发流水线执行
 */
export function triggerPipeline(id, data) {
  return request({
    url: `/pipelines/${id}/run`,
    method: 'post',
    data
  })
}

export function deployPipeline(id) {
  return request({
    url: `/pipelines/${id}/deploy`,
    method: 'post'
  })
}

/**
 * 获取流水线执行记录列表
 */
export function getPipelineRunList(params) {
  return request({
    url: '/pipeline-runs',
    method: 'get',
    params
  })
}

/**
 * 获取流水线执行记录详情
 */
export function getPipelineRunDetail(id) {
  return request({
    url: `/pipeline-runs/${id}`,
    method: 'get'
  })
}

/**
 * 停止流水线执行
 */
export function stopPipelineRun(id) {
  return request({
    url: `/pipeline-runs/${id}/stop`,
    method: 'post'
  })
}

/**
 * 获取构建产物列表
 */
export function getArtifactList(params) {
  return request({
    url: '/artifacts',
    method: 'get',
    params
  })
}
