import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as apiLogin } from '../api'
import type { User } from '../types'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  const login = async (username: string, password: string) => {
    const res = await apiLogin(username, password)
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
  }

  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, login, logout }
})
