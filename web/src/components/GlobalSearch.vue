<template>
  <div>
    <!-- Search Trigger Button -->
    <n-button quaternary @click="showModal = true">
      <template #icon>
        <n-icon><search-outline /></n-icon>
      </template>
      <span class="search-shortcut">Ctrl+K</span>
    </n-button>

    <!-- Search Modal -->
    <n-modal v-model:show="showModal" preset="card" style="width: 600px; max-width: 90vw" :bordered="false">
      <template #header>
        <n-input
          ref="searchInputRef"
          v-model:value="searchQuery"
          placeholder="搜索节点、客户端、用户..."
          size="large"
          clearable
          @input="handleSearch"
        >
          <template #prefix>
            <n-icon><search-outline /></n-icon>
          </template>
        </n-input>
      </template>

      <div class="search-results">
        <n-spin :show="loading">
          <div v-if="results.length === 0 && searchQuery && !loading" class="empty-state">
            <n-empty description="没有找到匹配的结果" />
          </div>
          <div v-else-if="results.length === 0 && !searchQuery" class="empty-state">
            <n-text depth="3">输入关键字开始搜索</n-text>
          </div>
          <n-list v-else hoverable clickable>
            <n-list-item
              v-for="item in results"
              :key="`${item.type}-${item.id}`"
              @click="handleSelect(item)"
            >
              <template #prefix>
                <n-icon :color="getTypeColor(item.type)" size="20">
                  <component :is="getTypeIcon(item.type)" />
                </n-icon>
              </template>
              <n-thing :title="item.name" :description="item.desc">
                <template #header-extra>
                  <n-tag :type="getTagType(item.type)" size="small">
                    {{ getTypeLabel(item.type) }}
                  </n-tag>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-spin>
      </div>

      <template #footer>
        <n-text depth="3" style="font-size: 12px">
          按 Enter 选择 · 按 Esc 关闭
        </n-text>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { SearchOutline, ServerOutline, DesktopOutline, PersonOutline } from '@vicons/ionicons5'
import { globalSearch } from '../api'

interface SearchResult {
  type: 'node' | 'client' | 'user'
  id: number
  name: string
  desc: string
}

const router = useRouter()
const showModal = ref(false)
const searchQuery = ref('')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const searchInputRef = ref<any>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const handleSearch = () => {
  if (debounceTimer) clearTimeout(debounceTimer)

  if (!searchQuery.value.trim()) {
    results.value = []
    return
  }

  debounceTimer = setTimeout(async () => {
    loading.value = true
    try {
      const data: any = await globalSearch(searchQuery.value)
      results.value = data as SearchResult[]
    } catch (e) {
      console.error('Search failed', e)
      results.value = []
    } finally {
      loading.value = false
    }
  }, 300)
}

const handleSelect = (item: SearchResult) => {
  showModal.value = false
  searchQuery.value = ''
  results.value = []

  switch (item.type) {
    case 'node':
      router.push({ name: 'nodes', query: { highlight: item.id } })
      break
    case 'client':
      router.push({ name: 'clients', query: { highlight: item.id } })
      break
    case 'user':
      router.push({ name: 'users', query: { highlight: item.id } })
      break
  }
}

const getTypeIcon = (type: string) => {
  switch (type) {
    case 'node': return ServerOutline
    case 'client': return DesktopOutline
    case 'user': return PersonOutline
    default: return ServerOutline
  }
}

const getTypeLabel = (type: string) => {
  switch (type) {
    case 'node': return '节点'
    case 'client': return '客户端'
    case 'user': return '用户'
    default: return type
  }
}

const getTypeColor = (type: string) => {
  switch (type) {
    case 'node': return '#3b82f6'
    case 'client': return '#22c55e'
    case 'user': return '#f59e0b'
    default: return '#888'
  }
}

const getTagType = (type: string): 'info' | 'success' | 'warning' => {
  switch (type) {
    case 'node': return 'info'
    case 'client': return 'success'
    case 'user': return 'warning'
    default: return 'info'
  }
}

// Keyboard shortcut
const handleKeydown = (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    showModal.value = true
  }
  if (e.key === 'Escape' && showModal.value) {
    showModal.value = false
  }
}

watch(showModal, (val) => {
  if (val) {
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  } else {
    searchQuery.value = ''
    results.value = []
  }
})

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.search-shortcut {
  margin-left: 8px;
  padding: 2px 6px;
  font-size: 11px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  color: rgba(255, 255, 255, 0.5);
}

.search-results {
  max-height: 400px;
  overflow-y: auto;
  margin: -12px -24px;
  padding: 12px 24px;
}

.empty-state {
  padding: 40px 0;
  text-align: center;
}

:deep(.n-list-item) {
  border-radius: 8px;
  margin-bottom: 4px;
  transition: background 0.2s;
}

:deep(.n-list-item:hover) {
  background: rgba(255, 255, 255, 0.05);
}

:deep(.n-card-header) {
  padding-bottom: 0;
}

:deep(.n-card__content) {
  padding-top: 12px;
}
</style>
