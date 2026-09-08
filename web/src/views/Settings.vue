<template>
  <div>
    <h2 class="page-title"><Icon icon="fluent-emoji:gear" width="26" /> 个人设置</h2>

    <div class="panel">
      <div class="panel-head"><h3>修改密码</h3></div>
      <n-form label-placement="left" label-width="90" style="max-width: 420px">
        <n-form-item label="原密码">
          <n-input v-model:value="oldPwd" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="新密码">
          <n-input v-model:value="newPwd" type="password" show-password-on="click"
            placeholder="至少 8 位" />
        </n-form-item>
        <n-form-item label="确认新密码">
          <n-input v-model:value="newPwd2" type="password" show-password-on="click" />
        </n-form-item>
        <n-button type="primary" :loading="loading" @click="submit">保存新密码</n-button>
      </n-form>
      <n-alert v-if="auth.mustChange" type="warning" style="max-width: 420px; margin-top: 14px">
        首次登录请先修改初始密码，修改完成前无法使用其他功能。
      </n-alert>
    </div>

    <div class="panel">
      <div class="panel-head"><h3>会话</h3></div>
      <n-button type="error" secondary @click="logout">退出登录</n-button>
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

const oldPwd = ref('')
const newPwd = ref('')
const newPwd2 = ref('')
const loading = ref(false)

async function submit() {
  if (!oldPwd.value || !newPwd.value) {
    message.warning('请填写完整')
    return
  }
  if (newPwd.value.length < 8) {
    message.warning('新密码至少 8 位')
    return
  }
  if (newPwd.value !== newPwd2.value) {
    message.warning('两次输入的新密码不一致')
    return
  }
  loading.value = true
  try {
    await auth.changePassword(oldPwd.value, newPwd.value)
    message.success('密码已更新')
    oldPwd.value = newPwd.value = newPwd2.value = ''
    if (router.currentRoute.value.query.redirect) {
      router.push('/')
    }
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.panel {
  background: #fff;
  border-radius: 14px;
  padding: 18px 22px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(19, 78, 74, 0.08);
}

.panel-head {
  margin-bottom: 14px;
}

.panel-head h3 {
  margin: 0;
  font-size: 15px;
}
</style>
