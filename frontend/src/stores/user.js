import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login, getUserInfo, logout } from '@/api/auth'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(null)

  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUserInfo = (info) => {
    userInfo.value = info
  }

  const doLogin = async (username, password) => {
    try {
      const res = await login({ username, password })
      setToken(res.data.token)
      setUserInfo({ ...res.data.user, roles: res.data.roles || [] })
      return true
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  const fetchUserInfo = async () => {
    try {
      const res = await getUserInfo()
      setUserInfo(res.data.user)
    } catch (error) {
      console.error('Failed to fetch user info:', error)
    }
  }

  const doLogout = async () => {
    try {
      await logout()
    } catch (error) {
      console.error('Logout failed:', error)
    } finally {
      token.value = ''
      userInfo.value = null
      localStorage.removeItem('token')
      router.push('/login')
    }
  }

  return {
    token,
    userInfo,
    setToken,
    setUserInfo,
    doLogin,
    fetchUserInfo,
    doLogout
  }
})
