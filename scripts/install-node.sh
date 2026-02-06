#!/bin/bash
# GOST Panel 节点端安装脚本
# 支持: Linux (amd64, arm64, armv7, armv6, mips, mipsle, mips64)
# 用法: curl -fsSL https://panel:8080/scripts/install-node.sh | bash -s -- -p http://panel:8080 -t TOKEN

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

REPO="AliceNetworks/gost-panel"
PANEL_URL=""
TOKEN=""
INSTALL_DIR="/opt/gost-panel"
GOST_VERSION="3.0.0-rc10"
FORCE_ARCH=""

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--panel) PANEL_URL="$2"; shift 2 ;;
        -t|--token) TOKEN="$2"; shift 2 ;;
        -a|--arch) FORCE_ARCH="$2"; shift 2 ;;
        -h|--help)
            echo "GOST Panel Node Installer"
            echo ""
            echo "Usage: $0 -p <panel_url> -t <token> [-a <arch>]"
            echo ""
            echo "Options:"
            echo "  -p, --panel   Panel URL (e.g., http://panel.example.com:8080)"
            echo "  -t, --token   Node token from panel"
            echo "  -a, --arch    Force architecture (amd64, arm64, armv7, armv6, mips, mipsle)"
            exit 0
            ;;
        *) log_error "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$PANEL_URL" || -z "$TOKEN" ]]; then
    log_error "Missing required parameters"
    echo "Usage: $0 -p <panel_url> -t <token>"
    exit 1
fi

echo "========================================"
echo "    GOST Panel Node Installer"
echo "========================================"
echo ""
log_info "Panel: $PANEL_URL"

# 检测系统架构
detect_arch() {
    if [[ -n "$FORCE_ARCH" ]]; then
        echo "$FORCE_ARCH"
        return
    fi

    local arch=$(uname -m)
    local gost_arch=""

    case $arch in
        x86_64|amd64)
            gost_arch="amd64"
            ;;
        aarch64|arm64)
            gost_arch="arm64"
            ;;
        armv7l|armv7)
            gost_arch="armv7"
            ;;
        armv6l|armv6)
            gost_arch="armv6"
            ;;
        armv5*)
            gost_arch="armv5"
            ;;
        mips)
            # 检测大小端
            if echo -n I | od -to2 | head -n1 | cut -f2 -d" " | cut -c6 | grep -q 1; then
                gost_arch="mipsle"
            else
                gost_arch="mips"
            fi
            ;;
        mips64)
            if echo -n I | od -to2 | head -n1 | cut -f2 -d" " | cut -c6 | grep -q 1; then
                gost_arch="mips64le"
            else
                gost_arch="mips64"
            fi
            ;;
        i386|i686)
            gost_arch="386"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac

    echo "$gost_arch"
}

GOST_ARCH=$(detect_arch)
log_info "Detected architecture: $GOST_ARCH"

# 检测包管理器
detect_pkg_manager() {
    if command -v apt-get &> /dev/null; then
        echo "apt"
    elif command -v yum &> /dev/null; then
        echo "yum"
    elif command -v apk &> /dev/null; then
        echo "apk"
    elif command -v opkg &> /dev/null; then
        echo "opkg"
    else
        echo "unknown"
    fi
}

PKG_MANAGER=$(detect_pkg_manager)
log_info "Package manager: $PKG_MANAGER"

# 安装依赖
install_deps() {
    log_info "Installing dependencies..."
    case $PKG_MANAGER in
        apt)
            apt-get update -qq
            apt-get install -y -qq curl tar ca-certificates
            ;;
        yum)
            yum install -y -q curl tar ca-certificates
            ;;
        apk)
            apk add --no-cache curl tar ca-certificates
            ;;
        opkg)
            opkg update
            opkg install curl tar ca-certificates
            ;;
        *)
            log_warn "Unknown package manager, assuming dependencies are installed"
            ;;
    esac
}

# 检测 init 系统
detect_init_system() {
    if command -v systemctl &> /dev/null && systemctl --version &> /dev/null 2>&1; then
        echo "systemd"
    elif [[ -f /etc/init.d/rcS ]] || command -v update-rc.d &> /dev/null; then
        echo "sysvinit"
    elif [[ -f /etc/rc.common ]]; then
        echo "procd"  # OpenWrt
    elif command -v rc-service &> /dev/null; then
        echo "openrc"
    else
        echo "unknown"
    fi
}

INIT_SYSTEM=$(detect_init_system)
log_info "Init system: $INIT_SYSTEM"

