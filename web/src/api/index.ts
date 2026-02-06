import axios from 'axios'
import type {
  LoginResponse,
  NodeCreateRequest,
  NodeUpdateRequest,
  ClientCreateRequest,
  ClientUpdateRequest,
  UserCreateRequest,
  UserUpdateRequest,
  NotifyChannelCreateRequest,
  NotifyChannelUpdateRequest,
  AlertRuleCreateRequest,
  AlertRuleUpdateRequest,
  PortForwardCreateRequest,
  PortForwardUpdateRequest,
  NodeGroupCreateRequest,
  NodeGroupUpdateRequest,
  NodeGroupMemberRequest,
  ProxyChainCreateRequest,
  ProxyChainUpdateRequest,
  ProxyChainHopRequest,
  TunnelCreateRequest,
  TunnelUpdateRequest,
  TagCreateRequest,
  TagUpdateRequest,
  PaginationParams,
  ProfileUpdateRequest,
} from '../types'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// 认证
export const login = (username: string, password: string): Promise<LoginResponse> =>
  api.post('/login', { username, password })

// 统计
export const getStats = () => api.get('/stats')

// 全局搜索
export const globalSearch = (query: string) => api.get('/search', { params: { q: query } })

// 节点
export const getNodes = () => api.get('/nodes')
export const getNode = (id: number) => api.get(`/nodes/${id}`)
export const createNode = (data: NodeCreateRequest) => api.post('/nodes', data)
export const updateNode = (id: number, data: NodeUpdateRequest) => api.put(`/nodes/${id}`, data)
export const deleteNode = (id: number) => api.delete(`/nodes/${id}`)
export const applyNodeConfig = (id: number) => api.post(`/nodes/${id}/apply`)
export const syncNodeConfig = (id: number) => api.post(`/nodes/${id}/sync`)
export const getNodeGostConfig = (id: number) => api.get(`/nodes/${id}/gost-config`)
export const getNodeProxyURI = (id: number) => api.get(`/nodes/${id}/proxy-uri`)
export const getNodeInstallScript = (id: number, os: string = 'linux') =>
  api.get(`/nodes/${id}/install-script`, { params: { os } })
export const pingNode = (id: number) => api.get(`/nodes/${id}/ping`)
export const pingAllNodes = () => api.get('/nodes/ping')

// 节点批量操作
export const batchDeleteNodes = (ids: number[]) => api.post('/nodes/batch-delete', { ids })
export const batchSyncNodes = (ids: number[]) => api.post('/nodes/batch-sync', { ids })

// 客户端
export const getClients = () => api.get('/clients')
export const getClient = (id: number) => api.get(`/clients/${id}`)
export const createClient = (data: ClientCreateRequest) => api.post('/clients', data)
export const updateClient = (id: number, data: ClientUpdateRequest) => api.put(`/clients/${id}`, data)
export const deleteClient = (id: number) => api.delete(`/clients/${id}`)
export const getClientInstallScript = (id: number, os: string = 'linux') =>
  api.get(`/clients/${id}/install-script`, { params: { os } })
export const getClientGostConfig = (id: number) => api.get(`/clients/${id}/gost-config`)
export const getClientProxyURI = (id: number) => api.get(`/clients/${id}/proxy-uri`)

// 客户端批量操作
export const batchDeleteClients = (ids: number[]) => api.post('/clients/batch-delete', { ids })

// 用户管理
export const getUsers = () => api.get('/users')
export const getUser = (id: number) => api.get(`/users/${id}`)
export const createUser = (data: UserCreateRequest) => api.post('/users', data)
export const updateUser = (id: number, data: UserUpdateRequest) => api.put(`/users/${id}`, data)
export const deleteUser = (id: number) => api.delete(`/users/${id}`)
export const changePassword = (oldPassword: string, newPassword: string) =>
  api.post('/change-password', { old_password: oldPassword, new_password: newPassword })
export const verifyUserEmail = (id: number) => api.post(`/users/${id}/verify-email`)
export const resendVerification = (id: number) => api.post(`/users/${id}/resend-verification`)

