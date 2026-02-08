#!/bin/bash
# GOST Panel 客户端安装脚本
# 支持: Linux (amd64, arm64, armv7, armv6, mips, mipsle, mips64)
# 用法: curl -fsSL URL | bash -s -- -p PANEL_URL -t TOKEN
#   或: wget -qO- URL | bash -s -- -p PANEL_URL -t TOKEN

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

# HTTP 下载 (自动检测 curl/wget)
dl() {
    local url="$1" output="$2"
    if command -v curl &>/dev/null; then
        [ -n "$output" ] && curl -fsSL "$url" -o "$output" || curl -fsSL "$url"
    elif command -v wget &>/dev/null; then
        [ -n "$output" ] && wget -qO "$output" "$url" || wget -qO- "$url"
    else
        log_error "curl and wget not found, please install one of them"
        exit 1
    fi
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--panel) PANEL_URL="$2"; shift 2 ;;
        -t|--token) TOKEN="$2"; shift 2 ;;
        -a|--arch) FORCE_ARCH="$2"; shift 2 ;;
        -h|--help)
            echo "GOST Panel Client Installer"
            echo ""
            echo "Usage: $0 -p <panel_url> -t <token> [-a <arch>]"
            echo ""
            echo "Options:"
            echo "  -p, --panel   Panel URL (e.g., http://panel.example.com:8080)"
            echo "  -t, --token   Client token from panel"
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
echo "   GOST Panel Client Installer"
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
        x86_64|amd64) gost_arch="amd64" ;;
        aarch64|arm64) gost_arch="arm64" ;;
        armv7l|armv7) gost_arch="armv7" ;;
        armv6l|armv6) gost_arch="armv6" ;;
        armv5*) gost_arch="armv5" ;;
        mips)
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
        i386|i686) gost_arch="386" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac

    echo "$gost_arch"
}

GOST_ARCH=$(detect_arch)
log_info "Detected architecture: $GOST_ARCH"

# 检测 init 系统
detect_init_system() {
    if command -v systemctl &> /dev/null && systemctl --version &> /dev/null 2>&1; then
        echo "systemd"
    elif [[ -f /etc/init.d/rcS ]] || command -v update-rc.d &> /dev/null; then
        echo "sysvinit"
    elif [[ -f /etc/rc.common ]]; then
        echo "procd"
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
    log_info "[1/4] Installing GOST..."

    if command -v gost &> /dev/null; then
        log_info "GOST already installed: $(which gost)"
        return
    fi

    local gost_url="https://github.com/go-gost/gost/releases/download/v${GOST_VERSION}/gost_${GOST_VERSION}_linux_${GOST_ARCH}.tar.gz"
    log_info "Downloading GOST..."

    dl "$gost_url" /tmp/gost.tar.gz
    mkdir -p /tmp/gost-extract
    tar -xzf /tmp/gost.tar.gz -C /tmp/gost-extract
    mv /tmp/gost-extract/gost /usr/local/bin/
    chmod +x /usr/local/bin/gost
    rm -rf /tmp/gost.tar.gz /tmp/gost-extract

    log_info "GOST installed to /usr/local/bin/gost"
}

# 下载配置
download_config() {
    log_info "[2/4] Downloading config..."

    mkdir -p /etc/gost
    dl "$PANEL_URL/agent/config/$TOKEN" /etc/gost/client.yml
    log_info "Config saved to /etc/gost/client.yml"
}

# 创建服务
install_service() {
    log_info "[3/4] Installing service..."

    case $INIT_SYSTEM in
        systemd)
            cat > /etc/systemd/system/gost-client.service << EOF
[Unit]
Description=GOST Panel Client
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/gost -C /etc/gost/client.yml
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
            systemctl daemon-reload
            systemctl enable gost-client
            systemctl start gost-client
            ;;

        procd)
            cat > /etc/init.d/gost-client << EOF
#!/bin/sh /etc/rc.common

START=99
STOP=10
USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command /usr/local/bin/gost -C /etc/gost/client.yml
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
EOF
            chmod +x /etc/init.d/gost-client
            /etc/init.d/gost-client enable
            /etc/init.d/gost-client start
            ;;

        openrc)
            cat > /etc/init.d/gost-client << EOF
#!/sbin/openrc-run

name="gost-client"
description="GOST Panel Client"
command="/usr/local/bin/gost"
command_args="-C /etc/gost/client.yml"
command_background="yes"
pidfile="/var/run/gost-client.pid"

depend() {
    need net
    after firewall
}
EOF
            chmod +x /etc/init.d/gost-client
            rc-update add gost-client default
            rc-service gost-client start
            ;;

        sysvinit)
            cat > /etc/init.d/gost-client << 'EOF'
#!/bin/sh
### BEGIN INIT INFO
# Provides:          gost-client
# Required-Start:    $network $remote_fs
# Required-Stop:     $network $remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Description:       GOST Panel Client
### END INIT INFO

DAEMON="/usr/local/bin/gost"
DAEMON_ARGS="-C /etc/gost/client.yml"
PIDFILE="/var/run/gost-client.pid"

start() {
    echo "Starting gost-client..."
    start-stop-daemon --start --background --make-pidfile --pidfile $PIDFILE --exec $DAEMON -- $DAEMON_ARGS
}

