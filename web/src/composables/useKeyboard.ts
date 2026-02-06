import { onMounted, onUnmounted, type Ref } from 'vue'

interface KeyboardOptions {
  onNew?: () => void
  onEscape?: () => void
  onSave?: () => void
  modalVisible?: Ref<boolean>
}

export function useKeyboard(options: KeyboardOptions) {
  const handleKeydown = (e: KeyboardEvent) => {
    // Ctrl+N / Cmd+N: New
    if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
      e.preventDefault()
      if (options.onNew && !options.modalVisible?.value) {
        options.onNew()
      }
    }

    // Escape: Close modal
    if (e.key === 'Escape') {
      if (options.onEscape) {
        options.onEscape()
      } else if (options.modalVisible?.value) {
        options.modalVisible.value = false
      }
    }

    // Ctrl+S / Cmd+S: Save (when modal is open)
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      if (options.modalVisible?.value && options.onSave) {
        e.preventDefault()
        options.onSave()
      }
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
  })
}
