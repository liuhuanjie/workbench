<template>
  <div class="login-wrap">
    <div class="login-card">
      <n-icon :component="CubeOutline" :size="34" color="#2563eb" />
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
import { useMessage } from 'naive-ui'
import { useAuthStore } from '../stores/auth'
import { CubeOutline } from '../icons'

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
  background: #f5f6f8;
}

.login-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 32px 28px;
  width: 340px;
  text-align: center;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.06);
}

.login-card h1 {
  font-size: 17px;
  margin: 10px 0 20px;
  color: #1f2937;
  font-weight: 600;
}

.login-card .n-input {
  margin-bottom: 14px;
  text-align: left;
}
</style>