// 个人账户设置
export const getProfile = () => api.get('/profile')
export const updateProfile = (data: ProfileUpdateRequest) => api.put('/profile', data)

// 用户注册和验证 (公开接口)
export const register = (username: string, email: string, password: string) =>
  api.post('/register', { username, email, password })
export const verifyEmail = (token: string) => api.post('/verify-email', { token })
export const forgotPassword = (email: string) => api.post('/forgot-password', { email })
export const resetPassword = (token: string, newPassword: string) =>
  api.post('/reset-password', { token, new_password: newPassword })
export const getRegistrationStatus = () => api.get('/registration-status')

// 流量历史
export const getTrafficHistory = (hours: number = 1, nodeId?: number) => {
  const params: Record<string, number> = { hours }
  if (nodeId !== undefined) {
    params.node_id = nodeId
  }
  return api.get('/traffic-history', { params })
}

// 分页查询接口
export const getNodesPaginated = (params: PaginationParams = {}) =>
  api.get('/nodes/paginated', { params })

export const getClientsPaginated = (params: PaginationParams = {}) =>
  api.get('/clients/paginated', { params })

// 通知渠道
export const getNotifyChannels = () => api.get('/notify-channels')
export const getNotifyChannel = (id: number) => api.get(`/notify-channels/${id}`)
export const createNotifyChannel = (data: NotifyChannelCreateRequest) => api.post('/notify-channels', data)
export const updateNotifyChannel = (id: number, data: NotifyChannelUpdateRequest) => api.put(`/notify-channels/${id}`, data)
export const deleteNotifyChannel = (id: number) => api.delete(`/notify-channels/${id}`)
export const testNotifyChannel = (id: number) => api.post(`/notify-channels/${id}/test`)

// 告警规则
export const getAlertRules = () => api.get('/alert-rules')
export const getAlertRule = (id: number) => api.get(`/alert-rules/${id}`)
export const createAlertRule = (data: AlertRuleCreateRequest) => api.post('/alert-rules', data)
export const updateAlertRule = (id: number, data: AlertRuleUpdateRequest) => api.put(`/alert-rules/${id}`, data)
export const deleteAlertRule = (id: number) => api.delete(`/alert-rules/${id}`)

// 告警日志
export const getAlertLogs = (params: { limit?: number, offset?: number } = {}) =>
  api.get('/alert-logs', { params })

// 操作日志
export const getOperationLogs = (params: { limit?: number, offset?: number, action?: string, resource?: string } = {}) =>
  api.get('/operation-logs', { params })

// 数据导出/导入
export const exportData = (format: 'json' | 'yaml' = 'json', type: 'all' | 'nodes' | 'clients' = 'all') =>
  api.get('/export', { params: { format, type }, responseType: 'blob' })
