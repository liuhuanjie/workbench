<template>
  <div>
    <h2 class="page-title"><Icon icon="fluent-emoji:newspaper" width="26" /> 业务新闻阅读</h2>

    <div class="filter-bar">
      <n-radio-group v-model:value="category" size="small" @update:value="reload">
        <n-radio-button value="">全部</n-radio-button>
        <n-radio-button value="central_amc">5大AMC</n-radio-button>
        <n-radio-button value="local_amc">地方AMC</n-radio-button>
        <n-radio-button value="bank_transfer">银行资产包</n-radio-button>
        <n-radio-button value="industry">行业要闻</n-radio-button>
      </n-radio-group>
      <n-input
        v-model:value="keyword"
        placeholder="关键词搜索"
        clearable
        size="small"
        style="width: 200px"
        @keyup.enter="reload"
      />
      <n-checkbox v-model:checked="unreadOnly" @update:checked="reload">只看未读</n-checkbox>
      <n-button type="primary" secondary size="small" @click="reload">搜索</n-button>
    </div>

    <div v-for="n in list" :key="n.id" class="item-card" :class="{ unread: !n.is_read }">
      <div class="item-actions">
        <n-button text @click="toggleFav(n)">
          <Icon
            :icon="n.fav_id ? 'fluent-emoji:glowing-star' : 'fluent-emoji:star'"
            width="20"
          />
        </n-button>
      </div>
      <p class="item-title">
        <a :href="n.url" target="_blank" rel="noopener" @click="markRead(n)">{{ n.title }}</a>
      </p>
      <div class="item-meta">
        <n-tag size="tiny" type="info" :bordered="false">{{ catName(n.category) }}</n-tag>
        <span v-if="n.org_name">{{ n.org_name }}</span>
        <span v-if="n.published_at">{{ n.published_at }}</span>
        <span v-if="!n.is_read" style="color: #14b8a6">未读</span>
      </div>
      <p v-if="n.summary" class="item-summary">{{ n.summary }}</p>
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

const category = ref('')
const keyword = ref('')
const unreadOnly = ref(false)

const catMap = {
  central_amc: '5大AMC',
  local_amc: '地方AMC',
  bank_transfer: '银行资产包',
  industry: '行业要闻',
}
const catName = (c) => catMap[c] || c

async function load() {
  loading.value = true
  try {
    const data = await api.newsList({
      category: category.value,
      keyword: keyword.value,
      unread: unreadOnly.value,
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

async function markRead(n) {
  if (!n.is_read) {
    n.is_read = 1
    api.newsMarkRead([n.id]).catch(() => {})
  }
}

async function toggleFav(n) {
  try {
    if (n.fav_id) {
      await api.favoriteRemove(n.fav_id)
      n.fav_id = 0
      message.success('已取消收藏')
    } else {
      await api.favoriteAdd('news', n.id, '')
      message.success('已收藏')
      load()
    }
  } catch (e) {
    message.error(e.message)
  }
}

onMounted(load)
</script>
