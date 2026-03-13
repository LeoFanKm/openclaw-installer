# OpenClaw Installer

面向 Windows 用户的 OpenClaw 一键安装器发布仓库。普通用户不需要先折腾命令行，直接从落地页或 Release 下载 EXE 即可开始安装。

- 一键安装页: https://clawpond.com/openclaw
- GitHub Release: https://github.com/LeoFanKm/openclaw-installer/releases/tag/v0.1.4
- Windows EXE: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/openclaw-windows-oneclick-2026.3.9.exe
- Checksums: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/checksums.txt
- 上游源码: https://github.com/openclaw/openclaw

## 从哪里开始

如果你是普通用户，优先访问一键安装页：

https://clawpond.com/openclaw

页面里已经整理好了：

- Windows 一键安装 EXE
- SHA256 校验文件
- 中文安装与恢复说明
- 低配设备预检建议
- 人工协助部署入口

## 这个仓库里有什么

- `main.go`: 启动本地 installer server，并自动打开浏览器
- `frontend/`: 嵌入式安装器前端
- `scripts/`: 构建脚本
- `.github/workflows/release.yml`: tag 发布流程

## 推荐使用方式

1. 打开 https://clawpond.com/openclaw
2. 下载 `openclaw-windows-oneclick-2026.3.9.exe`
3. 先看页面上的设备建议和恢复说明
4. 安装完成后，按页面指引继续连接渠道或提交协助部署

## Release 资产说明

当前推荐资产：

- `openclaw-windows-oneclick-2026.3.9.exe`
- `openclaw-windows-bundle-2026.3.9.zip`
- `openclaw-2026.3.9.tgz`
- `checksums.txt`

说明：

- `.exe` 适合普通用户，双击即可开始
- `.zip` 适合需要手动检查 bundle 内容或保留备用路径的用户
- `.tgz` 是修复后的 OpenClaw 包
- `checksums.txt` 用于校验下载文件来源和完整性

## 常见问题

### 1. 低配电脑能不能装？

可以尝试，但更建议先看一键安装页上的预检和最低配置说明。启动阶段会有短时 CPU 抬升，低于推荐值的机器更容易卡顿。

### 2. 启动后提示 `gateway token missing` 怎么办？

优先看 release 里的中文安装与恢复说明。大多数情况不是网关本身没 token，而是 Control UI 没带上 token。

### 3. 我不想自己排错怎么办？

直接使用一键安装页上的人工协助部署入口：

https://clawpond.com/openclaw#assisted

## 微信群

如果你需要进群交流或找人协助排障，可以扫码加入微信群：

<img src="./docs/wechat-qr-zh-v3.jpg" alt="OpenClaw 微信群二维码" width="280" />
