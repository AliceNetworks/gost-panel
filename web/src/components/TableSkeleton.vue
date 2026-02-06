<template>
  <div class="skeleton-table">
    <!-- Header skeleton -->
    <div class="skeleton-header">
      <div class="skeleton-cell skeleton-checkbox"></div>
      <div v-for="col in columns" :key="col" class="skeleton-cell" :style="{ flex: col }"></div>
    </div>
    <!-- Row skeletons -->
    <div v-for="row in rows" :key="row" class="skeleton-row">
      <div class="skeleton-cell skeleton-checkbox"></div>
      <div v-for="col in columns" :key="col" class="skeleton-cell" :style="{ flex: col }">
        <div class="skeleton-content" :style="{ width: getRandomWidth() }"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  rows?: number
  columns?: number[]
}>(), {
  rows: 5,
  columns: () => [1, 2, 1, 1, 1, 2]
})

const getRandomWidth = () => {
  const widths = ['60%', '70%', '80%', '90%', '50%']
  return widths[Math.floor(Math.random() * widths.length)]
}
</script>

<style scoped>
.skeleton-table {
  width: 100%;
}

.skeleton-header,
.skeleton-row {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 16px;
}

.skeleton-header {
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.skeleton-row {
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.skeleton-row:last-child {
  border-bottom: none;
}

.skeleton-cell {
  min-height: 20px;
}

.skeleton-checkbox {
  width: 20px;
  height: 20px;
  flex: none;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 4px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-content {
  height: 16px;
  background: linear-gradient(90deg, rgba(255,255,255,0.08) 25%, rgba(255,255,255,0.15) 50%, rgba(255,255,255,0.08) 75%);
  background-size: 200% 100%;
  border-radius: 4px;
  animation: skeleton-shimmer 1.5s ease-in-out infinite;
}

@keyframes skeleton-pulse {
  0%, 100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
}

@keyframes skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