# 安装 GOST
install_gost() {
    log_info "[1/5] Installing GOST..."

    if command -v gost &> /dev/null; then
        log_info "GOST already installed: $(which gost)"
        return
    fi

    local gost_url="https://github.com/go-gost/gost/releases/download/v${GOST_VERSION}/gost_${GOST_VERSION}_linux_${GOST_ARCH}.tar.gz"

    log_info "Downloading GOST from $gost_url"
    curl -fsSL "$gost_url" -o /tmp/gost.tar.gz

    mkdir -p /tmp/gost-extract
    tar -xzf /tmp/gost.tar.gz -C /tmp/gost-extract

    mv /tmp/gost-extract/gost /usr/local/bin/
    chmod +x /usr/local/bin/gost

    rm -rf /tmp/gost.tar.gz /tmp/gost-extract
    log_info "GOST installed to /usr/local/bin/gost"
}

# 下载 Agent
download_agent() {
    log_info "[2/5] Downloading agent..."

    mkdir -p "$INSTALL_DIR"

    # 如果服务已存在，先停止（避免文件被占用）
    if systemctl is-active --quiet gost-node 2>/dev/null; then
        log_info "Stopping existing gost-node service..."
        systemctl stop gost-node 2>/dev/null || true
    fi

    # 删除旧的 agent 文件
    rm -f "$INSTALL_DIR/gost-agent"

    # 获取最新版本
    local latest_version=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$latest_version" ]; then
        latest_version="v1.0.0"
    fi

    # 从 GitHub Releases 下载
    local agent_url="https://github.com/$REPO/releases/download/$latest_version/gost-agent-linux-$GOST_ARCH"

    log_info "Downloading agent from GitHub ($latest_version)..."
    if curl -fsSL "$agent_url" -o "$INSTALL_DIR/gost-agent" 2>/dev/null; then
        chmod +x "$INSTALL_DIR/gost-agent"
        log_info "Agent downloaded to $INSTALL_DIR/gost-agent"
        return 0
    fi

    # 回退: 从面板下载
    log_warn "GitHub download failed, trying panel..."
    local panel_agent_url="$PANEL_URL/agent/download/linux/$GOST_ARCH"
    if curl -fsSL "$panel_agent_url" -o "$INSTALL_DIR/gost-agent" 2>/dev/null; then
        chmod +x "$INSTALL_DIR/gost-agent"
        log_info "Agent downloaded from panel"
        return 0
    else
        log_warn "Agent binary not available, will use GOST directly"
        return 1
    fi
}

# 下载配置
download_config() {
    log_info "[3/5] Downloading config..."

    mkdir -p /etc/gost
    curl -fsSL "$PANEL_URL/agent/config/$TOKEN" -o /etc/gost/gost.yml
    log_info "Config saved to /etc/gost/gost.yml"
}

# 创建 systemd 服务
create_systemd_service() {
    local use_agent=$1

    if [[ "$use_agent" == "true" ]]; then
        cat > /etc/systemd/system/gost-node.service << EOF
[Unit]
Description=GOST Panel Node Agent
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/gost-agent -panel $PANEL_URL -token $TOKEN
Restart=always
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
    else
        cat > /etc/systemd/system/gost-node.service << EOF
[Unit]
Description=GOST Tunnel Node
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/gost -C /etc/gost/gost.yml
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
    fi

    systemctl daemon-reload
    systemctl enable gost-node
    systemctl start gost-node
}

# 创建 SysVinit 服务
create_sysvinit_service() {
    local use_agent=$1

    if [[ "$use_agent" == "true" ]]; then
        cat > /etc/init.d/gost-node << 'EOF'
#!/bin/sh
### BEGIN INIT INFO
# Provides:          gost-node
# Required-Start:    $network $remote_fs
# Required-Stop:     $network $remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Description:       GOST Panel Node Agent
### END INIT INFO

DAEMON="INSTALL_DIR/gost-agent"
DAEMON_ARGS="-panel PANEL_URL -token TOKEN"
PIDFILE="/var/run/gost-node.pid"

start() {
    echo "Starting gost-node..."
    start-stop-daemon --start --background --make-pidfile --pidfile $PIDFILE --exec $DAEMON -- $DAEMON_ARGS
}

stop() {
    echo "Stopping gost-node..."
    start-stop-daemon --stop --pidfile $PIDFILE
    rm -f $PIDFILE
}

case "$1" in
    start) start ;;
    stop) stop ;;
    restart) stop; start ;;
    *) echo "Usage: $0 {start|stop|restart}"; exit 1 ;;
esac
EOF
    else
        cat > /etc/init.d/gost-node << 'EOF'
#!/bin/sh
### BEGIN INIT INFO
# Provides:          gost-node
# Required-Start:    $network $remote_fs
# Required-Stop:     $network $remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Description:       GOST Tunnel Node
### END INIT INFO

DAEMON="/usr/local/bin/gost"
DAEMON_ARGS="-C /etc/gost/gost.yml"
PIDFILE="/var/run/gost-node.pid"

