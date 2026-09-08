<template>
  <div>
    <!-- 时钟卡片 -->
    <div class="clock-card">
      <div class="clock">{{ time }}</div>
      <div class="date-line">{{ dateLine }}</div>
    </div>

    <!-- 每日激励语 -->
    <div class="quote-card">
      <div class="quote-mark"><Icon icon="fluent-emoji:sparkles" width="26" /></div>
      <div class="quote-zh">{{ quote.zh }}</div>
      <div class="quote-en">{{ quote.en }}</div>
    </div>

    <!-- 今日速览 -->
    <div class="stat-cards">
      <div v-for="s in todayCards" :key="s.label" class="stat-card">
        <Icon :icon="s.icon" width="32" />
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
import { Icon } from '@iconify/vue'
import { api } from '../api'
import { todayQuote } from '../data/quotes'

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
  { icon: 'fluent-emoji:newspaper', label: '今日新增新闻', value: overview.value?.news_today ?? '—' },
  { icon: 'fluent-emoji:clipboard', label: '今日新增债权', value: overview.value?.debt_today ?? '—' },
  { icon: 'fluent-emoji:houses', label: '今日新增法拍', value: overview.value?.house_today ?? '—' },
])
</script>

<style scoped>
.clock-card {
  background: linear-gradient(135deg, #14b8a6 0%, #0ea5a4 60%, #38bdf8 100%);
  border-radius: 16px;
  padding: 36px 40px;
  color: #fff;
  box-shadow: 0 6px 24px rgba(20, 184, 166, 0.25);
}

.clock {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 56px;
  font-weight: 700;
  letter-spacing: 4px;
}

.date-line {
  margin-top: 8px;
  font-size: 15px;
  opacity: 0.92;
}

.quote-card {
  background: #fff;
  border-radius: 16px;
  padding: 26px 32px;
  margin-top: 16px;
  box-shadow: 0 1px 3px rgba(19, 78, 74, 0.08);
  display: flex;
  align-items: center;
  gap: 18px;
}

.quote-mark {
  flex-shrink: 0;
}

.quote-zh {
  font-size: 19px;
  font-weight: 600;
  color: #134e4a;
}

.quote-en {
  margin-top: 6px;
  font-size: 13.5px;
  font-style: italic;
  color: #64748b;
}

.stat-cards {
  display: flex;
  gap: 16px;
  margin-top: 16px;
  flex-wrap: wrap;
}

.stat-card {
  flex: 1;
  min-width: 200px;
  background: #fff;
  border-radius: 14px;
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(19, 78, 74, 0.08);
}

.stat-num {
  font-size: 28px;
  font-weight: 700;
  color: #0d9488;
}

.stat-label {
  font-size: 13px;
  color: #64748b;
}
</style>
