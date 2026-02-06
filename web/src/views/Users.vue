<template>
  <div class="users">
    <n-card>
      <template #header>
        <n-space justify="space-between" align="center">
          <span>用户管理</span>
          <n-space>
            <n-button type="primary" @click="openCreateModal">
              添加用户
            </n-button>
            <n-button @click="openChangePasswordModal">
              修改密码
            </n-button>
          </n-space>
        </n-space>
      </template>

      <!-- 骨架屏加载 -->
      <TableSkeleton v-if="loading && users.length === 0" :rows="3" :columns="[1, 2, 1, 1, 2]" />

      <!-- 空状态 -->
      <EmptyState
        v-else-if="!loading && users.length === 0"
        type="users"
        action-text="添加用户"
        @action="openCreateModal"
      />

      <!-- 数据表格 -->
      <n-data-table
        v-else
        :columns="columns"
        :data="users"
        :loading="loading"
        :row-key="(row: any) => row.id"
      />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="editingUser ? '编辑用户' : '添加用户'" style="width: 500px;">
      <n-form :model="form" label-placement="left" label-width="100">
        <n-form-item label="用户名">
          <n-input v-model:value="form.username" placeholder="用户名" :disabled="!!editingUser" />
        </n-form-item>
        <n-form-item label="密码" v-if="!editingUser">
          <n-input v-model:value="form.password" type="password" placeholder="密码" show-password-on="click" />
        </n-form-item>
        <n-form-item label="角色">
          <n-select v-model:value="form.role" :options="roleOptions" />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="form.email" placeholder="user@example.com" />
        </n-form-item>
        <n-form-item label="启用账户">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>
        <n-form-item label="邮箱已验证">
          <n-switch v-model:value="form.email_verified" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Change Password Modal -->
    <n-modal v-model:show="showPasswordModal" preset="dialog" title="修改密码" style="width: 500px;">
      <n-form :model="passwordForm" label-placement="left" label-width="100">
        <n-form-item label="旧密码">
          <n-input v-model:value="passwordForm.oldPassword" type="password" placeholder="输入旧密码" show-password-on="click" />
        </n-form-item>
        <n-form-item label="新密码">
          <n-input v-model:value="passwordForm.newPassword" type="password" placeholder="输入新密码" show-password-on="click" />
        </n-form-item>
        <n-form-item label="确认密码">
          <n-input v-model:value="passwordForm.confirmPassword" type="password" placeholder="再次输入新密码" show-password-on="click" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showPasswordModal = false">取消</n-button>
          <n-button type="primary" :loading="changingPassword" @click="handleChangePassword">确认</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NButton, NSpace, NTag, useMessage, useDialog, NTooltip } from 'naive-ui'
import { getUsers, createUser, updateUser, deleteUser, changePassword, verifyUserEmail, resendVerification } from '../api'
import EmptyState from '../components/EmptyState.vue'
import TableSkeleton from '../components/TableSkeleton.vue'
import { useKeyboard } from '../composables/useKeyboard'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const saving = ref(false)
const changingPassword = ref(false)
const users = ref<any[]>([])
const showCreateModal = ref(false)
const showPasswordModal = ref(false)
const editingUser = ref<any>(null)

const roleOptions = [
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
  { label: '只读用户', value: 'viewer' },
]

const defaultForm = () => ({
  username: '',
  password: '',
  role: 'user',
  email: '',
  enabled: true,
  email_verified: true,
})

