# GOST Panel

[![GitHub Release](https://img.shields.io/github/v/release/AliceNetworks/gost-panel)](https://github.com/AliceNetworks/gost-panel/releases)
[![GitHub License](https://img.shields.io/github/license/AliceNetworks/gost-panel)](https://github.com/AliceNetworks/gost-panel/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![Vue Version](https://img.shields.io/badge/vue-3.x-green.svg)](https://vuejs.org/)

基于 [GOST v3](https://github.com/go-gost/gost) 的全功能代理管理面板，提供现代化 Web UI 和完整的 API 接口。

## 功能特性

### 协议与传输

- **17 种代理协议**: SOCKS5, SOCKS4/4A, HTTP, HTTP/2, Shadowsocks (SS), Shadowsocks UDP (SSU), Auto (多协议探测), Relay, TCP, UDP, SNI, DNS, SSH, Redirect (TCP 透明代理), REDU (UDP 透明代理), TUN (全局代理), TAP (二层网络)
- **26 种传输方式**: TCP, UDP, TCP+UDP, TLS, mTLS, mTCP, WS, WSS, mWS, mWSS, H2, H2C, HTTP/3, H3 (HTTP/3 Tunnel), WebTransport (WT), QUIC, KCP, gRPC, PHT, PHTS, SSH, DTLS, Obfs-HTTP, Obfs-TLS, Fake TCP (FTCP), ICMP Tunnel
- **端口转发**: TCP/UDP/RTCP (远程反向 TCP)/RUDP (远程反向 UDP)/Relay 中继，支持代理链
- **隧道转发**: 入口节点 → 出口节点链式代理
- **代理链**: 多跳代理，自定义跳点顺序

### 节点与客户端

- **多节点管理**: 多 VPS 节点管理，实时状态监控，批量操作 (启用/禁用/同步/删除)
- **Agent 自动化**: 一键安装脚本 (Linux/Windows)，自动注册、心跳、配置同步、版本更新
- **客户端管理**: 反向隧道客户端，访问内网服务
- **节点组/负载均衡**: 轮询、随机、哈希策略，健康检查，权重/优先级配置
- **17 种架构支持**: linux/amd64, arm64, armv7, armv6, mips/mipsle/mips64, windows/amd64+arm64+x86 等

### GOST 配置对象 (全部 14 种)

- **Services** - 代理服务
- **Chains** - 转发链
- **Bypasses** - 分流规则 (黑/白名单)
- **Admissions** - 准入控制
- **Hosts** - 主机映射/域名解析
- **Resolvers** - DNS 解析器
- **Observers** - 流量观测
- **Authers** - 认证器
- **Limiters** - 速率/带宽限制
- **RLimiters** - 连接频率限制
- **Ingresses** - HTTP/HTTPS 反向代理路由
- **Recorders** - 流量记录 (File/Redis/HTTP)
- **Routers** - 自定义路由/网关
- **SDs** - 服务发现 (Consul/Etcd/Redis/HTTP)

### 高级功能

- **探测抵抗**: code/web/host/file 伪装模式
- **PROXY Protocol**: v1/v2 保留源 IP
- **KCP 高级参数**: MTU/发送窗口/接收窗口/分片可调
- **TLS 高级配置**: ALPN 协议列表支持
- **DNS 高级配置**: 支持任意 DNS 协议
- **Plugin 系统**: JSON 配置字段支持

### 面板功能

- **Dashboard**: 实时统计 + ECharts 图表 + 可拖拽卡片布局
- **WebSocket 实时推送**: 节点/客户端状态实时更新
- **双因素认证 (2FA)**: TOTP (Google/Microsoft Authenticator) + 备份码
- **套餐管理**: 流量配额、速率限制、资源限制 (节点/客户端/隧道/转发/代理链/节点组)
- **通知告警**: Telegram / Webhook / SMTP 邮件
- **操作日志**: 完整审计日志
- **配置版本历史**: 自动快照、手动创建、恢复、删除
- **一键克隆**: 节点/客户端/端口转发/隧道/代理链/节点组/规则 (Bypass/Admission/Ingress/Recorder/Router/SD)
- **全局搜索**: 所有列表页支持实时搜索过滤
- **数据导出**: JSON/YAML 格式导入导出 + 数据库备份恢复
- **Swagger API 文档**: `/api/docs`
- **暗色主题**: Glassmorphism 风格 UI
- **移动端适配**: 响应式布局
- **快捷键**: 快速新建/保存操作
- **多用户**: admin/user/viewer 角色权限控制
- **资源隔离**: 用户只能操作自己的资源 (ownership 权限检查)
- **多架构构建**: Panel (linux/amd64, linux/arm64, windows/amd64), Agent (17 架构)

## 快速开始

### 一键安装 (推荐)

```bash
curl -fsSL https://raw.githubusercontent.com/AliceNetworks/gost-panel/main/scripts/install.sh | bash
```

自动完成：检测架构、下载最新版本、安装到 `/opt/gost-panel`、配置 systemd 服务、生成 JWT 密钥并启动。

安装完成后访问 `http://your-ip:8080` 即可使用。

**升级面板：** 重新执行上述命令即可。

### 一键卸载

```bash
curl -fsSL https://raw.githubusercontent.com/AliceNetworks/gost-panel/main/scripts/uninstall.sh | bash
```

卸载时可选择保留数据库备份。

### 二进制手动安装

从 [Releases](https://github.com/AliceNetworks/gost-panel/releases) 页面下载对应平台的预编译二进制文件。

```bash
# 下载最新版本 (以 Linux amd64 为例)
VERSION=$(curl -s https://api.github.com/repos/AliceNetworks/gost-panel/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -fsSL "https://github.com/AliceNetworks/gost-panel/releases/download/${VERSION}/gost-panel-linux-amd64.tar.gz" | tar -xz
chmod +x gost-panel-linux-amd64

# 运行 (默认端口 8080)
./gost-panel-linux-amd64

# 自定义端口
./gost-panel-linux-amd64 -listen :9000
```

### 命令行参数

```bash
gost-panel [options]

选项:
  -listen string    监听地址 (默认 ":8080")
                    示例: :9000, 0.0.0.0:8080, 127.0.0.1:8080
  -db string        数据库路径 (默认 "./data/panel.db")
  -debug            启用调试模式
  -version          显示版本信息
  -help             显示帮助
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| LISTEN_ADDR | 监听地址 | :8080 |
| DB_PATH | 数据库路径 | ./data/panel.db |
| JWT_SECRET | JWT 密钥 (生产环境必须设置) | 随机生成 |
| DEBUG | 启用调试模式 | false |
| ALLOWED_ORIGINS | 允许的 CORS 来源 (逗号分隔) | - |

### Docker 部署

```bash
docker run -d \
  --name gost-panel \
  -p 8080:8080 \
  -v gost-panel-data:/app/data \
  ghcr.io/alicenetworks/gost-panel:latest
```

### Docker Compose 部署

创建 `docker-compose.yml`：

```yaml
services:
  gost-panel:
    image: ghcr.io/alicenetworks/gost-panel:latest
    container_name: gost-panel
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - JWT_SECRET=change-me-in-production   # 生产环境请修改
    restart: unless-stopped
```

启动：

```bash
docker compose up -d
```

常用操作：

```bash
# 查看日志
docker compose logs -f

# 更新到最新版本
docker compose pull && docker compose up -d

# 停止并删除容器 (数据保留在 ./data 目录)
docker compose down
```

### 源码构建

```bash
git clone https://github.com/AliceNetworks/gost-panel.git
cd gost-panel

# 使用构建脚本 (推荐)
./scripts/build.sh all

# 或手动构建
cd web && npm install && npm run build && cd ..
go build -o gost-panel ./cmd/panel
./gost-panel
```

> **注意**: 必须先编译前端再编译后端，因为 Go 使用 `go:embed` 嵌入前端资源。

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

**首次登录后请立即修改默认密码！** (系统会自动提示)

## 节点部署

在面板创建节点后，从节点详情页复制安装命令：

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

## 客户端部署

用于反向隧道 (访问内网服务):

**Linux:**
```bash
curl -fsSL "https://your-panel.com/scripts/install-client.sh" | bash -s -- -p "https://your-panel.com" -t "CLIENT_TOKEN"
```

**Windows:**
```powershell
irm "https://your-panel.com/scripts/install-client.ps1" -OutFile "$env:TEMP\install-client.ps1"; & "$env:TEMP\install-client.ps1" -PanelUrl "https://your-panel.com" -Token "CLIENT_TOKEN"
```

## API 文档

面板内置 Swagger API 文档，访问 `http://your-panel:8080/api/docs` 查看完整接口列表。

### 主要接口

| 类别 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 认证 | POST | /api/login | 用户登录 |
| 认证 | POST | /api/login/2fa | 双因素验证 |
| 认证 | POST | /api/register | 用户注册 |
| 节点 | GET/POST | /api/nodes | 列表/创建 |
| 客户端 | GET/POST | /api/clients | 列表/创建 |
| 端口转发 | GET/POST | /api/port-forwards | 列表/创建 |
| 隧道 | GET/POST | /api/tunnels | 列表/创建 |
| 代理链 | GET/POST | /api/proxy-chains | 列表/创建 |
| 节点组 | GET/POST | /api/node-groups | 列表/创建 |
| 规则 | GET/POST | /api/bypasses | 分流规则 |
| 规则 | GET/POST | /api/admissions | 准入控制 |
| 规则 | GET/POST | /api/host-mappings | 主机映射 |
| 规则 | GET/POST | /api/ingresses | 反向代理 |
| 规则 | GET/POST | /api/recorders | 流量记录 |
| 规则 | GET/POST | /api/routers | 路由管理 |
| 规则 | GET/POST | /api/sds | 服务发现 |
| 用户 | GET/POST | /api/users | 用户管理 |
| 套餐 | GET/POST | /api/plans | 套餐管理 |
| 通知 | GET/POST | /api/notify-channels | 通知渠道 |
| 告警 | GET/POST | /api/alert-rules | 告警规则 |
| 统计 | GET | /api/stats | Dashboard 统计 |
| 数据 | GET/POST | /api/export, /api/import | 数据导入导出 |
| 实时 | GET | /ws | WebSocket 推送 |

## 项目结构

```
gost-panel/
├── cmd/
│   ├── panel/       # 面板主程序
│   └── agent/       # 节点 Agent
├── internal/
│   ├── api/         # HTTP API 处理 + Swagger 文档
│   ├── config/      # 配置管理
│   ├── gost/        # GOST 配置生成器
│   ├── model/       # 数据库模型
│   ├── notify/      # 告警通知服务
│   └── service/     # 业务逻辑 + 权限控制
├── web/             # Vue 3 + TypeScript 前端
│   ├── src/views/   # 页面组件
│   ├── src/api/     # API 调用
│   ├── src/stores/  # Pinia 状态管理
│   └── src/types/   # TypeScript 类型定义
└── scripts/         # 安装/卸载/构建脚本
```

## 技术栈

- **后端**: [Go](https://go.dev/), [Gin](https://github.com/gin-gonic/gin), [GORM](https://gorm.io/), SQLite
- **前端**: [Vue 3](https://vuejs.org/), TypeScript, [Naive UI](https://www.naiveui.com/), [ECharts](https://echarts.apache.org/)
- **构建**: [Vite](https://vitejs.dev/), GitHub Actions
- **安全**: JWT 认证, TOTP 双因素, bcrypt 密码哈希, 资源隔离

## 相关链接

- [GOST 官方文档](https://gost.run/)
- [GOST GitHub](https://github.com/go-gost/gost)
- [问题反馈](https://github.com/AliceNetworks/gost-panel/issues)
- [更新日志](https://github.com/AliceNetworks/gost-panel/releases)

## Stargazers

感谢所有 Star 支持者！

[![Star History Chart](https://api.star-history.com/svg?repos=AliceNetworks/gost-panel&type=Date)](https://star-history.com/#AliceNetworks/gost-panel&Date)

## 贡献

欢迎提交 [Pull Request](https://github.com/AliceNetworks/gost-panel/pulls)。重大更改请先开 [Issue](https://github.com/AliceNetworks/gost-panel/issues) 讨论。

## 免责声明

1. 本项目仅供学习交流和合法用途，使用者必须遵守所在地区的法律法规。
2. 本项目是 [GOST v3](https://github.com/go-gost/gost) 的管理面板，不对 GOST 核心功能负责。
3. 使用本软件所产生的任何直接或间接后果，由使用者自行承担，项目开发者不承担任何责任。
4. 严禁将本项目用于任何违法违规用途，包括但不限于非法代理、翻墙、数据窃取等。
5. 如果您不同意以上条款，请勿使用本项目。

## 许可证

本项目基于 [MIT License](LICENSE) 开源。

```
MIT License

Copyright (c) 2025 AliceNetworks

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
