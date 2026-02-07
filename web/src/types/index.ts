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
  enabled: boolean
  password_changed: boolean
  email_verified: boolean
  last_login_at?: string
  last_login_ip?: string
  // 套餐
  plan_id?: number
  plan?: Plan
  plan_start_at?: string
  plan_expire_at?: string
  plan_traffic_used?: number
  // 流量配额
  traffic_quota?: number
  quota_used?: number
  quota_reset_day?: number
  quota_reset_at?: string
  quota_exceeded?: boolean
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
  api_port?: number
  proxy_user?: string
  status: string
  traffic_in: number
  traffic_out: number
  connections: number
  // 协议配置
  protocol: string
  transport?: string
  transport_opts?: string
  // Shadowsocks
  ss_method?: string
  // VMess
  vmess_alter_id?: number
  // TLS
  tls_enabled?: boolean
  tls_cert_file?: string
  tls_key_file?: string
  tls_sni?: string
  // WebSocket
  ws_path?: string
  ws_host?: string
  // 限速
  speed_limit?: number
  conn_rate_limit?: number
  // DNS
  dns_server?: string
  // 流量配额
  traffic_quota?: number
  quota_reset_day?: number
  quota_used?: number
  quota_reset_at?: string
  quota_exceeded?: boolean
  // 所有者
  owner_id?: number
  last_seen?: string
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
export type BypassCreateRequest = Record<string, unknown>
export type BypassUpdateRequest = Record<string, unknown>
export type AdmissionCreateRequest = Record<string, unknown>
export type AdmissionUpdateRequest = Record<string, unknown>
export type HostMappingCreateRequest = Record<string, unknown>
export type HostMappingUpdateRequest = Record<string, unknown>
export type TagCreateRequest = { name: string; color?: string }
export type TagUpdateRequest = { name?: string; color?: string }

// 客户端相关
export interface Client extends BaseEntity {
  name: string
  node_id: number
  node?: Node
  local_port: number
  remote_port: number
  proxy_user?: string
  status: string
  traffic_in: number
  traffic_out: number
  // 流量配额
  traffic_quota?: number
  quota_reset_day?: number
  quota_used?: number
  quota_reset_at?: string
  quota_exceeded?: boolean
  // 所有者
  owner_id?: number
  last_seen?: string
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
  cooldown_min?: number
  last_alert_at?: string
}

// 端口转发
export interface PortForward extends BaseEntity {
  name: string
  node_id: number
  type: string
  protocol: string
  local_addr: string
  remote_addr: string
  listen_host: string
  listen_port: number
  target_host: string
  target_port: number
  chain_id?: number
  enabled: boolean
  owner_id?: number
  node_name?: string
}

// 节点组
export interface NodeGroup extends BaseEntity {
  name: string
  strategy: string
  selector?: string
  fail_timeout?: number
  max_fails?: number
  health_check?: boolean
  check_interval?: number
  owner_id?: number
  members?: NodeGroupMember[]
}

export interface NodeGroupMember {
  id: number
  group_id: number
  node_id: number
  weight: number
  priority?: number
  enabled?: boolean
  node?: Node
}

// 代理链
export interface ProxyChain extends BaseEntity {
  name: string
  description?: string
  listen_addr: string
  listen_type: string
  target_addr?: string
  enabled: boolean
  owner_id?: number
  hops?: ProxyChainHop[]
}

export interface ProxyChainHop {
  id: number
  chain_id: number
  node_id: number
  hop_order: number
  enabled: boolean
  node?: Node
}

// 隧道
export interface Tunnel extends BaseEntity {
  name: string
  description?: string
  entry_node_id: number
  entry_port: number
  protocol: string
  exit_node_id: number
  target_addr: string
  enabled: boolean
  traffic_in?: number
  traffic_out?: number
  traffic_quota?: number
  quota_reset_day?: number
  speed_limit?: number
  owner_id?: number
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
  detail: string
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

// 套餐相关
export interface Plan extends BaseEntity {
  name: string
  description?: string
  traffic_quota: number
  speed_limit: number
  duration: number
  max_nodes: number
  max_clients: number
  enabled: boolean
  sort_order: number
  user_count?: number
}

export type PlanCreateRequest = Record<string, unknown>
export type PlanUpdateRequest = Record<string, unknown>

// Bypass 分流规则
export interface Bypass extends BaseEntity {
  name: string
  whitelist: boolean
  matchers: string // JSON array
  node_id?: number
  owner_id?: number
}

// Admission 准入控制
export interface Admission extends BaseEntity {
  name: string
  whitelist: boolean
  matchers: string // JSON array
  node_id?: number
  owner_id?: number
}

// HostMapping 主机映射
export interface HostMapping extends BaseEntity {
  name: string
  mappings: string // JSON array of {hostname, ip, prefer}
  node_id?: number
  owner_id?: number
}

// Ingress 反向代理
export interface Ingress extends BaseEntity {
  name: string
  rules: string // JSON: [{"hostname":"example.com","endpoint":"192.168.1.1:8080"}]
  node_id?: number
  owner_id?: number
}

// Recorder 流量记录
export interface Recorder extends BaseEntity {
  name: string
  type: string // file, redis, http
  config: string // JSON config
  node_id?: number
  owner_id?: number
}

export type IngressCreateRequest = Record<string, unknown>
export type IngressUpdateRequest = Record<string, unknown>
export type RecorderCreateRequest = Record<string, unknown>
export type RecorderUpdateRequest = Record<string, unknown>

// Router 路由管理
export interface Router extends BaseEntity {
  name: string
  routes: string // JSON: [{"net":"192.168.0.0/16","gateway":"192.168.0.1"}]
  node_id?: number
  owner_id?: number
}

// SD 服务发现
export interface SD extends BaseEntity {
  name: string
  type: string // http, consul, etcd, redis
  config: string // JSON config
  node_id?: number
  owner_id?: number
}

export type RouterCreateRequest = Record<string, unknown>
export type RouterUpdateRequest = Record<string, unknown>
export type SDCreateRequest = Record<string, unknown>
export type SDUpdateRequest = Record<string, unknown>

// 配置版本历史
export interface ConfigVersion extends BaseEntity {
  node_id: number
  config: string
  comment?: string
}