stop() {
    echo "Stopping gost-client..."
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
            chmod +x /etc/init.d/gost-client
            update-rc.d gost-client defaults 2>/dev/null || true
            /etc/init.d/gost-client start
            ;;

        *)
            log_warn "Unknown init system, creating startup script"
            mkdir -p "$INSTALL_DIR"
            cat > "$INSTALL_DIR/start-client.sh" << EOF
#!/bin/bash
nohup /usr/local/bin/gost -C /etc/gost/client.yml > /var/log/gost-client.log 2>&1 &
EOF
            chmod +x "$INSTALL_DIR/start-client.sh"
            ;;
    esac
}

# 安装心跳
install_heartbeat() {
    log_info "[3.5/4] Setting up heartbeat..."

    # 创建心跳脚本 (含自动卸载检测)
    cat > /etc/gost/heartbeat.sh << HEARTBEAT
#!/bin/bash
# GOST Client Heartbeat (auto-uninstall on 410 Gone)
HTTP_CODE=""

if command -v curl &>/dev/null; then
    HTTP_CODE=\$(curl -s -o /dev/null -w "%{http_code}" -X POST "${PANEL_URL}/agent/client-heartbeat/${TOKEN}" 2>/dev/null)
elif command -v wget &>/dev/null; then
    HTTP_CODE=\$(wget -S -q --post-data="" "${PANEL_URL}/agent/client-heartbeat/${TOKEN}" -O /dev/null 2>&1 | awk '/HTTP\//{print \$2}' | tail -1)
fi

# 410 Gone = client deleted from panel, auto-uninstall
if [ "\$HTTP_CODE" = "410" ]; then
    echo "[GOST] Client deleted from panel, auto-uninstalling..."

    # Stop and remove gost-client service
    if command -v systemctl &>/dev/null; then
        systemctl stop gost-client 2>/dev/null
        systemctl disable gost-client 2>/dev/null
        rm -f /etc/systemd/system/gost-client.service
        # Stop and remove heartbeat timer
        systemctl stop gost-heartbeat.timer 2>/dev/null
        systemctl disable gost-heartbeat.timer 2>/dev/null
        rm -f /etc/systemd/system/gost-heartbeat.service
        rm -f /etc/systemd/system/gost-heartbeat.timer
        systemctl daemon-reload
    elif [ -f /etc/init.d/gost-client ]; then
        /etc/init.d/gost-client stop 2>/dev/null
        if command -v update-rc.d &>/dev/null; then
            update-rc.d -f gost-client remove 2>/dev/null
        elif command -v rc-update &>/dev/null; then
            rc-update del gost-client default 2>/dev/null
        fi
        rm -f /etc/init.d/gost-client
    fi

    # Remove cron entry
    (crontab -l 2>/dev/null | grep -v "gost/heartbeat") | crontab - 2>/dev/null

    # Remove files
    rm -rf /etc/gost
    rm -f /usr/local/bin/gost

    echo "[GOST] Uninstall complete."
fi
HEARTBEAT
    chmod +x /etc/gost/heartbeat.sh

    case $INIT_SYSTEM in
        systemd)
            # systemd timer (每分钟心跳)
            cat > /etc/systemd/system/gost-heartbeat.service << 'EOF'
[Unit]
Description=GOST Client Heartbeat

[Service]
Type=oneshot
ExecStart=/etc/gost/heartbeat.sh
EOF

            cat > /etc/systemd/system/gost-heartbeat.timer << 'EOF'
[Unit]
Description=GOST Client Heartbeat Timer

[Timer]
OnBootSec=10s
OnUnitActiveSec=1m

[Install]
WantedBy=timers.target
EOF
            systemctl daemon-reload
            systemctl enable gost-heartbeat.timer
            systemctl start gost-heartbeat.timer
            ;;
        *)
            # cron fallback (每分钟心跳)
            (crontab -l 2>/dev/null | grep -v "gost/heartbeat"; echo "* * * * * /etc/gost/heartbeat.sh") | crontab -
            ;;
    esac

    # 发送首次心跳
    /etc/gost/heartbeat.sh
    log_info "Heartbeat configured"
}

# 显示连接信息
show_info() {
    log_info "[4/4] Extracting connection info..."

    # 尝试解析配置获取本地端口
    local local_port=$(grep -oP '(?<=addr: ":)\d+' /etc/gost/client.yml 2>/dev/null | head -1 || echo "38777")

    echo ""
    echo "========================================"
    echo "    Installation Complete!"
    echo "========================================"
    echo ""
    echo "Local SOCKS5 proxy: socks5://127.0.0.1:$local_port"
    echo ""

    case $INIT_SYSTEM in
        systemd)
            echo "Service status:"
            systemctl status gost-client --no-pager || true
            echo ""
            echo "Commands:"
            echo "  systemctl status gost-client   - Check status"
            echo "  systemctl restart gost-client  - Restart"
            echo "  journalctl -u gost-client -f   - View logs"
            ;;
        procd)
            echo "Commands:"
            echo "  /etc/init.d/gost-client status  - Check status"
            echo "  /etc/init.d/gost-client restart - Restart"
            ;;
        *)
            echo "Commands:"
            echo "  /etc/init.d/gost-client status  - Check status"
            echo "  /etc/init.d/gost-client restart - Restart"
            ;;
    esac
}

# 主流程
main() {
    install_gost
    download_config
    install_service
    install_heartbeat
    show_info
}

main
