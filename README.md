# GOST Panel

[![GitHub Release](https://img.shields.io/github/v/release/AliceNetworks/gost-panel)](https://github.com/AliceNetworks/gost-panel/releases)
[![GitHub License](https://img.shields.io/github/license/AliceNetworks/gost-panel)](https://github.com/AliceNetworks/gost-panel/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![Vue Version](https://img.shields.io/badge/vue-3.x-green.svg)](https://vuejs.org/)

基于 [GOST v3](https://github.com/go-gost/gost) 的现代化代理管理面板。

## 功能特性

- **多协议支持**: SOCKS5, HTTP, Shadowsocks, Trojan, VMess
- **多传输层**: TCP, TLS, WebSocket, HTTP/2, QUIC, KCP, gRPC
- **节点管理**: 多 VPS 节点管理，实时状态监控，批量操作
- **客户端管理**: 反向隧道客户端，访问内网服务
- **负载均衡**: 节点组支持轮询、随机、哈希策略
- **端口转发**: TCP/UDP/RTCP/RUDP 转发规则
- **流量监控**: 实时统计和历史图表
- **告警系统**: Telegram、Webhook、邮件通知
- **多用户**: 基于角色的访问控制 (admin/user/viewer)
- **流量配额**: 节点和客户端流量限制
- **实时推送**: WebSocket 实时节点状态更新
- **隧道转发**: 入口节点连接出口节点，链式代理
- **代理链**: 多跳代理配置
- **自动更新**: Agent 自动检测并更新新版本
- **暗色主题**: 现代 Glassmorphism 风格暗色 UI
- **操作日志**: 完整的审计日志记录
- **数据导出**: 支持 JSON/YAML 格式导入导出
- **配置备份**: 一键备份和恢复数据库

## 快速开始

### 二进制安装 (推荐)

从 [Releases](https://github.com/AliceNetworks/gost-panel/releases) 页面下载对应平台的预编译二进制文件。

```bash
# 下载最新版本 (以 Linux amd64 为例)
VERSION=$(curl -s https://api.github.com/repos/AliceNetworks/gost-panel/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -fsSL "https://github.com/AliceNetworks/gost-panel/releases/download/${VERSION}/gost-panel_linux_amd64.tar.gz" | tar -xz
chmod +x gost-panel

# 运行
./gost-panel
```

### 源码安装

#### 环境要求

- Go 1.21+
- Node.js 18+
- 节点需安装 [GOST v3](https://github.com/go-gost/gost/releases)

```bash
# 克隆仓库
git clone https://github.com/AliceNetworks/gost-panel.git
cd gost-panel

# 先编译前端 (后端使用 go:embed 嵌入前端文件)
cd web
npm install
NODE_OPTIONS="--max-old-space-size=1024" npm run build
cd ..

# 编译后端 (低内存服务器限制 CPU)
GOMAXPROCS=1 go build -o gost-panel ./cmd/panel

# 运行
./gost-panel
```

> **注意**: 必须先编译前端再编译后端，因为 Go 二进制文件使用 `go:embed` 嵌入前端资源。

### Docker 部署

```bash
docker-compose up -d
```

### 配置文件

创建 `config.yaml`:

```yaml
server:
  port: 8080
  jwt_secret: "your-secret-key"

database:
  path: "./data/panel.db"
```

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

**首次登录后请立即修改默认密码！**

## 项目结构

```
gost-panel/
├── cmd/
│   ├── panel/       # 面板主程序
│   └── agent/       # 节点 Agent
├── internal/
│   ├── api/         # HTTP API 处理
│   ├── config/      # 配置管理
│   ├── gost/        # GOST API 客户端
│   ├── model/       # 数据库模型
│   ├── notify/      # 告警服务
│   └── service/     # 业务逻辑
├── web/             # Vue.js 前端
└── scripts/         # 部署脚本
```

## API 接口

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/login | 用户登录 |
| POST | /api/register | 用户注册 |
| POST | /api/change-password | 修改密码 |
| GET | /api/profile | 获取个人资料 |
| PUT | /api/profile | 更新个人资料 |

### 管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/stats | 仪表盘统计 |
| GET | /api/search | 全局搜索 |
| GET | /api/nodes | 节点列表 |
| POST | /api/nodes | 创建节点 |
| GET | /api/clients | 客户端列表 |
| POST | /api/clients | 创建客户端 |
| GET | /api/port-forwards | 端口转发列表 |
| GET | /api/node-groups | 节点组列表 |
| GET | /api/notify-channels | 通知渠道列表 |
| GET | /api/alert-rules | 告警规则列表 |
| GET | /api/tunnels | 隧道列表 |
| GET | /api/proxy-chains | 代理链列表 |
| GET | /api/operation-logs | 操作日志 |
| GET | /ws | WebSocket 实时推送 |

### 数据管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/export | 导出数据 (JSON/YAML) |
| POST | /api/import | 导入数据 |
| GET | /api/backup | 下载数据库备份 |
| POST | /api/restore | 恢复数据库 |

## 节点部署

### 一键安装 (推荐)

在面板创建节点后，从节点详情页复制安装命令。

**Linux:**
```bash
curl -fsSL "https://your-panel.com/scripts/install-node.sh" | bash -s -- -p "https://your-panel.com" -t "YOUR_TOKEN"
```

**Windows (管理员 PowerShell):**
```powershell
irm "https://your-panel.com/scripts/install-node.ps1" -OutFile "$env:TEMP\install-node.ps1"; & "$env:TEMP\install-node.ps1" -PanelUrl "https://your-panel.com" -Token "YOUR_TOKEN"
```

### 支持的平台

| 操作系统 | 架构 | 服务管理 |
|----------|------|----------|
| Linux | amd64, arm64, armv7, armv6, mips, mipsle, mips64 | systemd, sysvinit, procd (OpenWrt), openrc |
| Windows | amd64, arm64, x86 | NSSM 服务, 计划任务 |

### 手动安装

详细的手动安装步骤请参考 [节点部署指南](https://github.com/AliceNetworks/gost-panel/wiki/Node-Deployment) (Wiki)。

## 客户端部署

用于反向隧道客户端 (访问内网服务):

**Linux:**
```bash
curl -fsSL "https://your-panel.com/scripts/install-client.sh" | bash -s -- -p "https://your-panel.com" -t "CLIENT_TOKEN"
```

**Windows:**
```powershell
irm "https://your-panel.com/scripts/install-client.ps1" -OutFile "$env:TEMP\install-client.ps1"; & "$env:TEMP\install-client.ps1" -PanelUrl "https://your-panel.com" -Token "CLIENT_TOKEN"
```

## 开发

```bash
# 后端开发
go run ./cmd/panel

# 前端开发 (热重载)
cd web
npm run dev
```

## 技术栈

- **后端**: [Go](https://go.dev/), [Gin](https://github.com/gin-gonic/gin), [GORM](https://gorm.io/), SQLite
- **前端**: [Vue 3](https://vuejs.org/), TypeScript, [Naive UI](https://www.naiveui.com/), [ECharts](https://echarts.apache.org/)
- **构建**: [Vite](https://vitejs.dev/), Docker

## 相关链接

- [GOST 官方文档](https://gost.run/)
- [GOST GitHub](https://github.com/go-gost/gost)
- [问题反馈](https://github.com/AliceNetworks/gost-panel/issues)
- [更新日志](https://github.com/AliceNetworks/gost-panel/releases)

## 许可证

[MIT License](https://github.com/AliceNetworks/gost-panel/blob/main/LICENSE)

## 贡献

欢迎提交 [Pull Request](https://github.com/AliceNetworks/gost-panel/pulls)。重大更改请先开 [Issue](https://github.com/AliceNetworks/gost-panel/issues) 讨论。
