import { ref } from 'vue'

const notificationPermission = ref<NotificationPermission>('default')
const notifiedNodes = new Set<number>()
const NOTIFICATION_COOLDOWN = 5 * 60 * 1000 // 5 minutes

// Get favicon from document or use default
const getNotificationIcon = (): string => {
  const favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement
  return favicon?.href || '/vite.svg'
}

export function useBrowserNotification() {
  const requestPermission = async () => {
    if (!('Notification' in window)) {
      return false
    }

    if (Notification.permission === 'granted') {
      notificationPermission.value = 'granted'
      return true
    }

    if (Notification.permission !== 'denied') {
      const permission = await Notification.requestPermission()
      notificationPermission.value = permission
      return permission === 'granted'
    }

    return false
  }

  const showNotification = (title: string, options?: NotificationOptions) => {
    if (notificationPermission.value !== 'granted') return null

    const icon = getNotificationIcon()
    const notification = new Notification(title, {
      icon,
      badge: icon,
      ...options,
    })

    notification.onclick = () => {
      window.focus()
      notification.close()
    }

    return notification
  }

  const notifyNodeOffline = (nodeId: number, nodeName: string) => {
    if (notifiedNodes.has(nodeId)) return

    notifiedNodes.add(nodeId)
    showNotification('节点离线告警', {
      body: `节点 "${nodeName}" 已离线`,
      tag: `node-offline-${nodeId}`,
      requireInteraction: true,
    })

    // Allow re-notification after cooldown
    setTimeout(() => {
      notifiedNodes.delete(nodeId)
    }, NOTIFICATION_COOLDOWN)
  }

  const notifyNodeOnline = (nodeId: number, nodeName: string) => {
    notifiedNodes.delete(nodeId)
    showNotification('节点恢复上线', {
      body: `节点 "${nodeName}" 已恢复上线`,
      tag: `node-online-${nodeId}`,
    })
  }

  const checkPermission = () => {
    if ('Notification' in window) {
      notificationPermission.value = Notification.permission
    }
    return notificationPermission.value
  }

  return {
    notificationPermission,
    requestPermission,
    showNotification,
    notifyNodeOffline,
    notifyNodeOnline,
    checkPermission,
  }
}
