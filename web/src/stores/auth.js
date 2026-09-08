import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, getToken, setToken, setMustChange, clearToken, getMustChange } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(getToken())
  const mustChange = ref(getMustChange())

  async function login(username, password) {
    const data = await api.login(username, password)
    token.value = data.token
    mustChange.value = !!data.must_change_password
    setToken(data.token)
    setMustChange(mustChange.value)
  }

  async function changePassword(oldPassword, newPassword) {
    await api.changePassword(oldPassword, newPassword)
    mustChange.value = false
    setMustChange(false)
  }

  function logout() {
    clearToken()
    token.value = ''
    mustChange.value = false
  }

  return { token, mustChange, login, changePassword, logout }
})
