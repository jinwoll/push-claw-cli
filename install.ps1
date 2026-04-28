# 推送虾 CLI 一键安装脚本（Windows PowerShell）
# 用法: irm https://raw.githubusercontent.com/jinwoll/push-claw-cli/main/install.ps1 | iex
$ErrorActionPreference = "Stop"

# ===== 配置 =====
$ReleasesBase = if ($env:MINIXIA_BINARY_URL) { $env:MINIXIA_BINARY_URL } else { "https://github.com/jinwoll/push-claw-cli/releases/latest/download" }
$InstallDir = if ($env:MINIXIA_INSTALL_DIR) { $env:MINIXIA_INSTALL_DIR } else { "$env:LOCALAPPDATA\push-claw" }
$BinaryName = "push-claw.exe"

function Write-Info    { param($msg) Write-Host "ℹ $msg" -ForegroundColor Cyan }
function Write-Success { param($msg) Write-Host "✅ $msg" -ForegroundColor Green }
function Write-Warn    { param($msg) Write-Host "⚠ $msg" -ForegroundColor Yellow }
function Write-Err     { param($msg) Write-Host "❌ $msg" -ForegroundColor Red; exit 1 }

# ===== 检测 PowerShell 版本 =====
if ($PSVersionTable.PSVersion.Major -lt 5) {
    Write-Err "需要 PowerShell 5.1 或更高版本。"
}

# ===== 检测架构 =====
Write-Info "🦐 推送虾 CLI 安装程序"
Write-Host ""

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64"  { "x86_64" }
    "ARM64"  { "arm64" }
    default  { Write-Err "不支持的架构: $env:PROCESSOR_ARCHITECTURE" }
}
Write-Info "系统: windows/$Arch"

# ===== 构造下载 URL =====
$FileName = "push-claw-windows-${Arch}.exe"
$DownloadUrl = "$ReleasesBase/$FileName"
$ChecksumUrl = "$ReleasesBase/checksums.sha256"

# ===== 下载到临时目录 =====
$TmpDir = Join-Path $env:TEMP "push-claw-install-$(Get-Random)"
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null
$TmpFile = Join-Path $TmpDir $FileName

Write-Info "正在下载 $FileName…"
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile -UseBasicParsing
} catch {
    Write-Err "下载失败: $_"
}

# ===== SHA256 校验 =====
$ChecksumFile = Join-Path $TmpDir "checksums.sha256"
try {
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumFile -UseBasicParsing
    Write-Info "正在校验完整性…"
    $ExpectedHash = (Get-Content $ChecksumFile | Where-Object { $_ -match $FileName } | ForEach-Object { ($_ -split '\s+')[0] })
    if ($ExpectedHash) {
        $ActualHash = (Get-FileHash -Path $TmpFile -Algorithm SHA256).Hash.ToLower()
        if ($ExpectedHash.ToLower() -ne $ActualHash) {
            Write-Err "SHA256 校验失败！文件可能已损坏。"
        }
        Write-Success "校验通过"
    }
} catch {
    Write-Warn "跳过校验（校验和文件不可用）"
}

# ===== 安装到目标目录 =====
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$Dest = Join-Path $InstallDir $BinaryName
Write-Info "安装到 $Dest"
Copy-Item -Path $TmpFile -Destination $Dest -Force

# ===== 添加到用户 PATH =====
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Info "正在添加 $InstallDir 到用户 PATH…"
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Success "已添加到 PATH（新终端窗口生效）"
}

# ===== 验证安装 =====
try {
    $VersionOutput = & $Dest --version 2>&1
    Write-Success "安装成功！ $VersionOutput"
} catch {
    Write-Warn "二进制已放置到 $Dest，但验证失败，请检查。"
}

# ===== 清理临时文件 =====
Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-Success "🚀 下一步：运行 push-claw init 开始配置"
