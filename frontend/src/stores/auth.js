import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const reviewEnabled = computed(() => user.value?.reviewEnabled !== false)

  async function login(username, password) {
    const deviceKey = generateDeviceKey()
    const res = await authApi.login({ username, password, deviceKey })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
    return res
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {}
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function fetchMe() {
    const res = await authApi.me()
    user.value = res.data
    localStorage.setItem('user', JSON.stringify(res.data))
    return res.data
  }

  function generateDeviceKey() {
    const nav = window.navigator
    return btoa([nav.userAgent, screen.width, screen.height, nav.language].join('|')).slice(0, 50)
  }

  return { token, user, isLoggedIn, isAdmin, reviewEnabled, login, logout, fetchMe }
})
