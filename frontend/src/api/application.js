import request from '@/utils/request'

// 获取应用列表
export const getApplicationList = (params) => {
  return request({
    url: '/applications/',
    method: 'get',
    params
  })
}

// 获取应用详情
export const getApplicationDetail = (id) => {
  return request({
    url: `/applications/${id}/`,
    method: 'get'
  })
}

// 创建应用
export const createApplication = (data) => {
  return request({
    url: '/applications/',
    method: 'post',
    data
  })
}

// 更新应用
export const updateApplication = (id, data) => {
  return request({
    url: `/applications/${id}/`,
    method: 'put',
    data
  })
}

// 删除应用
export const deleteApplication = (id) => {
  return request({
    url: `/applications/${id}/`,
    method: 'delete'
  })
}

// 获取组件列表
export const getComponentList = (params) => {
  return request({
    url: '/components/',
    method: 'get',
    params
  })
}

// 创建组件
export const createComponent = (data) => {
  return request({
    url: '/components/',
    method: 'post',
    data
  })
}

// 更新组件
export const updateComponent = (id, data) => {
  return request({
    url: `/components/${id}/`,
    method: 'put',
    data
  })
}

// 删除组件
export const deleteComponent = (id) => {
  return request({
    url: `/components/${id}/`,
    method: 'delete'
  })
}
