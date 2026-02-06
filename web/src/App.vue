<template>
  <n-config-provider :theme="currentTheme" :theme-overrides="currentThemeOverrides">
    <n-message-provider>
      <n-dialog-provider>
        <!-- Background Orbs for Glassmorphism effect (dark mode only) -->
        <div v-if="themeStore.isDark" class="bg-orb bg-orb-1"></div>
        <div v-if="themeStore.isDark" class="bg-orb bg-orb-2"></div>
        <router-view />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { darkTheme, type GlobalThemeOverrides } from 'naive-ui'
import { useThemeStore } from './stores/theme'

const themeStore = useThemeStore()

// 根据主题状态选择主题
const currentTheme = computed(() => themeStore.isDark ? darkTheme : null)

// 暗色主题配置 - Glassmorphism 风格
const darkThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#3b82f6',
    primaryColorHover: '#60a5fa',
    primaryColorPressed: '#2563eb',
    primaryColorSuppl: '#3b82f6',
    successColor: '#22c55e',
    successColorHover: '#4ade80',
    successColorPressed: '#16a34a',
    warningColor: '#f59e0b',
    warningColorHover: '#fbbf24',
    warningColorPressed: '#d97706',
    errorColor: '#ef4444',
    errorColorHover: '#f87171',
    errorColorPressed: '#dc2626',
    infoColor: '#06b6d4',
    textColorBase: '#ffffff',
    textColor1: '#ffffff',
    textColor2: 'rgba(255, 255, 255, 0.82)',
    textColor3: 'rgba(255, 255, 255, 0.52)',
    bodyColor: '#0a0e27',
    cardColor: 'rgba(255, 255, 255, 0.05)',
    modalColor: 'rgba(15, 21, 53, 0.95)',
    popoverColor: 'rgba(15, 21, 53, 0.95)',
    tableColor: 'transparent',
    inputColor: 'rgba(255, 255, 255, 0.05)',
    actionColor: 'rgba(255, 255, 255, 0.05)',
    hoverColor: 'rgba(255, 255, 255, 0.08)',
    tableColorHover: 'rgba(255, 255, 255, 0.05)',
    borderColor: 'rgba(255, 255, 255, 0.1)',
    dividerColor: 'rgba(255, 255, 255, 0.1)',
    borderRadius: '12px',
    borderRadiusSmall: '8px',
    fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
  },
  Card: {
    color: 'rgba(255, 255, 255, 0.05)',
    colorModal: 'rgba(15, 21, 53, 0.95)',
    borderColor: 'rgba(255, 255, 255, 0.1)',
    borderRadius: '16px',
    paddingMedium: '20px',
    titleFontSizeMedium: '16px',
    titleFontWeight: '600',
  },
  Button: {
    colorPrimary: '#3b82f6',
    colorHoverPrimary: '#60a5fa',
    colorPressedPrimary: '#2563eb',
    borderRadiusMedium: '10px',
    borderRadiusSmall: '8px',
    fontWeightStrong: '500',
  },
  Tag: {
    borderRadius: '8px',
  },
  DataTable: {
    thColor: 'rgba(255, 255, 255, 0.03)',
    tdColor: 'transparent',
    tdColorHover: 'rgba(255, 255, 255, 0.05)',
    borderRadius: '12px',
    borderColor: 'rgba(255, 255, 255, 0.08)',
  },
  Menu: {
    itemColorHover: 'rgba(255, 255, 255, 0.08)',
    itemColorActive: 'rgba(59, 130, 246, 0.15)',
    itemColorActiveHover: 'rgba(59, 130, 246, 0.2)',
    borderRadius: '8px',
    itemBorderRadius: '8px',
  },
  Input: {
    color: 'rgba(255, 255, 255, 0.05)',
    colorFocus: 'rgba(255, 255, 255, 0.08)',
    borderColor: 'rgba(255, 255, 255, 0.1)',
    borderRadius: '10px',
    placeholderColor: 'rgba(255, 255, 255, 0.3)',
  },
  Select: {
    peers: {
      InternalSelection: {
        color: 'rgba(255, 255, 255, 0.05)',
        colorActive: 'rgba(255, 255, 255, 0.08)',
      },
    },
  },
  Statistic: {
    valueFontSize: '28px',
    labelFontSize: '14px',
  },
  Switch: {
    railColorActive: '#3b82f6',
  },
  Modal: {
    color: 'rgba(15, 21, 53, 0.98)',
    borderRadius: '16px',
  },
  Dropdown: {
    color: 'rgba(15, 21, 53, 0.98)',
    borderRadius: '12px',
    optionColorHover: 'rgba(255, 255, 255, 0.08)',
  },
  Tooltip: {
    color: 'rgba(15, 21, 53, 0.98)',
    borderRadius: '8px',
  },
  Message: {
    borderRadius: '10px',
  },
  Notification: {
    borderRadius: '12px',
  },
}

