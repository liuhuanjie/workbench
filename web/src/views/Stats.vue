<template>
  <div>
    <h2 class="page-title"><Icon icon="fluent-emoji:bar-chart" width="26" /> 数据统计</h2>

    <div class="stat-cards">
      <div class="stat-card">
        <Icon icon="fluent-emoji:star-struck" width="32" />
        <div>
          <div class="stat-num">{{ o?.favorite_total ?? '—' }}</div>
          <div class="stat-label">收藏总数</div>
        </div>
      </div>
      <div class="stat-card">
        <Icon icon="fluent-emoji:newspaper" width="32" />
        <div>
          <div class="stat-num">{{ o?.news_unread ?? '—' }}</div>
          <div class="stat-label">新闻未读</div>
        </div>
      </div>
      <div class="stat-card">
        <Icon icon="fluent-emoji:clipboard" width="32" />
        <div>
          <div class="stat-num">{{ o?.debt_unread ?? '—' }}</div>
          <div class="stat-label">债权未读</div>
        </div>
      </div>
      <div class="stat-card">
        <Icon icon="fluent-emoji:houses" width="32" />
        <div>
          <div class="stat-num">{{ o?.house_unread ?? '—' }}</div>
          <div class="stat-label">法拍未读</div>
        </div>
      </div>
    </div>

    <div class="chart-card">
      <h3>近 7 日新增趋势</h3>
      <svg viewBox="0 0 720 220" class="trend-svg">
        <!-- 网格线 -->
        <line v-for="gy in [40, 90, 140, 190]" :key="gy" :x1="40" :y1="gy" x2="700" :y2="gy"
          stroke="#e6fffa" stroke-width="1" />
        <!-- 折线 -->
        <polyline :points="linePoints" fill="none" stroke="#14b8a6" stroke-width="2.5"
          stroke-linejoin="round" stroke-linecap="round" />
        <!-- 数据点与标签 -->
        <g v-for="p in points" :key="p.label">
          <circle :cx="p.x" :cy="p.y" r="4" fill="#fff" stroke="#14b8a6" stroke-width="2" />
          <text :x="p.x" :y="p.y - 10" text-anchor="middle" font-size="12" fill="#0d9488">
            {{ p.v }}
          </text>
          <text :x="p.x" y="212" text-anchor="middle" font-size="11" fill="#64748b">
            {{ p.label }}
          </text>
        </g>
      </svg>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { api } from '../api'

const o = ref(null)

onMounted(async () => {
  try {
    o.value = await api.statsOverview()
  } catch {
    /* 忽略 */
  }
})

const points = computed(() => {
  const t = o.value?.trend || []
  if (!t.length) return []
  const max = Math.max(1, ...t.map((p) => p.news + p.debt + p.house))
  const w = 720,
    h = 220,
    padX = 40,
    padY = 40
  return t.map((p, i) => ({
    x: padX + (i * (w - 2 * padX)) / Math.max(1, t.length - 1),
    y: h - padY - ((p.news + p.debt + p.house) / max) * (h - 2 * padY),
    label: p.date,
    v: p.news + p.debt + p.house,
  }))
})

const linePoints = computed(() => points.value.map((p) => `${p.x},${p.y}`).join(' '))
</script>

<style scoped>
.stat-cards {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.stat-card {
  flex: 1;
  min-width: 160px;
  background: #fff;
  border-radius: 14px;
  padding: 20px 22px;
  display: flex;
  align-items: center;
  gap: 14px;
  box-shadow: 0 1px 3px rgba(19, 78, 74, 0.08);
}

.stat-num {
  font-size: 26px;
  font-weight: 700;
  color: #0d9488;
}

.stat-label {
  font-size: 13px;
  color: #64748b;
}

.chart-card {
  background: #fff;
  border-radius: 14px;
  padding: 20px 24px;
  margin-top: 16px;
  box-shadow: 0 1px 3px rgba(19, 78, 74, 0.08);
}

.chart-card h3 {
  margin: 0 0 12px;
  font-size: 15px;
}

.trend-svg {
  width: 100%;
  height: auto;
}
</style>
