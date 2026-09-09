<template>
  <div>
    <!-- 时钟卡片 -->
    <div class="clock-card">
      <div class="clock">{{ time }}</div>
      <div class="date-line">{{ dateLine }}</div>
    </div>

    <!-- 每日激励语 -->
    <div class="quote-card">
      <div class="quote-mark"><n-icon :component="BulbOutline" :size="20" /></div>
      <div>
        <div class="quote-zh">{{ quote.zh }}</div>
        <div class="quote-en">{{ quote.en }}</div>
      </div>
    </div>

    <!-- 今日速览 -->
    <div class="stat-cards">
      <div v-for="s in todayCards" :key="s.label" class="stat-card">
        <n-icon :component="s.icon" :size="22" />
        <div>
          <div class="stat-num">{{ s.value }}</div>
          <div class="stat-label">{{ s.label }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api } from '../api'
import { todayQuote } from '../data/quotes'
import { BulbOutline, NewspaperOutline, ReceiptOutline, BusinessOutline } from '../icons'

const time = ref('')
const dateLine = ref('')
const quote = todayQuote()
const overview = ref(null)
let timer = null

function tick() {
  const d = new Date()
  time.value = d.toTimeString().slice(0, 8)
  const week = ['日', '一', '二', '三', '四', '五', '六'][d.getDay()]
  dateLine.value = `${d.getFullYear()} 年 ${d.getMonth() + 1} 月 ${d.getDate()} 日 · 星期${week}`
}

onMounted(() => {
  tick()
  timer = setInterval(tick, 1000)
  api.statsOverview().then((o) => (overview.value = o)).catch(() => {})
})

onUnmounted(() => clearInterval(timer))

const todayCards = computed(() => [
  { icon: NewspaperOutline, label: '今日新增新闻', value: overview.value?.news_today ?? '—' },
  { icon: ReceiptOutline, label: '今日新增债权', value: overview.value?.debt_today ?? '—' },
  { icon: BusinessOutline, label: '今日新增法拍', value: overview.value?.house_today ?? '—' },
])
</script>

<style scoped>
.clock-card {
  background: linear-gradient(135deg, #1e3a8a 0%, #1d4ed8 100%);
  border-radius: 4px;
  padding: 20px 24px;
  color: #fff;
}

.clock {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 40px;
  font-weight: 600;
  letter-spacing: 3px;
}

.date-line {
  margin-top: 4px;
  font-size: 13px;
  opacity: 0.9;
}

.quote-card {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 14px 16px;
  margin-top: 10px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  color: var(--primary);
}

.quote-mark {
  flex-shrink: 0;
  margin-top: 2px;
}

.quote-zh {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
}

.quote-en {
  margin-top: 4px;
  font-size: 12.5px;
  font-style: italic;
  color: var(--text-sub);
}

.stat-cards {
  display: flex;
  gap: 10px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.stat-card {
  flex: 1;
  min-width: 180px;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--primary);
}

.stat-num {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-main);
}

.stat-label {
  font-size: 12px;
  color: var(--text-sub);
}
</style>
