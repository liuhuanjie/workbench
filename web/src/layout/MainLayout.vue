<template>
  <div class="layout">
    <aside class="sider">
      <div class="logo">
        <n-icon :component="CubeOutline" :size="18" />
        <span>资产工作台</span>
      </div>
      <nav>
        <router-link
          v-for="m in menus"
          :key="m.path"
          :to="m.path"
          class="menu-item"
          exact-active-class="active"
        >
          <n-icon :component="m.icon" :size="16" />
          <span>{{ m.label }}</span>
        </router-link>
      </nav>
      <div class="version">资产工作台 v1.0</div>
    </aside>
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import {
  CubeOutline,
  HomeOutline,
  NewspaperOutline,
  ReceiptOutline,
  BusinessOutline,
  BookmarkOutline,
  StatsChartOutline,
  ServerOutline,
  SettingsOutline,
} from '../icons'

const menus = [
  { path: '/', label: '工作台首页', icon: HomeOutline },
  { path: '/news', label: '业务新闻', icon: NewspaperOutline },
  { path: '/debt', label: '每日债权检索', icon: ReceiptOutline },
  { path: '/house', label: '每日住宅检索', icon: BusinessOutline },
  { path: '/favorite', label: '我的收藏', icon: BookmarkOutline },
  { path: '/stats', label: '数据统计', icon: StatsChartOutline },
  { path: '/data', label: '数据管理', icon: ServerOutline },
  { path: '/settings', label: '个人设置', icon: SettingsOutline },
]
</script>
