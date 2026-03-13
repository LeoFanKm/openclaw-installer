# OpenClaw Installer

**[中文](#zh) | [English](#en) | [日本語](#ja) | [한국어](#ko)**

---

<a id="zh"></a>

## 中文

面向 Windows 用户的 OpenClaw 一键安装器发布仓库。普通用户不需要先折腾命令行，直接从落地页或 Release 下载 EXE 即可开始安装。

- 一键安装页: https://clawpond.com/openclaw
- GitHub Release: https://github.com/LeoFanKm/openclaw-installer/releases/tag/v0.1.4
- Windows EXE: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/openclaw-windows-oneclick-2026.3.9.exe
- Checksums: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/checksums.txt
- 上游源码: https://github.com/openclaw/openclaw

### 从哪里开始

如果你是普通用户，优先访问一键安装页：

https://clawpond.com/openclaw

页面里已经整理好了：

- Windows 一键安装 EXE
- SHA256 校验文件
- 中文安装与恢复说明
- 低配设备预检建议
- 人工协助部署入口

### 这个仓库里有什么

- `main.go`: 启动本地 installer server，并自动打开浏览器
- `frontend/`: 嵌入式安装器前端
- `scripts/`: 构建脚本
- `.github/workflows/release.yml`: tag 发布流程

### 推荐使用方式

1. 打开 https://clawpond.com/openclaw
2. 下载 `openclaw-windows-oneclick-2026.3.9.exe`
3. 先看页面上的设备建议和恢复说明
4. 安装完成后，按页面指引继续连接渠道或提交协助部署

### Release 资产说明

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

### 常见问题

#### 1. 低配电脑能不能装？

可以尝试，但更建议先看一键安装页上的预检和最低配置说明。启动阶段会有短时 CPU 抬升，低于推荐值的机器更容易卡顿。

#### 2. 启动后提示 `gateway token missing` 怎么办？

优先看 release 里的中文安装与恢复说明。大多数情况不是网关本身没 token，而是 Control UI 没带上 token。

#### 3. 我不想自己排错怎么办？

直接使用一键安装页上的人工协助部署入口：

https://clawpond.com/openclaw#assisted

### 微信群

如果你需要进群交流或找人协助排障，可以扫码加入微信群：

<img src="./docs/wechat-qr-zh-v3.jpg" alt="OpenClaw 微信群二维码" width="280" />

---

<a id="en"></a>

## English

One-click installer for OpenClaw local development environment on Windows. No command-line setup needed — download the EXE directly from the landing page or GitHub Release.

- Landing page: https://clawpond.com/openclaw
- GitHub Release: https://github.com/LeoFanKm/openclaw-installer/releases/tag/v0.1.4
- Windows EXE: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/openclaw-windows-oneclick-2026.3.9.exe
- Checksums: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/checksums.txt
- Upstream source: https://github.com/openclaw/openclaw

### Where to Start

For regular users, visit the one-click installation page first:

https://clawpond.com/openclaw

The page includes:

- Windows one-click installer EXE
- SHA256 checksum file
- Installation and recovery guide
- Pre-check recommendations for low-spec devices
- Assisted deployment entry point

### What's in This Repository

- `main.go`: Starts a local installer server and automatically opens the browser
- `frontend/`: Embedded installer frontend
- `scripts/`: Build scripts
- `.github/workflows/release.yml`: Tag-based release workflow

### Recommended Steps

1. Open https://clawpond.com/openclaw
2. Download `openclaw-windows-oneclick-2026.3.9.exe`
3. Read the device requirements and recovery guide on the page first
4. After installation, follow the page instructions to connect channels or submit an assisted deployment request

### Release Assets

Current recommended assets:

- `openclaw-windows-oneclick-2026.3.9.exe`
- `openclaw-windows-bundle-2026.3.9.zip`
- `openclaw-2026.3.9.tgz`
- `checksums.txt`

Notes:

- `.exe` — For regular users, just double-click to start
- `.zip` — For users who want to inspect the bundle contents manually or keep a backup path
- `.tgz` — Patched OpenClaw source package
- `checksums.txt` — Verify the integrity and source of downloaded files

### FAQ

#### 1. Can I install on a low-spec machine?

You can try, but we strongly recommend checking the pre-check and minimum requirements on the landing page first. CPU usage spikes briefly during startup, and machines below the recommended spec are more likely to stall.

#### 2. What if I see `gateway token missing` after launch?

Check the installation and recovery guide in the Release first. In most cases the gateway itself has a token — the issue is that the Control UI did not pass it through.

#### 3. I don't want to troubleshoot myself — what can I do?

Use the assisted deployment entry on the landing page:

https://clawpond.com/openclaw#assisted

### WeChat Community

Scan the QR code to join the WeChat support group:

<img src="./docs/wechat-qr-zh-v3.jpg" alt="OpenClaw WeChat QR Code" width="280" />

---

<a id="ja"></a>

## 日本語

Windows ユーザー向け OpenClaw ローカル開発環境のワンクリックインストーラーです。コマンドライン操作は不要 — ランディングページまたは GitHub Release から EXE を直接ダウンロードしてインストールできます。

- ワンクリックインストールページ: https://clawpond.com/openclaw
- GitHub Release: https://github.com/LeoFanKm/openclaw-installer/releases/tag/v0.1.4
- Windows EXE: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/openclaw-windows-oneclick-2026.3.9.exe
- チェックサム: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/checksums.txt
- アップストリームソース: https://github.com/openclaw/openclaw

### はじめに

一般ユーザーの方は、まずワンクリックインストールページをご覧ください：

https://clawpond.com/openclaw

ページには以下が用意されています：

- Windows ワンクリックインストーラー EXE
- SHA256 チェックサムファイル
- インストールと復旧ガイド
- 低スペック端末向けの事前確認アドバイス
- 有人サポートデプロイ申請入口

### このリポジトリの内容

- `main.go`: ローカルインストーラーサーバーを起動し、ブラウザを自動で開く
- `frontend/`: 埋め込みインストーラーフロントエンド
- `scripts/`: ビルドスクリプト
- `.github/workflows/release.yml`: タグベースのリリースワークフロー

### 推奨手順

1. https://clawpond.com/openclaw を開く
2. `openclaw-windows-oneclick-2026.3.9.exe` をダウンロード
3. ページの端末要件と復旧ガイドを先に確認する
4. インストール完了後、ページの指示に従いチャンネル接続またはサポートデプロイ申請へ進む

### リリースアセットの説明

現在推奨されているアセット：

- `openclaw-windows-oneclick-2026.3.9.exe`
- `openclaw-windows-bundle-2026.3.9.zip`
- `openclaw-2026.3.9.tgz`
- `checksums.txt`

説明：

- `.exe` — 一般ユーザー向け、ダブルクリックで開始
- `.zip` — バンドル内容を手動確認したい方、バックアップパスを保持したい方向け
- `.tgz` — 修正済み OpenClaw パッケージ
- `checksums.txt` — ダウンロードファイルの整合性と出所の検証用

### よくある質問

#### 1. 低スペックのPCにインストールできますか？

試すことはできますが、まずランディングページの事前確認と最低スペック要件をご確認ください。起動時に CPU 使用率が一時的に上昇するため、推奨スペック以下の端末では動作が不安定になる場合があります。

#### 2. 起動後に `gateway token missing` と表示された場合は？

まず Release 内のインストールと復旧ガイドをご確認ください。多くの場合、ゲートウェイ自体にトークンが存在しており、Control UI がトークンを渡していないことが原因です。

#### 3. 自分でトラブルシューティングしたくない場合は？

ランディングページの有人サポートデプロイ申請入口をご利用ください：

https://clawpond.com/openclaw#assisted

### WeChat コミュニティ

QR コードをスキャンして WeChat サポートグループに参加してください：

<img src="./docs/wechat-qr-zh-v3.jpg" alt="OpenClaw WeChat QRコード" width="280" />

---

<a id="ko"></a>

## 한국어

Windows 사용자를 위한 OpenClaw 로컬 개발 환경 원클릭 설치 프로그램입니다. 명령줄 작업 없이 랜딩 페이지 또는 GitHub Release에서 EXE를 직접 다운로드하여 설치할 수 있습니다.

- 원클릭 설치 페이지: https://clawpond.com/openclaw
- GitHub Release: https://github.com/LeoFanKm/openclaw-installer/releases/tag/v0.1.4
- Windows EXE: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/openclaw-windows-oneclick-2026.3.9.exe
- 체크섬: https://github.com/LeoFanKm/openclaw-installer/releases/download/v0.1.4/checksums.txt
- 업스트림 소스: https://github.com/openclaw/openclaw

### 시작하기

일반 사용자라면 먼저 원클릭 설치 페이지를 방문하세요：

https://clawpond.com/openclaw

페이지에는 다음이 준비되어 있습니다：

- Windows 원클릭 설치 EXE
- SHA256 체크섬 파일
- 설치 및 복구 가이드
- 저사양 기기 사전 점검 권장 사항
- 수동 지원 배포 신청 입구

### 이 저장소의 내용

- `main.go`: 로컬 설치 서버를 시작하고 브라우저를 자동으로 엽니다
- `frontend/`: 내장형 설치 프로그램 프론트엔드
- `scripts/`: 빌드 스크립트
- `.github/workflows/release.yml`: 태그 기반 릴리스 워크플로

### 권장 사용 방법

1. https://clawpond.com/openclaw 열기
2. `openclaw-windows-oneclick-2026.3.9.exe` 다운로드
3. 페이지의 기기 요구사항 및 복구 가이드를 먼저 확인
4. 설치 완료 후 페이지 안내에 따라 채널 연결 또는 지원 배포 신청 진행

### 릴리스 자산 설명

현재 권장 자산：

- `openclaw-windows-oneclick-2026.3.9.exe`
- `openclaw-windows-bundle-2026.3.9.zip`
- `openclaw-2026.3.9.tgz`
- `checksums.txt`

설명：

- `.exe` — 일반 사용자용, 더블클릭으로 시작
- `.zip` — 번들 내용을 수동으로 확인하거나 백업 경로를 유지하려는 사용자용
- `.tgz` — 패치된 OpenClaw 패키지
- `checksums.txt` — 다운로드 파일의 무결성 및 출처 검증용

### 자주 묻는 질문

#### 1. 저사양 PC에도 설치할 수 있나요?

시도는 가능하지만, 먼저 랜딩 페이지의 사전 점검 및 최소 사양 요구사항을 확인하는 것을 강력히 권장합니다. 시작 시 CPU 사용률이 일시적으로 상승하며, 권장 사양 이하의 기기에서는 속도 저하가 발생할 수 있습니다.

#### 2. 시작 후 `gateway token missing` 오류가 표시되면 어떻게 하나요?

Release의 설치 및 복구 가이드를 먼저 확인하세요. 대부분의 경우 게이트웨이 자체에는 토큰이 있지만 Control UI가 토큰을 전달하지 않은 것이 원인입니다.

#### 3. 직접 문제를 해결하고 싶지 않다면?

랜딩 페이지의 수동 지원 배포 신청 입구를 이용하세요：

https://clawpond.com/openclaw#assisted

### WeChat 커뮤니티

QR 코드를 스캔하여 WeChat 지원 그룹에 참여하세요：

<img src="./docs/wechat-qr-zh-v3.jpg" alt="OpenClaw WeChat QR 코드" width="280" />
