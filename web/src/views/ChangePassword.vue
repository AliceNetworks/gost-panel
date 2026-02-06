<template>
  <div class="change-password-container">
    <n-card class="change-password-card">
      <template #header>
        <div class="card-header">
          <n-icon size="24" color="#f0a020">
            <LockClosedOutline />
          </n-icon>
          <span>{{ isForced ? '首次登录 - 修改默认密码' : '修改密码' }}</span>
        </div>
      </template>

      <n-alert v-if="isForced" type="warning" style="margin-bottom: 20px">
        检测到您正在使用默认密码，为了账户安全，请立即修改密码。
      </n-alert>

      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="100">
        <n-form-item label="当前密码" path="oldPassword">
          <n-input
            v-model:value="form.oldPassword"
            type="password"
            placeholder="请输入当前密码"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="新密码" path="newPassword">
          <n-input
            v-model:value="form.newPassword"
            type="password"
            placeholder="至少8位，含大小写字母和数字"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="确认密码" path="confirmPassword">
          <n-input
            v-model:value="form.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password-on="click"
            @keyup.enter="handleSubmit"
          />
        </n-form-item>
        <n-form-item>
          <n-space>
            <n-button type="primary" :loading="loading" @click="handleSubmit">
              确认修改
            </n-button>
            <n-button v-if="!isForced" @click="router.back()">
              取消
            </n-button>
          </n-space>
        </n-form-item>
      </n-form>

      <n-divider />
      <div class="password-tips">
        <h4>密码要求：</h4>
        <ul>
          <li :class="{ valid: hasMinLength }">至少 8 个字符</li>
          <li :class="{ valid: hasUppercase }">包含大写字母 (A-Z)</li>
          <li :class="{ valid: hasLowercase }">包含小写字母 (a-z)</li>
          <li :class="{ valid: hasDigit }">包含数字 (0-9)</li>
        </ul>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { LockClosedOutline } from '@vicons/ionicons5'
import { changePassword } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const userStore = useUserStore()

const isForced = computed(() => route.query.force === '1')
const loading = ref(false)
const formRef = ref()

const form = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// 密码强度检测
const hasMinLength = computed(() => form.value.newPassword.length >= 8)
const hasUppercase = computed(() => /[A-Z]/.test(form.value.newPassword))
const hasLowercase = computed(() => /[a-z]/.test(form.value.newPassword))
const hasDigit = computed(() => /[0-9]/.test(form.value.newPassword))

const validatePasswordMatch = (_rule: any, value: string) => {
  if (value !== form.value.newPassword) {
    return new Error('两次输入的密码不一致')
  }
  return true
}

const validatePasswordStrength = (_rule: any, value: string) => {
  if (value.length < 8) {
    return new Error('密码至少需要8个字符')
  }
  if (!/[A-Z]/.test(value)) {
    return new Error('密码需要包含大写字母')
  }
  if (!/[a-z]/.test(value)) {
    return new Error('密码需要包含小写字母')
  }
  if (!/[0-9]/.test(value)) {
    return new Error('密码需要包含数字')
  }
  return true
}

const rules = {
  oldPassword: { required: true, message: '请输入当前密码', trigger: 'blur' },
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { validator: validatePasswordStrength, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validatePasswordMatch, trigger: 'blur' },
  ],
}

const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    await changePassword(form.value.oldPassword, form.value.newPassword)
    message.success('密码修改成功')

    // 更新用户状态
    if (userStore.user) {
      userStore.user.password_changed = true
    }

    // 如果是强制修改，跳转到首页
    if (isForced.value) {
      router.push('/')
    } else {
      router.back()
    }
  } catch (e: any) {
    message.error(e.response?.data?.error || '密码修改失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.change-password-container {
  max-width: 500px;
  margin: 40px auto;
  padding: 0 20px;
}

.change-password-card {
  border-radius: 12px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
}

.password-tips {
  font-size: 14px;
  color: #666;
}

.password-tips h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
}

.password-tips ul {
  margin: 0;
  padding-left: 20px;
}

.password-tips li {
  margin: 4px 0;
  color: #999;
}

.password-tips li.valid {
  color: #18a058;
}

.password-tips li.valid::marker {
  content: '✓ ';
}
</style>
