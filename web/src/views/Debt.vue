<template>
  <div>
    <h2 class="page-title">
      <n-icon :component="ReceiptOutline" :size="18" />
      每日债权检索
    </h2>

    <div class="filter-bar">
      <n-select
        v-model:value="sourceKey"
        :options="sourceOptions"
        placeholder="来源"
        clearable
        size="small"
        style="width: 170px"
        @update:value="reload"
      />
      <n-input
        v-model:value="region"
        placeholder="地区"
        clearable
        size="small"
        style="width: 110px"
        @keyup.enter="reload"
      />
      <n-select
        v-model:value="status"
        :options="statusOptions"
        placeholder="状态"
        clearable
        size="small"
        style="width: 110px"
        @update:value="reload"
      />
      <n-input
        v-model:value="keyword"
        placeholder="关键词搜索"
        clearable
        size="small"
        style="width: 170px"
        @keyup.enter="reload"
      />
      <n-button type="primary" secondary size="small" @click="reload">搜索</n-button>
    </div>

    <n-data-table
      :columns="columns"
      :data="list"
      :loading="loading"
      :bordered="false"
      size="small"
      :row-class-name="rowClass"
    />

    <div style="display: flex; justify-content: flex-end; margin-top: 10px">
      <n-pagination
        v-model:page="page"
        :page-size="pageSize"
        :item-count="total"
        size="small"
        @update:page="load"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NButton, NTag } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { api } from '../api'
import { sourceLabel } from '../sources'
import { ReceiptOutline } from '../icons'

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

const sourceOptions = ref([])
const statusOptions = [
  { label: '在售', value: '在售' },
  { label: '已结束', value: '已结束' },
]

function rowClass(row) {
  return row.is_read ? '' : 'row-unread'
}

function price(v) {
  return v ? v.toLocaleString('zh-CN', { maximumFractionDigits: 1 }) : '—'
}

const columns = [
  {
    title: '',
    key: 'fav',
    width: 40,
    render: (row) =>
      h(
        NButton,
        { text: true, onClick: () => toggleFav(row) },
        { default: () => h('span', { style: 'font-size:14px' }, row.fav_id ? '★' : '☆') }
      ),
  },
  {
    title: '标的',
    key: 'title',
    minWidth: 320,
    render: (row) =>
      h('div', null, [
        h(
          'a',
          {
            href: row.url,
            target: '_blank',
            rel: 'noopener',
            style: 'color:#1F2937;text-decoration:none',
            onClick: () => markRead(row),
          },
          row.title
        ),
        h('div', { style: 'margin-top:2px' }, [
          h('span', { class: 'source-tag' }, sourceLabel(row.source_key)),
        ]),
      ]),
  },
  {
    title: '转让方',
    key: 'transferor',
    width: 150,
    render: (row) => row.transferor || '—',
  },
  {
    title: '金额(万)',
    key: 'amount_wan',
    width: 100,
    align: 'right',
    render: (row) => price(row.amount_wan),
  },
  { title: '地区', key: 'region', width: 90, render: (row) => row.region || '—' },
  { title: '截止', key: 'end_time', width: 130, render: (row) => row.end_time || '—' },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row) =>
      row.status
        ? h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => row.status })
        : '—',
  },
]

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

async function markRead(row) {
  if (!row.is_read) {
    row.is_read = 1
    api.debtMarkRead([row.id]).catch(() => {})
  }
}

async function toggleFav(row) {
  try {
    if (row.fav_id) {
      await api.favoriteRemove(row.fav_id)
      row.fav_id = 0
      message.success('已取消收藏')
    } else {
      await api.favoriteAdd('debt', row.id, '')
      message.success('已收藏')
      load()
    }
  } catch (e) {
    message.error(e.message)
  }
}

onMounted(async () => {
  try {
    const sources = await api.sourceList()
    sourceOptions.value = sources
      .filter((s) => s.category === 'debt')
      .map((s) => ({ label: s.name, value: s.source_key }))
  } catch {
    /* 数据源列表加载失败不阻塞 */
  }
  load()
})
</script>

<style>
.row-unread td {
  font-weight: 600;
}
</style>
