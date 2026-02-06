<template>
  <div class="reset-container">
    <!-- Background Orbs -->
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>
    <div class="bg-orb orb-3"></div>

    <n-card class="reset-card">
      <div class="reset-header">
        <h1 class="reset-title">重置密码</h1>
        <p class="reset-subtitle">请输入您的新密码</p>
      </div>

      <!-- Invalid token -->
      <div v-if="!token" class="reset-content">
        <n-result status="error" title="链接无效" description="密码重置链接无效或缺少令牌">
          <template #footer>
            <n-button type="primary" @click="router.push('/forgot-password')">重新获取</n-button>
          </template>
        </n-result>
      </div>

      <!-- Success state -->
      <div v-else-if="success" class="reset-content">
        <n-result status="success" title="密码已重置" description="您的密码已成功重置，现在可以使用新密码登录了">
          <template #footer>
            <n-button type="primary" @click="router.push('/login')">前往登录</n-button>
          </template>
        </n-result>
      </div>

      <!-- Form -->
      <n-form v-else ref="formRef" :model="form">
        <n-form-item label="新密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="请输入新密码 (至少6个字符)"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="确认密码">
          <n-input
            v-model:value="form.confirmPassword"
            type="password"
            placeholder="再次输入新密码"
            show-password-on="click"
            @keyup.enter="handleSubmit"
          />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="handleSubmit" class="reset-btn">
          重置密码
        </n-button>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { resetPassword } from '../api'

const router = useRouter()
const route = useRoute()
const message = useMessage()

const loading = ref(false)
const success = ref(false)
const form = ref({
  password: '',
  confirmPassword: '',
})

const token = computed(() => route.query.token as string)

const handleSubmit = async () => {
  if (!form.value.password) {
    message.error('请输入新密码')
    return
  }
  if (form.value.password.length < 6) {
    message.error('密码至少6个字符')
    return
  }
  if (form.value.password !== form.value.confirmPassword) {
    message.error('两次输入的密码不一致')
    return
  }

  loading.value = true
  try {
    await resetPassword(token.value, form.value.password)
    success.value = true
  } catch (e: any) {
    message.error(e.response?.data?.error || '重置失败，链接可能已过期')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!token.value) {
    // Token missing
  }
})
</script>

<style scoped>
.reset-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0a0e27;
  position: relative;
  overflow: hidden;
}

.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.4;
  pointer-events: none;
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: #10b981;
  top: -200px;
  right: -100px;
  animation: float 8s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: #3b82f6;
  bottom: -150px;
  left: -100px;
  animation: float 10s ease-in-out infinite reverse;
}

.orb-3 {
  width: 300px;
  height: 300px;
  background: #8b5cf6;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: pulse 6s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-30px) rotate(5deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 0.2; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 0.4; transform: translate(-50%, -50%) scale(1.1); }
}

.reset-card {
  width: 400px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05) !important;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  z-index: 10;
}

.reset-header {
  text-align: center;
  margin-bottom: 32px;
}

.reset-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #10b981 0%, #3b82f6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 8px 0;
}

.reset-subtitle {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin: 0;
}

.reset-content {
  padding: 20px 0;
}

.reset-btn {
  margin-top: 8px;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 12px !important;
  background: linear-gradient(135deg, #10b981 0%, #3b82f6 100%) !important;
  border: none !important;
  box-shadow: 0 4px 15px rgba(16, 185, 129, 0.4);
  transition: all 0.3s ease;
}

.reset-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.5);
}

:deep(.n-form-item-label) {
  color: rgba(255, 255, 255, 0.7) !important;
}

:deep(.n-result) {
  background: transparent;
}

:deep(.n-result-header__title) {
  color: rgba(255, 255, 255, 0.9) !important;
}

:deep(.n-result-header__description) {
  color: rgba(255, 255, 255, 0.6) !important;
}
</style>
