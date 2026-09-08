<template>
  <div class="layout">
    <aside class="sider">
      <div class="logo">
        <Icon icon="fluent-emoji:house-with-garden" width="34" />
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
          <Icon :icon="m.icon" width="22" />
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
import { Icon } from '@iconify/vue'

const menus = [
  { path: '/', label: '工作台首页', icon: 'fluent-emoji:house-with-garden' },
  { path: '/news', label: '业务新闻', icon: 'fluent-emoji:newspaper' },
  { path: '/debt', label: '每日债权检索', icon: 'fluent-emoji:clipboard' },
  { path: '/house', label: '每日住宅检索', icon: 'fluent-emoji:houses' },
  { path: '/favorite', label: '我的收藏', icon: 'fluent-emoji:glowing-star' },
  { path: '/stats', label: '数据统计', icon: 'fluent-emoji:bar-chart' },
  { path: '/data', label: '数据管理', icon: 'fluent-emoji:card-file-box' },
  { path: '/settings', label: '个人设置', icon: 'fluent-emoji:gear' },
]
</script>