start() {
    echo "Starting gost-node..."
    start-stop-daemon --start --background --make-pidfile --pidfile $PIDFILE --exec $DAEMON -- $DAEMON_ARGS
}

stop() {
    echo "Stopping gost-node..."
    start-stop-daemon --stop --pidfile $PIDFILE
    rm -f $PIDFILE
}

case "$1" in
    start) start ;;
    stop) stop ;;
    restart) stop; start ;;
    *) echo "Usage: $0 {start|stop|restart}"; exit 1 ;;
esac
EOF
    fi

    # 替换变量
    sed -i "s|INSTALL_DIR|$INSTALL_DIR|g" /etc/init.d/gost-node
    sed -i "s|PANEL_URL|$PANEL_URL|g" /etc/init.d/gost-node
    sed -i "s|TOKEN|$TOKEN|g" /etc/init.d/gost-node

    chmod +x /etc/init.d/gost-node
    update-rc.d gost-node defaults 2>/dev/null || true
    /etc/init.d/gost-node start
}

# 创建 OpenWrt procd 服务
create_procd_service() {
    local use_agent=$1

    if [[ "$use_agent" == "true" ]]; then
        cat > /etc/init.d/gost-node << EOF
#!/bin/sh /etc/rc.common

START=99
STOP=10
USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command $INSTALL_DIR/gost-agent -panel $PANEL_URL -token $TOKEN
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
EOF
    else
        cat > /etc/init.d/gost-node << EOF
#!/bin/sh /etc/rc.common

START=99
STOP=10
USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command /usr/local/bin/gost -C /etc/gost/gost.yml
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
EOF
    fi

    chmod +x /etc/init.d/gost-node
    /etc/init.d/gost-node enable
    /etc/init.d/gost-node start
}

# 创建 OpenRC 服务
create_openrc_service() {
    local use_agent=$1

    if [[ "$use_agent" == "true" ]]; then
        cat > /etc/init.d/gost-node << EOF
#!/sbin/openrc-run

name="gost-node"
description="GOST Panel Node Agent"
command="$INSTALL_DIR/gost-agent"
command_args="-panel $PANEL_URL -token $TOKEN"
command_background="yes"
pidfile="/var/run/gost-node.pid"

depend() {
    need net
    after firewall
}
EOF
    else
        cat > /etc/init.d/gost-node << EOF
#!/sbin/openrc-run

name="gost-node"
description="GOST Tunnel Node"
command="/usr/local/bin/gost"
command_args="-C /etc/gost/gost.yml"
command_background="yes"
pidfile="/var/run/gost-node.pid"

depend() {
    need net
    after firewall
}
EOF
    fi

    chmod +x /etc/init.d/gost-node
    rc-update add gost-node default
    rc-service gost-node start
}

# 安装服务
install_service() {
    local use_agent=$1
    log_info "[4/5] Installing service ($INIT_SYSTEM)..."

    case $INIT_SYSTEM in
        systemd)
            create_systemd_service "$use_agent"
            ;;
        sysvinit)
            create_sysvinit_service "$use_agent"
            ;;
        procd)
            create_procd_service "$use_agent"
            ;;
        openrc)
            create_openrc_service "$use_agent"
            ;;
        *)
            log_warn "Unknown init system, creating startup script only"
            cat > "$INSTALL_DIR/start.sh" << EOF
#!/bin/bash
cd $INSTALL_DIR
nohup $INSTALL_DIR/gost-agent -panel $PANEL_URL -token $TOKEN > /var/log/gost-node.log 2>&1 &
EOF
            chmod +x "$INSTALL_DIR/start.sh"
            log_warn "Run '$INSTALL_DIR/start.sh' to start manually"
            ;;
    esac
}

# 主流程
main() {
    install_deps
    install_gost

    local use_agent="false"
    if download_agent; then
        use_agent="true"
    fi

    download_config
    install_service "$use_agent"

    log_info "[5/5] Verifying installation..."

    echo ""
    echo "========================================"
    echo "    Installation Complete!"
    echo "========================================"
    echo ""

    case $INIT_SYSTEM in
        systemd)
            echo "Service status:"
            systemctl status gost-node --no-pager || true
            echo ""
            echo "Commands:"
            echo "  systemctl status gost-node   - Check status"
            echo "  systemctl restart gost-node  - Restart"
            echo "  journalctl -u gost-node -f   - View logs"
            ;;
        procd)
            echo "Commands:"
            echo "  /etc/init.d/gost-node status  - Check status"
            echo "  /etc/init.d/gost-node restart - Restart"
            echo "  logread -f                    - View logs"
            ;;
        *)
            echo "Commands:"
            echo "  /etc/init.d/gost-node status  - Check status"
            echo "  /etc/init.d/gost-node restart - Restart"
            ;;
    esac
}

main