const form = ref(defaultForm())

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const formatTime = (time: string) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getRoleLabel = (role: string) => {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    user: '普通用户',
    viewer: '只读用户',
  }
  return roleMap[role] || role
}

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '用户名', key: 'username', width: 120 },
  {
    title: '邮箱',
    key: 'email',
    ellipsis: { tooltip: true },
    render: (row: any) => {
      if (!row.email) return '-'
      return h(NSpace, { size: 'small', align: 'center' }, () => [
        row.email,
        row.email_verified
          ? h(NTag, { type: 'success', size: 'tiny' }, () => '已验证')
          : h(NTag, { type: 'warning', size: 'tiny' }, () => '未验证')
      ])
    }
  },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render: (row: any) => {
      const typeMap: Record<string, any> = {
        admin: 'error',
        user: 'success',
        viewer: 'info',
      }
      return h(NTag, { type: typeMap[row.role] || 'default', size: 'small' }, () => getRoleLabel(row.role))
    },
  },
  {
    title: '状态',
    key: 'enabled',
    width: 80,
    render: (row: any) =>
      h(NTag, { type: row.enabled !== false ? 'success' : 'default', size: 'small' }, () => row.enabled !== false ? '启用' : '禁用'),
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 150,
    render: (row: any) => formatTime(row.created_at),
  },
  {
    title: '最后登录',
    key: 'last_login_at',
    width: 150,
    render: (row: any) => {
      if (!row.last_login_at) return '-'
      return h(NTooltip, {}, {
        trigger: () => formatTime(row.last_login_at),
        default: () => `IP: ${row.last_login_ip || '未知'}`
      })
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
        !row.email_verified && row.email ? h(NButton, { size: 'small', type: 'info', onClick: () => handleVerifyEmail(row) }, () => '验证') : null,
        !row.email_verified && row.email ? h(NButton, { size: 'small', type: 'warning', onClick: () => handleResendVerification(row) }, () => '重发') : null,
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row), disabled: row.username === 'admin' }, () => '删除'),
      ]),
  },
]

const loadUsers = async () => {
  loading.value = true
  try {
    const data: any = await getUsers()
    users.value = data || []
  } catch (e) {
    message.error('加载用户失败')
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  form.value = defaultForm()
  editingUser.value = null
  showCreateModal.value = true
}

const handleEdit = (row: any) => {
  editingUser.value = row
  form.value = {
    ...defaultForm(),
    ...row,
    password: '',
    enabled: row.enabled !== false,
    email_verified: row.email_verified || false
  }
  showCreateModal.value = true
}

const handleSave = async () => {
  if (!form.value.username) {
    message.error('请输入用户名')
    return
  }
  if (!editingUser.value && !form.value.password) {
    message.error('请输入密码')
    return
  }

  saving.value = true
  try {
    if (editingUser.value) {
      await updateUser(editingUser.value.id, form.value)
      message.success('用户已更新')
    } else {
      await createUser(form.value)
      message.success('用户已创建')
    }
    showCreateModal.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存用户失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = (row: any) => {
  if (row.username === 'admin') {
    message.error('不能删除管理员账号')
    return
  }

  dialog.warning({
    title: '删除用户',
    content: `确定要删除用户 "${row.username}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteUser(row.id)
        message.success('用户已删除')
        loadUsers()
      } catch (e) {
        message.error('删除用户失败')
      }
    },
  })
}

const handleVerifyEmail = async (row: any) => {
  try {
    await verifyUserEmail(row.id)
    message.success('邮箱已验证')
    loadUsers()
  } catch (e: any) {
    message.error(e.response?.data?.error || '验证失败')
  }
}

const handleResendVerification = async (row: any) => {
  try {
    await resendVerification(row.id)
    message.success('验证邮件已发送')
  } catch (e: any) {
    message.error(e.response?.data?.error || '发送失败')
  }
}

const openChangePasswordModal = () => {
  passwordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  showPasswordModal.value = true
}

const handleChangePassword = async () => {
  if (!passwordForm.value.oldPassword || !passwordForm.value.newPassword || !passwordForm.value.confirmPassword) {
    message.error('请填写所有字段')
    return
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    message.error('两次输入的新密码不一致')
    return
  }
  if (passwordForm.value.newPassword.length < 6) {
    message.error('密码长度至少为 6 位')
    return
  }

  changingPassword.value = true
  try {
    await changePassword(passwordForm.value.oldPassword, passwordForm.value.newPassword)
    message.success('密码已修改')
    showPasswordModal.value = false
  } catch (e: any) {
    message.error(e.response?.data?.error || '修改密码失败')
  } finally {
    changingPassword.value = false
  }
}

onMounted(() => {
  loadUsers()
})

// Keyboard shortcuts
useKeyboard({
  onNew: openCreateModal,
  modalVisible: showCreateModal,
  onSave: handleSave,
})
</script>

<style scoped>
</style>
