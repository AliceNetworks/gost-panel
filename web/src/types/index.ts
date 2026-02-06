// 基础类型
export interface BaseEntity {
  id: number
  created_at?: string
  updated_at?: string
}

// 用户相关
export interface User extends BaseEntity {
  username: string
  email?: string
  role: string
  password_changed: boolean
  email_verified: boolean
  last_login?: string
  last_login_ip?: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface ProfileUpdateRequest {
  email?: string
}

// 节点相关
export interface Node extends BaseEntity {
  name: string
  host: string
  port: number
  protocol: string
  status: string
  username?: string
  password?: string
  config?: Record<string, unknown>
  gost_api_addr?: string
  gost_api_auth?: string
  traffic_in: number
  traffic_out: number
  connections: number
  latency?: number
  tags?: Tag[]
}

// 通用请求类型 - 使用 Record 来匹配各种表单数据
export type NodeCreateRequest = Record<string, unknown>
export type NodeUpdateRequest = Record<string, unknown>
export type ClientCreateRequest = Record<string, unknown>
export type ClientUpdateRequest = Record<string, unknown>
export type UserCreateRequest = Record<string, unknown>
export type UserUpdateRequest = Record<string, unknown>
export type NotifyChannelCreateRequest = Record<string, unknown>
export type NotifyChannelUpdateRequest = Record<string, unknown>
export type AlertRuleCreateRequest = Record<string, unknown>
export type AlertRuleUpdateRequest = Record<string, unknown>
export type PortForwardCreateRequest = Record<string, unknown>
export type PortForwardUpdateRequest = Record<string, unknown>
export type NodeGroupCreateRequest = Record<string, unknown>
export type NodeGroupUpdateRequest = Record<string, unknown>
export type NodeGroupMemberRequest = Record<string, unknown>
export type ProxyChainCreateRequest = Record<string, unknown>
export type ProxyChainUpdateRequest = Record<string, unknown>
export type ProxyChainHopRequest = Record<string, unknown>
export type TunnelCreateRequest = Record<string, unknown>
export type TunnelUpdateRequest = Record<string, unknown>
export type TagCreateRequest = { name: string; color?: string }
export type TagUpdateRequest = { name?: string; color?: string }

// 客户端相关
export interface Client extends BaseEntity {
  name: string
  token: string
  node_id: number
  status: string
  listen_port: number
  target_addr: string
  config?: Record<string, unknown>
  traffic_in: number
  traffic_out: number
  last_heartbeat?: string
  node?: Node
}

// 通知渠道
export interface NotifyChannel extends BaseEntity {
  name: string
  type: string
  enabled: boolean
  config: Record<string, string>
}

// 告警规则
export interface AlertRule extends BaseEntity {
  name: string
  type: string
  enabled: boolean
  condition: Record<string, unknown>
  channel_ids: number[]
}

// 端口转发
export interface PortForward extends BaseEntity {
  name: string
  node_id: number
  protocol: string
  listen_port: number
  target_addr: string
  enabled: boolean
  node?: Node
}

// 节点组
export interface NodeGroup extends BaseEntity {
  name: string
  strategy: string
  selector?: Record<string, unknown>
  members?: NodeGroupMember[]
}

export interface NodeGroupMember extends BaseEntity {
  group_id: number
  node_id: number
  weight: number
  node?: Node
}

// 代理链
export interface ProxyChain extends BaseEntity {
  name: string
  enabled: boolean
  hops?: ProxyChainHop[]
}

export interface ProxyChainHop extends BaseEntity {
  chain_id: number
  node_id: number
  order: number
  node?: Node
}

// 隧道
export interface Tunnel extends BaseEntity {
  name: string
  entry_node_id: number
  exit_node_id: number
  listen_port: number
  target_addr: string
  enabled: boolean
  entry_node?: Node
  exit_node?: Node
}

// 标签
export interface Tag extends BaseEntity {
  name: string
  color?: string
}

// 统计
export interface Stats {
  total_nodes: number
  online_nodes: number
  total_clients: number
  online_clients: number
  total_users: number
  total_traffic_in: number
  total_traffic_out: number
  total_connections: number
}

// 流量历史
export interface TrafficHistory {
  timestamp: string
  traffic_in: number
  traffic_out: number
  connections: number
}

// 操作日志
export interface OperationLog extends BaseEntity {
  user_id: number
  username: string
  action: string
  resource: string
  resource_id: number
  details: string
  ip: string
  user_agent: string
  status: string
}

// 分页
export interface PaginationParams {
  page?: number
  page_size?: number
  search?: string
  sort_by?: string
  sort_desc?: boolean
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

// 搜索结果
export interface SearchResult {
  nodes: Node[]
  clients: Client[]
  users: User[]
}

// 网站配置
export interface SiteConfig {
  site_name?: string
  site_url?: string
  favicon_url?: string
  logo_url?: string
  footer_text?: string
  custom_css?: string
  registration_enabled?: string
  email_verification_enabled?: string
}