// 亮色主题配置 - 柔和护眼风格
const lightThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#4f7cff',
    primaryColorHover: '#3d6ce8',
    primaryColorPressed: '#2d5cd4',
    primaryColorSuppl: '#4f7cff',
    successColor: '#34c759',
    successColorHover: '#28a745',
    successColorPressed: '#1e7e34',
    warningColor: '#ff9500',
    warningColorHover: '#e68600',
    warningColorPressed: '#cc7700',
    errorColor: '#ff3b30',
    errorColorHover: '#e6342b',
    errorColorPressed: '#cc2e26',
    infoColor: '#5ac8fa',
    // 柔和的文字颜色
    textColorBase: '#2c3e50',
    textColor1: '#2c3e50',
    textColor2: '#4a5568',
    textColor3: '#718096',
    // 柔和的背景色 - 米色调
    bodyColor: '#f8f6f1',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    popoverColor: '#ffffff',
    tableColor: '#ffffff',
    inputColor: '#faf9f7',
    actionColor: '#f5f3ef',
    hoverColor: '#f0ede8',
    tableColorHover: '#f5f3ef',
    // 柔和的边框
    borderColor: '#e8e4db',
    dividerColor: '#e8e4db',
    borderRadius: '12px',
    borderRadiusSmall: '8px',
    fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
  },
  Card: {
    color: '#ffffff',
    borderColor: '#e8e4db',
    borderRadius: '16px',
    paddingMedium: '20px',
    titleFontSizeMedium: '16px',
    titleFontWeight: '600',
    titleTextColor: '#2c3e50',
  },
  Button: {
    colorPrimary: '#4f7cff',
    colorHoverPrimary: '#3d6ce8',
    colorPressedPrimary: '#2d5cd4',
    borderRadiusMedium: '10px',
    borderRadiusSmall: '8px',
    fontWeightStrong: '500',
  },
  Tag: {
    borderRadius: '8px',
  },
  DataTable: {
    thColor: '#faf9f7',
    tdColor: '#ffffff',
    tdColorHover: '#f5f3ef',
    borderRadius: '12px',
    borderColor: '#e8e4db',
    thTextColor: '#4a5568',
    tdTextColor: '#2c3e50',
  },
  Menu: {
    itemColorHover: '#f0ede8',
    itemColorActive: 'rgba(79, 124, 255, 0.12)',
    itemColorActiveHover: 'rgba(79, 124, 255, 0.18)',
    itemTextColor: '#4a5568',
    itemTextColorActive: '#4f7cff',
    itemTextColorHover: '#2c3e50',
    borderRadius: '8px',
    itemBorderRadius: '8px',
  },
  Input: {
    color: '#faf9f7',
    colorFocus: '#ffffff',
    borderColor: '#e8e4db',
    borderColorFocus: '#4f7cff',
    borderRadius: '10px',
    placeholderColor: '#a0aec0',
    textColor: '#2c3e50',
  },
  Select: {
    peers: {
      InternalSelection: {
        color: '#faf9f7',
        colorActive: '#ffffff',
        textColor: '#2c3e50',
        placeholderColor: '#a0aec0',
        border: '1px solid #e8e4db',
      },
    },
  },
  Statistic: {
    valueFontSize: '28px',
    labelFontSize: '14px',
    labelTextColor: '#718096',
  },
  Switch: {
    railColorActive: '#4f7cff',
  },
  Modal: {
    color: '#ffffff',
    borderRadius: '16px',
    titleTextColor: '#2c3e50',
  },
  Dropdown: {
    color: '#ffffff',
    borderRadius: '12px',
    optionColorHover: '#f0ede8',
    optionTextColor: '#2c3e50',
  },
  Tooltip: {
    color: '#2c3e50',
    textColor: '#ffffff',
    borderRadius: '8px',
  },
  Message: {
    borderRadius: '10px',
  },
  Notification: {
    borderRadius: '12px',
  },
  Pagination: {
    itemColorHover: '#f0ede8',
    itemColorActive: 'rgba(79, 124, 255, 0.12)',
    itemTextColor: '#4a5568',
    itemTextColorActive: '#4f7cff',
  },
}

// 根据主题状态选择配置
const currentThemeOverrides = computed(() =>
  themeStore.isDark ? darkThemeOverrides : lightThemeOverrides
)
</script>

<style>
/* 根据主题切换背景 */
html, body {
  transition: background-color 0.3s ease;
}

html.dark, html.dark body {
  background: #0a0e27 !important;
}

html:not(.dark), html:not(.dark) body {
  background: #f8f6f1 !important;
}
</style>
