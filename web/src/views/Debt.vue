<template>
  <div>
    <h2 class="page-title"><Icon icon="fluent-emoji:clipboard" width="26" /> 每日债权检索</h2>

    <div class="filter-bar">
      <n-select
        v-model:value="sourceKey"
        :options="sourceOptions"
        placeholder="来源"
        clearable
        size="small"
        style="width: 180px"
        @update:value="reload"
      />
      <n-input
        v-model:value="region"
        placeholder="地区"
        clearable
        size="small"
        style="width: 130px"
        @keyup.enter="reload"
      />
      <n-select
        v-model:value="status"
        :options="statusOptions"
        placeholder="状态"
        clearable
        size="small"
        style="width: 120px"
        @update:value="reload"
      />
      <n-input
        v-model:value="keyword"
        placeholder="关键词搜索"
        clearable
        size="small"
        style="width: 190px"
        @keyup.enter="reload"
      />
      <n-button type="primary" secondary size="small" @click="reload">搜索</n-button>
    </div>

    <div v-for="d in list" :key="d.id" class="item-card" :class="{ unread: !d.is_read }">
      <div class="item-actions">
        <n-button text @click="toggleFav(d)">
          <Icon
            :icon="d.fav_id ? 'fluent-emoji:glowing-star' : 'fluent-emoji:star'"
            width="20"
          />
        </n-button>
      </div>
      <p class="item-title">
        <a :href="d.url" target="_blank" rel="noopener" @click="markRead(d)">{{ d.title }}</a>
      </p>
      <div class="item-meta">
        <n-tag size="tiny" type="success" :bordered="false">{{ sourceName(d.source_key) }}</n-tag>
        <span v-if="d.transferor">转让方：{{ d.transferor }}</span>
        <span v-if="d.amount_wan">金额：{{ d.amount_wan }} 万元</span>
        <span v-if="d.region">{{ d.region }}</span>
        <span v-if="d.end_time">截止：{{ d.end_time }}</span>
        <n-tag v-if="d.status" size="tiny" :bordered="false">{{ d.status }}</n-tag>
      </div>
    </div>

    <n-empty v-if="!loading && !list.length" description="暂无数据" style="padding: 48px 0" />

    <div style="display: flex; justify-content: center; margin-top: 16px">
      <n-pagination
        v-model:page="page"
        :page-size="pageSize"
        :item-count="total"
        @update:page="load"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useMessage } from 'naive-ui'
import { api } from '../api'

const message = useMessage()
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)

const sourceKey = ref(null)
const region = ref('')
const status = ref(null)
const keyword = ref('')

const sources = ref([])
const sourceOptions = ref([])
const statusOptions = [
  { label: '在售', value: '在售' },
  { label: '已结束', value: '已结束' },
]

const sourceName = (key) => sources.value.find((s) => s.source_key === key)?.name || key

async function load() {
  loading.value = true
  try {
    const data = await api.debtList({
      source_key: sourceKey.value,
      region: region.value,
      status: status.value,
      keyword: keyword.value,
      page: page.value,
      page_size: pageSize,
    })
    total.value = data.total
    list.value = data.list || []
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

function reload() {
  page.value = 1
  load()
}

async function markRead(d) {
  if (!d.is_read) {
    d.is_read = 1
    api.debtMarkRead([d.id]).catch(() => {})
  }
}

async function toggleFav(d) {
  try {
    if (d.fav_id) {
      await api.favoriteRemove(d.fav_id)
      d.fav_id = 0
      message.success('已取消收藏')
    } else {
      await api.favoriteAdd('debt', d.id, '')
      message.success('已收藏')
      load()
    }
  } catch (e) {
    message.error(e.message)
  }
}

onMounted(async () => {
  try {
    sources.value = (await api.sourceList()).filter((s) => s.category === 'debt')
    sourceOptions.value = sources.value.map((s) => ({
      label: s.name,
      value: s.source_key,
    }))
  } catch {
    /* 数据源列表加载失败不阻塞 */
  }
  load()
})
</script>
