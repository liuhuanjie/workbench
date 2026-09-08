<template>
  <div>
    <h2 class="page-title"><Icon icon="fluent-emoji:glowing-star" width="26" /> 我的收藏</h2>

    <div class="filter-bar">
      <n-radio-group v-model:value="type" size="small" @update:value="reload">
        <n-radio-button value="">全部</n-radio-button>
        <n-radio-button value="news">新闻</n-radio-button>
        <n-radio-button value="debt">债权</n-radio-button>
        <n-radio-button value="house">法拍</n-radio-button>
      </n-radio-group>
      <span style="font-size: 13px; color: #64748b">共 {{ total }} 条</span>
    </div>

    <div v-for="f in list" :key="f.id" class="item-card">
      <div class="item-actions">
        <n-popconfirm @positive-click="removeFav(f)">
          <template #trigger>
            <n-button text>
              <Icon icon="fluent-emoji:cross-mark" width="18" />
            </n-button>
          </template>
          确认取消收藏？
        </n-popconfirm>
      </div>
      <p class="item-title">
        <a :href="f.url" target="_blank" rel="noopener">{{ f.title }}</a>
      </p>
      <div class="item-meta">
        <n-tag size="tiny" :bordered="false" type="info">{{ typeName(f.item_type) }}</n-tag>
        <span v-if="f.extra">{{ f.extra }}</span>
        <span>收藏于 {{ f.created_at }}</span>
      </div>
      <p v-if="f.note" class="item-summary">备注：{{ f.note }}</p>
    </div>

    <n-empty v-if="!loading && !list.length" description="暂无收藏" style="padding: 48px 0" />

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
const type = ref('')

const typeMap = { news: '新闻', debt: '债权', house: '法拍' }
const typeName = (t) => typeMap[t] || t

async function load() {
  loading.value = true
  try {
    const data = await api.favoriteList({ item_type: type.value, page: page.value, page_size: pageSize })
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

async function removeFav(f) {
  try {
    await api.favoriteRemove(f.id)
    message.success('已取消收藏')
    load()
  } catch (e) {
    message.error(e.message)
  }
}

onMounted(load)
</script>
