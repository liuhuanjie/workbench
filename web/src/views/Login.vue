<template>
  <div class="login-wrap">
    <div class="login-card">
      <Icon icon="fluent-emoji:house-with-garden" width="56" />
      <h1>资产工作台</h1>
      <n-input v-model:value="username" placeholder="用户名" size="large" @keyup.enter="doLogin" />
      <n-input
        v-model:value="password"
        type="password"
        show-password-on="click"
        placeholder="密码"
        size="large"
        @keyup.enter="doLogin"
      />
      <n-button type="primary" size="large" block :loading="loading" @click="doLogin">
        登 录
      </n-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useMessage } from 'naive-ui'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const message = useMessage()
const auth = useAuthStore()
const username = ref('')
const password = ref('')
const loading = ref(false)

async function doLogin() {
  if (!username.value || !password.value) {
    message.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    message.success('登录成功')
    router.push(auth.mustChange ? '/settings' : '/')
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #14b8a6 0%, #38bdf8 100%);
}

.login-card {
  background: #fff;
  border-radius: 16px;
  padding: 40px 36px;
  width: 360px;
  text-align: center;
  box-shadow: 0 10px 40px rgba(19, 78, 74, 0.15);
}

.login-card h1 {
  font-size: 20px;
  margin: 12px 0 24px;
  color: #134e4a;
}

.login-card .n-input {
  margin-bottom: 14px;
  text-align: left;
}
</style>
