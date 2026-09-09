<template>
  <div>
    <h2 class="page-title">
      <n-icon :component="BusinessOutline" :size="18" />
      每日住宅检索
    </h2>

    <div class="filter-bar">
      <n-select
        v-model:value="city"
        :options="cityOptions"
        placeholder="城市"
        clearable
        size="small"
        style="width: 110px"
        @update:value="reload"
      />
      <n-input
        v-model:value="district"
        placeholder="行政区"
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
      <n-date-picker
        v-model:formatted-value="dateFrom"
        type="date"
        clearable
        size="small"
        style="width: 132px"
        value-format="yyyy-MM-dd"
        placeholder="开拍起"
      />
      <n-date-picker
        v-model:formatted-value="dateTo"
        type="date"
        clearable
        size="small"
        style="width: 132px"
        value-format="yyyy-MM-dd"
        placeholder="开拍止"
      />
      <n-input
        v-model:value="keyword"
        placeholder="关键词搜索"
        clearable
        size="small"
        style="width: 160px"
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
import { BusinessOutline } from '../icons'

const message = useMessage()
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)

const city = ref(null)
const district = ref('')
const status = ref(null)
const dateFrom = ref(null)
const dateTo = ref(null)
const keyword = ref('')

const cityOptions = [
  { label: '上海', value: '上海' },
  { label: '北京', value: '北京' },
]
const statusOptions = [
  { label: '待拍', value: '待拍' },
  { label: '拍卖中', value: '拍卖中' },
  { label: '已成交', value: '已成交' },
  { label: '流拍', value: '流拍' },
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
        {
          text: true,
          onClick: () => toggleFav(row),
        },
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
          row.court ? h('span', { style: 'margin-left:8px;color:#9CA3AF' }, row.court) : null,
        ]),
      ]),
  },
  {
    title: '地区',
    key: 'region',
    width: 110,
    render: (row) => [row.city, row.district].filter(Boolean).join(' · ') || '—',
  },
  {
    title: '起拍价(万)',
    key: 'start_price_wan',
    width: 100,
    align: 'right',
    render: (row) => price(row.start_price_wan),
  },
  {
    title: '评估价(万)',
    key: 'eval_price_wan',
    width: 100,
    align: 'right',
    render: (row) => price(row.eval_price_wan),
  },
  { title: '开拍时间', key: 'auction_date', width: 130, render: (row) => row.auction_date || '—' },
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
    const data = await api.houseList({
      city: city.value,
      district: district.value,
      status: status.value,
      date_from: dateFrom.value ? String(dateFrom.value).slice(0, 10) : '',
      date_to: dateTo.value ? String(dateTo.value).slice(0, 10) : '',
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

async function markRead(h) {
  if (!h.is_read) {
    h.is_read = 1
    api.houseMarkRead([h.id]).catch(() => {})
  }
}

async function toggleFav(row) {
  try {
    if (row.fav_id) {
      await api.favoriteRemove(row.fav_id)
      row.fav_id = 0
      message.success('已取消收藏')
    } else {
      await api.favoriteAdd('house', row.id, '')
      message.success('已收藏')
      load()
    }
  } catch (e) {
    message.error(e.message)
  }
}

onMounted(load)
</script>

<style>
.row-unread td {
  font-weight: 600;
}
</style>
