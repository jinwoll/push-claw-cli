#!/bin/sh
# 推送虾 CLI 一键安装脚本（Linux / macOS / WSL / Git Bash）
# 用法: curl -fsSL https://raw.githubusercontent.com/jinwoll/push-claw-cli/main/install.sh | sh
set -e

# ===== 配置 =====
# GitHub Releases：/releases/latest/download/ 下需包含各平台二进制与 checksums.sha256（由 GoReleaser 生成）
RELEASES_BASE="${MINIXIA_BINARY_URL:-https://github.com/jinwoll/push-claw-cli/releases/latest/download}"
INSTALL_DIR="${MINIXIA_INSTALL_DIR:-}"
VERSION="${MINIXIA_VERSION:-latest}"
BINARY_NAME="push-claw"

# ===== 颜色输出 =====
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { printf "${CYAN}ℹ ${NC}%s\n" "$1"; }
success() { printf "${GREEN}✅ ${NC}%s\n" "$1"; }
warn()    { printf "${YELLOW}⚠ ${NC}%s\n" "$1"; }
error()   { printf "${RED}❌ ${NC}%s\n" "$1" >&2; exit 1; }

# ===== 检测下载工具 =====
detect_downloader() {
    if command -v curl >/dev/null 2>&1; then
        DOWNLOADER="curl"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOADER="wget"
    else
        error "需要 curl 或 wget，请先安装其中一个。"
    fi
}

download() {
    local url="$1" dest="$2"
    if [ "$DOWNLOADER" = "curl" ]; then
        curl -fsSL "$url" -o "$dest"
    else
        wget -qO "$dest" "$url"
    fi
}

# ===== 检测操作系统 =====
detect_os() {
    local uname_s
    uname_s=$(uname -s)
    case "$uname_s" in
        Linux*)   OS="linux" ;;
        Darwin*)  OS="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
        *)        error "不支持的操作系统: $uname_s" ;;
    esac
}

# ===== 检测 CPU 架构 =====
detect_arch() {
    local uname_m
    uname_m=$(uname -m)
    case "$uname_m" in
        x86_64|amd64)   ARCH="x86_64" ;;
        arm64|aarch64)  ARCH="arm64" ;;
        *)              error "不支持的架构: $uname_m" ;;
    esac
}

# ===== 确定安装路径 =====
determine_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        return
    fi

    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    elif [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    elif command -v sudo >/dev/null 2>&1; then
        INSTALL_DIR="/usr/local/bin"
        NEED_SUDO=true
    else
        INSTALL_DIR="$HOME/.local/bin"
    fi
}

# ===== 主流程 =====
main() {
    info "🦐 推送虾 CLI 安装程序"
    echo ""

    detect_downloader
    detect_os
    detect_arch

    info "系统: ${OS}/${ARCH}"

    # 构造下载文件名和 URL
    EXT=""
    if [ "$OS" = "windows" ]; then
        EXT=".exe"
    fi
    FILENAME="${BINARY_NAME}-${OS}-${ARCH}${EXT}"
    DOWNLOAD_URL="${RELEASES_BASE}/${FILENAME}"
    CHECKSUM_URL="${RELEASES_BASE}/checksums.sha256"

    # 创建临时目录
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # 下载二进制
    info "正在下载 ${FILENAME}…"
    download "$DOWNLOAD_URL" "$TMP_DIR/$FILENAME"

    # 下载并校验 SHA256（如果可用）
    if download "$CHECKSUM_URL" "$TMP_DIR/checksums.sha256" 2>/dev/null; then
        info "正在校验完整性…"
        EXPECTED=$(grep "$FILENAME" "$TMP_DIR/checksums.sha256" | awk '{print $1}')
        if [ -n "$EXPECTED" ]; then
            if command -v sha256sum >/dev/null 2>&1; then
                ACTUAL=$(sha256sum "$TMP_DIR/$FILENAME" | awk '{print $1}')
            elif command -v shasum >/dev/null 2>&1; then
                ACTUAL=$(shasum -a 256 "$TMP_DIR/$FILENAME" | awk '{print $1}')
            fi
            if [ -n "$ACTUAL" ] && [ "$EXPECTED" != "$ACTUAL" ]; then
                error "SHA256 校验失败！文件可能已损坏。"
            fi
            success "校验通过"
        fi
    else
        warn "跳过校验（校验和文件不可用）"
    fi

    # 确定安装路径并安装
    determine_install_dir
    mkdir -p "$INSTALL_DIR"

    DEST="$INSTALL_DIR/$BINARY_NAME$EXT"
    info "安装到 ${DEST}"

    if [ "${NEED_SUDO:-}" = "true" ]; then
        sudo mv "$TMP_DIR/$FILENAME" "$DEST"
        sudo chmod +x "$DEST"
    else
        mv "$TMP_DIR/$FILENAME" "$DEST"
        chmod +x "$DEST"
    fi

    # 验证安装
    if "$DEST" --version >/dev/null 2>&1; then
        success "安装成功！"
    else
        warn "二进制已放置到 ${DEST}，但验证失败，请检查。"
    fi

    # 检查 PATH
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *)
            echo ""
            warn "$INSTALL_DIR 不在 PATH 中，请添加到你的 shell 配置文件："
            echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
            echo ""
            ;;
    esac

    echo ""
    success "🚀 下一步：运行 push-claw init 开始配置"
}

main "$@"
