<template>
  <div class="empty-state">
    <div class="empty-icon">
      <n-icon :size="64" :color="iconColor">
        <component :is="iconComponent" />
      </n-icon>
    </div>
    <div class="empty-title">{{ title }}</div>
    <div class="empty-desc" v-if="description">{{ description }}</div>
    <div class="empty-action" v-if="actionText">
      <n-button type="primary" @click="$emit('action')">
        <template #icon v-if="actionIcon">
          <n-icon><component :is="actionIcon" /></n-icon>
        </template>
        {{ actionText }}
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NIcon, NButton } from 'naive-ui'
import {
  ServerOutline,
  DesktopOutline,
  PeopleOutline,
  NotificationsOutline,
  SwapHorizontalOutline,
  GitNetworkOutline,
  LinkOutline,
  SearchOutline,
} from '@vicons/ionicons5'

const props = withDefaults(defineProps<{
  type?: 'nodes' | 'clients' | 'users' | 'notify' | 'forwards' | 'groups' | 'tunnels' | 'search' | 'default'
  title?: string
  description?: string
  actionText?: string
  actionIcon?: any
}>(), {
  type: 'default'
})

defineEmits(['action'])

const iconMap: Record<string, any> = {
  nodes: ServerOutline,
  clients: DesktopOutline,
  users: PeopleOutline,
  notify: NotificationsOutline,
  forwards: SwapHorizontalOutline,
  groups: GitNetworkOutline,
  tunnels: LinkOutline,
  search: SearchOutline,
  default: SearchOutline,
}

const titleMap: Record<string, string> = {
  nodes: '暂无节点',
  clients: '暂无客户端',
  users: '暂无用户',
  notify: '暂无通知渠道',
  forwards: '暂无端口转发规则',
  groups: '暂无负载均衡组',
  tunnels: '暂无隧道',
  search: '未找到匹配结果',
  default: '暂无数据',
}

const descMap: Record<string, string> = {
  nodes: '添加第一个节点来开始使用',
  clients: '添加客户端来配置内网穿透',
  users: '创建用户来分配访问权限',
  notify: '配置通知渠道接收告警消息',
  forwards: '创建端口转发规则',
  groups: '创建节点组实现负载均衡',
  tunnels: '创建隧道连接入口和出口节点',
  search: '尝试修改搜索关键词',
  default: '',
}

const iconComponent = computed(() => iconMap[props.type] || iconMap.default)
const iconColor = computed(() => 'rgba(255, 255, 255, 0.3)')

const title = computed(() => props.title || titleMap[props.type] || titleMap.default)
const description = computed(() => props.description ?? descMap[props.type])
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.empty-icon {
  margin-bottom: 16px;
  opacity: 0.6;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.85);
  margin-bottom: 8px;
}

.empty-desc {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 20px;
  max-width: 300px;
}

.empty-action {
  margin-top: 8px;
}
</style>
