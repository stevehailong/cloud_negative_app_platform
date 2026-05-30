import request from '@/utils/request'

// 登录
export const login = (data) => {
  return request({
    url: '/auth/login',
    method: 'post',
    data
  })
}

// 注册
export const register = (data) => {
  return request({
    url: '/auth/register',
    method: 'post',
    data
  })
}

// 获取用户信息
export const getUserInfo = () => {
  return request({
    url: '/auth/userinfo',
    method: 'get'
  })
}

// 退出登录
export const logout = () => {
  return request({
    url: '/auth/logout',
    method: 'post'
  })
}

// 修改密码
export const updatePassword = (data) => {
  return request({
    url: '/auth/password',
    method: 'put',
    data
  })
}

// 更新用户资料
export const updateProfile = (data) => {
  return request({
    url: '/auth/profile',
    method: 'put',
    data
  })
}
