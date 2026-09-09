<template>
  <div>
    <h2 class="page-title">
      <n-icon :component="ServerOutline" :size="18" />
      数据管理
    </h2>

    <!-- 数据源健康状态 -->
    <div class="panel">
      <div class="panel-head">
        <h3>数据源状态</h3>
        <n-button size="small" quaternary @click="loadSources">
          <template #icon><n-icon :component="RefreshOutline" :size="15" /></template>
          刷新
        </n-button>
      </div>
      <n-data-table :columns="sourceColumns" :data="sources" :bordered="false" size="small"
        :pagination="false" />
    </div>

    <!-- 运行配置 -->
    <div class="panel">
      <div class="panel-head">
        <h3>运行配置</h3>
      </div>
      <div class="config-form">
        <div class="config-item">
          <span>已读标记每日 0 点重置</span>
          <n-switch v-model:value="dailyReset" @update:value="saveConfig('daily_reset.read_mark', dailyReset ? 'on' : 'off')" />
        </div>
        <div class="config-item">
          <span>业务数据保留天数</span>
          <n-input-number v-model:value="retentionDays" :min="7" :max="365" size="small"
            style="width: 120px" />
        </div>
        <div class="config-item">
          <span>备份保留份数</span>
          <n-input-number v-model:value="backupKeep" :min="3" :max="90" size="small"
            style="width: 120px" />
        </div>
        <n-button type="primary" size="small" @click="saveRetentionAndBackup">保存配置</n-button>
      </div>
    </div>

    <!-- 备份 -->
    <div class="panel">
      <div class="panel-head">
        <h3>数据备份</h3>
        <n-button size="small" type="primary" secondary @click="runBackup">立即备份</n-button>
      </div>
      <n-data-table :columns="backupColumns" :data="backups" :bordered="false" size="small"
        :pagination="false" />
      <div class="migrate-tip">
        迁移整站：旧服务器执行 <code>docker save workbench-web:latest workbench-app:latest | gzip &gt; images.tar.gz</code>
        并打包 <code>data/</code> 目录，新服务器 <code>docker load</code> 后 <code>docker compose up -d</code> 即可完整恢复。
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, h, onMounted } from 'vue'
import { ServerOutline, RefreshOutline } from '../icons'
import { NTag, NButton, NSwitch, useMessage } from 'naive-ui'
import { api } from '../api'

const message = useMessage()

// ---------- 数据源 ----------
const sources = ref([])

const catMap = { news: '新闻', debt: '债权', house: '法拍' }

async function loadSources() {
  try {
    sources.value = await api.sourceList()
  } catch (e) {
    message.error(e.message)
  }
}

async function toggleSource(row, enabled) {
  try {
    await api.sourceToggle(row.source_key, enabled)
    row.enabled = enabled ? 1 : 0
  } catch (e) {
    message.error(e.message)
  }
}

async function runSource(row) {
  try {
    await api.sourceRun(row.source_key)
    message.success(`已触发「${row.name}」抓取，稍后刷新查看`)
  } catch (e) {
    message.error(e.message)
  }
}

const sourceColumns = [
  { title: '数据源', key: 'name', width: 200 },
  { title: '类别', key: 'category', width: 70, render: (r) => catMap[r.category] || r.category },
  {
    title: '状态',
    key: 'last_status',
    width: 90,
    render: (r) =>
      h(
        NTag,
        { size: 'small', type: r.last_status === 'success' ? 'success' : r.last_status === 'fail' ? 'error' : 'default', bordered: false },
        { default: () => (r.last_status === 'success' ? '正常' : r.last_status === 'fail' ? '失败' : '未运行') }
      ),
  },
  { title: '最近抓取', key: 'last_run_at', width: 150 },
  { title: '条数', key: 'last_item_count', width: 70 },
  {
    title: '错误信息',
    key: 'last_error',
    ellipsis: { tooltip: true },
    render: (r) => r.last_error || '—',
  },
  {
    title: '启用',
    key: 'enabled',
    width: 70,
    render: (r) =>
      h(NSwitch, { size: 'small', value: r.enabled === 1, 'onUpdate:value': (v) => toggleSource(r, v) }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (r) =>
      h(
        NButton,
        { size: 'tiny', secondary: true, type: 'primary', onClick: () => runSource(r) },
        { default: () => '立即抓取' }
      ),
  },
]

// ---------- 配置 ----------
const dailyReset = ref(true)
const retentionDays = ref(90)
const backupKeep = ref(14)

async function loadConfigs() {
  try {
    const list = await api.configList()
    const m = Object.fromEntries(list.map((c) => [c.key, c.value]))
    dailyReset.value = m['daily_reset.read_mark'] !== 'off'
    retentionDays.value = parseInt(m['retention.days'] || '90', 10)
    backupKeep.value = parseInt(m['backup.keep'] || '14', 10)
  } catch (e) {
    message.error(e.message)
  }
}

async function saveConfig(key, value) {
  try {
    await api.configUpdate(key, value)
    message.success('已保存')
  } catch (e) {
    message.error(e.message)
  }
}

async function saveRetentionAndBackup() {
  await saveConfig('retention.days', String(retentionDays.value))
  await saveConfig('backup.keep', String(backupKeep.value))
}

// ---------- 备份 ----------
const backups = ref([])

async function loadBackups() {
  try {
    backups.value = await api.backupList()
  } catch (e) {
    message.error(e.message)
  }
}

async function runBackup() {
  try {
    await api.backupRun()
    message.success('已触发备份，稍后刷新查看')
    setTimeout(loadBackups, 3000)
  } catch (e) {
    message.error(e.message)
  }
}

const backupColumns = [
  { title: '备份文件', key: 'name' },
  {
    title: '大小',
    key: 'size',
    width: 120,
    render: (r) => (r.size / 1024 / 1024).toFixed(2) + ' MB',
  },
]

onMounted(() => {
  loadSources()
  loadConfigs()
  loadBackups()
})
</script>

<style scoped>
.panel {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 12px 14px;
  margin-bottom: 10px;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.panel-head h3 {
  margin: 0;
  font-size: 13.5px;
}

.config-form {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: center;
}

.config-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.migrate-tip {
  margin-top: 12px;
  font-size: 12px;
  color: var(--text-sub);
  background: #f8fafc;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 8px 12px;
  line-height: 1.7;
}

.migrate-tip code {
  background: #eef2f7;
  padding: 1px 5px;
  border-radius: 3px;
  color: #334155;
}
</style>
