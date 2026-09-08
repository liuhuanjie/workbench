<template>
  <div>
    <h2 class="page-title"><Icon icon="fluent-emoji:houses" width="26" /> 每日住宅检索</h2>

    <div class="filter-bar">
      <n-select
        v-model:value="city"
        :options="cityOptions"
        placeholder="城市"
        clearable
        size="small"
        style="width: 120px"
        @update:value="reload"
      />
      <n-input
        v-model:value="district"
        placeholder="行政区"
        clearable
        size="small"
        style="width: 120px"
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
      <n-date-picker
        v-model:formatted-value="dateFrom"
        type="date"
        clearable
        size="small"
        style="width: 140px"
        value-format="yyyy-MM-dd"
        placeholder="开拍起"
      />
      <n-date-picker
        v-model:formatted-value="dateTo"
        type="date"
        clearable
        size="small"
        style="width: 140px"
        value-format="yyyy-MM-dd"
        placeholder="开拍止"
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

    <div v-for="h in list" :key="h.id" class="item-card" :class="{ unread: !h.is_read }">
      <div class="item-actions">
        <n-button text @click="toggleFav(h)">
          <Icon
            :icon="h.fav_id ? 'fluent-emoji:glowing-star' : 'fluent-emoji:star'"
            width="20"
          />
        </n-button>
      </div>
      <p class="item-title">
        <a :href="h.url" target="_blank" rel="noopener" @click="markRead(h)">{{ h.title }}</a>
      </p>
      <div class="item-meta">
        <n-tag size="tiny" type="info" :bordered="false">{{ h.city }}</n-tag>
        <span v-if="h.district">{{ h.district }}</span>
        <span v-if="h.area_sqm">{{ h.area_sqm }} ㎡</span>
        <span v-if="h.start_price_wan">起拍：{{ h.start_price_wan }} 万元</span>
        <span v-if="h.eval_price_wan">评估：{{ h.eval_price_wan }} 万元</span>
        <span v-if="h.court">{{ h.court }}</span>
        <span v-if="h.auction_date">开拍：{{ h.auction_date }}</span>
        <n-tag v-if="h.status" size="tiny" :bordered="false" type="warning">{{ h.status }}</n-tag>
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

async function toggleFav(h) {
  try {
    if (h.fav_id) {
      await api.favoriteRemove(h.fav_id)
      h.fav_id = 0
      message.success('已取消收藏')
    } else {
      await api.favoriteAdd('house', h.id, '')
      message.success('已收藏')
      load()
    }
  } catch (e) {
    message.error(e.message)
  }
}

onMounted(load)
</script>