export const importData = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 数据库备份/恢复
export const backupDatabase = () => api.get('/backup', { responseType: 'blob' })
export const restoreDatabase = (file: File) => {
  const formData = new FormData()
  formData.append('backup', file)
  return api.post('/restore', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 端口转发
export const getPortForwards = () => api.get('/port-forwards')
export const getPortForward = (id: number) => api.get(`/port-forwards/${id}`)
export const createPortForward = (data: PortForwardCreateRequest) => api.post('/port-forwards', data)
export const updatePortForward = (id: number, data: PortForwardUpdateRequest) => api.put(`/port-forwards/${id}`, data)
export const deletePortForward = (id: number) => api.delete(`/port-forwards/${id}`)

// 节点组 (负载均衡)
export const getNodeGroups = () => api.get('/node-groups')
export const getNodeGroup = (id: number) => api.get(`/node-groups/${id}`)
export const createNodeGroup = (data: NodeGroupCreateRequest) => api.post('/node-groups', data)
export const updateNodeGroup = (id: number, data: NodeGroupUpdateRequest) => api.put(`/node-groups/${id}`, data)
export const deleteNodeGroup = (id: number) => api.delete(`/node-groups/${id}`)
export const getNodeGroupMembers = (id: number) => api.get(`/node-groups/${id}/members`)
export const addNodeGroupMember = (id: number, data: NodeGroupMemberRequest) => api.post(`/node-groups/${id}/members`, data)
export const removeNodeGroupMember = (groupId: number, memberId: number) => api.delete(`/node-groups/${groupId}/members/${memberId}`)
export const getNodeGroupConfig = (id: number) => api.get(`/node-groups/${id}/config`)

// 代理链/隧道转发
export const getProxyChains = () => api.get('/proxy-chains')
export const getProxyChain = (id: number) => api.get(`/proxy-chains/${id}`)
export const createProxyChain = (data: ProxyChainCreateRequest) => api.post('/proxy-chains', data)
export const updateProxyChain = (id: number, data: ProxyChainUpdateRequest) => api.put(`/proxy-chains/${id}`, data)
export const deleteProxyChain = (id: number) => api.delete(`/proxy-chains/${id}`)
export const getProxyChainHops = (id: number) => api.get(`/proxy-chains/${id}/hops`)
export const addProxyChainHop = (id: number, data: ProxyChainHopRequest) => api.post(`/proxy-chains/${id}/hops`, data)
export const updateProxyChainHop = (chainId: number, hopId: number, data: ProxyChainHopRequest) => api.put(`/proxy-chains/${chainId}/hops/${hopId}`, data)
export const removeProxyChainHop = (chainId: number, hopId: number) => api.delete(`/proxy-chains/${chainId}/hops/${hopId}`)
export const getProxyChainConfig = (id: number) => api.get(`/proxy-chains/${id}/config`)

// 隧道转发 (入口-出口模式)
export const getTunnels = () => api.get('/tunnels')
export const getTunnel = (id: number) => api.get(`/tunnels/${id}`)
export const createTunnel = (data: TunnelCreateRequest) => api.post('/tunnels', data)
export const updateTunnel = (id: number, data: TunnelUpdateRequest) => api.put(`/tunnels/${id}`, data)
export const deleteTunnel = (id: number) => api.delete(`/tunnels/${id}`)
export const syncTunnel = (id: number) => api.post(`/tunnels/${id}/sync`)
export const getTunnelEntryConfig = (id: number) => api.get(`/tunnels/${id}/entry-config`)
export const getTunnelExitConfig = (id: number) => api.get(`/tunnels/${id}/exit-config`)

// 预配置模板
export const getTemplates = (category?: string) => {
  const params = category ? { category } : {}
  return api.get('/templates', { params })
}
export const getTemplateCategories = () => api.get('/templates/categories')
export const getTemplate = (id: string) => api.get(`/templates/${id}`)

// 客户端模板
export const getClientTemplates = (category?: string) => {
  const params = category ? { category } : {}
  return api.get('/client-templates', { params })
}
export const getClientTemplateCategories = () => api.get('/client-templates/categories')
export const getClientTemplate = (id: string) => api.get(`/client-templates/${id}`)

// 网站配置
export const getPublicSiteConfig = () => axios.get('/api/site-config').then(r => r.data)
export const getSiteConfigs = () => api.get('/site-configs')
export const updateSiteConfigs = (data: Record<string, string>) => api.put('/site-configs', data)

// 节点标签
export const getTags = () => api.get('/tags')
export const getTag = (id: number) => api.get(`/tags/${id}`)
export const createTag = (data: TagCreateRequest) => api.post('/tags', data)
export const updateTag = (id: number, data: TagUpdateRequest) => api.put(`/tags/${id}`, data)
export const deleteTag = (id: number) => api.delete(`/tags/${id}`)
export const getNodesByTag = (tagId: number) => api.get(`/tags/${tagId}/nodes`)

// 节点的标签操作
export const getNodeTags = (nodeId: number) => api.get(`/nodes/${nodeId}/tags`)
export const addNodeTag = (nodeId: number, tagId: number) => api.post(`/nodes/${nodeId}/tags`, { tag_id: tagId })
export const setNodeTags = (nodeId: number, tagIds: number[]) => api.put(`/nodes/${nodeId}/tags`, { tag_ids: tagIds })
export const removeNodeTag = (nodeId: number, tagId: number) => api.delete(`/nodes/${nodeId}/tags/${tagId}`)

export default api
