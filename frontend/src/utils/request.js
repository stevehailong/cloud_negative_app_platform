import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data
    
    // 兼容不同的成功状态码：0 或 200
    if (res.code !== 0 && res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    
    return res
  },
  (error) => {
    console.error('Request error:', error)
    
    if (error.response) {
      const { status, data } = error.response
      const msg = data?.message || ''
      
      switch (status) {
        case 401:
          ElMessage.error(msg || '未授权，请重新登录')
          const userStore = useUserStore()
          userStore.token = ''
          userStore.userInfo = null
          localStorage.removeItem('token')
          if (router.currentRoute.value.path !== '/login') {
            router.push('/login')
          }
          break
        case 403:
          ElMessage.error(msg || '拒绝访问')
          break
        case 404:
          ElMessage.error(msg || '请求的资源不存在')
          break
        case 500:
          ElMessage.error(msg || '服务器内部错误')
          break
        case 502:
          ElMessage.error('服务暂不可用（502），请稍后重试')
          break
        default:
          ElMessage.error(msg || `请求失败 (${status})`)
      }
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请稍后重试')
    } else {
      ElMessage.error('网络错误，请检查您的网络连接')
    }
    
    return Promise.reject(error)
  }
)

export default request
