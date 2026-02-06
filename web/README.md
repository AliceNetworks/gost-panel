# GOST Panel 前端

基于 Vue 3 + TypeScript + Vite 的 GOST 管理面板前端。

## 技术栈

- Vue 3 (Composition API + `<script setup>`)
- TypeScript
- Naive UI 组件库
- ECharts 图表
- Vite 构建工具

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 类型检查
npm run type-check

# 构建生产版本
npm run build
```

## 构建注意事项

低内存服务器请限制 Node.js 内存使用：

```bash
NODE_OPTIONS="--max-old-space-size=1024" npm run build
```

## 目录结构

```
src/
├── api/          # API 请求
├── components/   # 公共组件
├── router/       # 路由配置
├── stores/       # Pinia 状态管理
├── views/        # 页面组件
└── App.vue       # 根组件
```

## 功能页面

- **Dashboard** - 仪表盘，实时统计和图表
- **Nodes** - 节点管理
- **Clients** - 客户端管理
- **PortForwards** - 端口转发
- **NodeGroups** - 节点组/负载均衡
- **Tunnels** - 隧道管理
- **ProxyChains** - 代理链
- **Notify** - 通知告警配置
- **Users** - 用户管理
- **Settings** - 系统设置
