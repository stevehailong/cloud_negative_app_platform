import request from '@/utils/request'

// 获取用户列表
export function getUserList(params) {
  return request({
    url: '/users/',
    method: 'get',
    params
  })
}

// 获取用户详情
export function getUserDetail(id) {
  return request({
    url: `/users/${id}/`,
    method: 'get'
  })
}

// 更新用户状态
export function updateUserStatus(id, status) {
  return request({
    url: `/users/${id}/status/`,
    method: 'put',
    data: { status }
  })
}

// 为用户分配角色
export function assignRoles(userId, roleIds) {
  return request({
    url: '/users/assign-roles/',
    method: 'post',
    data: { userId, roleIds }
  })
}

// 获取用户的角色
export function getUserRoles(id) {
  return request({
    url: `/users/${id}/roles/`,
    method: 'get'
  })
}

// 获取所有角色列表
export function getRoleList() {
  return request({
    url: '/roles/',
    method: 'get'
  })
}

// 创建用户
export function createUser(data) {
  return request({
    url: '/users/',
    method: 'post',
    data
  })
}
