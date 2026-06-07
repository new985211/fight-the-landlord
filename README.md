<div align="center">
    <img src="https://raw.githubusercontent.com/palemoky/fight-the-landlord/main/docs/logo.png" alt="Logo" height="100px" />

# 🎮 欢乐斗地主

**一个真正公平的斗地主游戏 - 无控牌、无算法操控、纯粹的运气与技巧**

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

</div>

## 💡 项目初衷

- ✅ **真随机发牌**：每局洗牌完全随机，无任何控牌算法
- ✅ **公平匹配**：纯随机或房间匹配，不考虑胜率、段位
- ✅ **开源透明**：所有代码公开
- ✅ **无内购无广告**：纯粹的游戏体验

> **核心理念**：斗地主应该是运气与技巧的博弈，而不是算法与钱包的较量。

## 🍴 Fork 改进

本项目 Fork 自 [palemoky/fight-the-landlord](https://github.com/palemoky/fight-the-landlord)，在原项目基础上做了以下改进：

### 🔧 纯静态编译，零 glibc 依赖

所有二进制均使用 `CGO_ENABLED=0` 静态编译，**不依赖任何系统 glibc 版本**。无论目标机器是 glibc 2.17（CentOS 7）还是 glibc 2.34+（新版发行版），均可直接运行，彻底告别 `GLIBC_2.34 not found` 等问题。

### 📦 三种独立构建模式

项目提供三种独立的构建入口，按需选择：

| 构建产物 | 入口 | 说明 |
|----------|------|------|
| `ddz`（全合一） | `cmd/allinone/` | 内嵌 Redis + 服务端 + 客户端，一键启动 |
| `server` | `cmd/server/` | 独立服务端，用于部署专用服务器 |
| `client` | `cmd/client/` | 独立客户端，连接远程服务器 |

全合一二进制适合本地开箱即玩；独立服务端/客户端适合生产部署和局域网对战。

### 🌍 多平台交叉编译

支持 Linux (amd64/arm64)、Windows (amd64)、macOS (Intel/Apple Silicon) 的交叉编译，详见[编译章节](#-源码编译)。

## 🚀 快速开始（全合一二进制）

项目提供了一个**全合一二进制文件**，将服务端、客户端、Redis 打包在一起，无需安装任何依赖即可运行。

### 下载预编译二进制

从 [GitHub Releases](https://github.com/new985211/fight-the-landlord/releases) 下载对应架构的二进制文件，或按下方说明自行编译。

如果预编译二进制提示 glibc 版本不满足，请使用下方的静态编译方式自行构建。

### 启动游戏

```bash
# 一键启动（服务端 + 客户端 + Redis 全部自动运行）
./ddz

# 仅启动服务端（守护进程，供局域网其他玩家连接）
./ddz -mode server

# 仅启动客户端（连接远程服务器）
./ddz -mode client -server 192.168.1.100:1780

# 查看版本
./ddz -version
```

启动后使用键盘操作，常用按键：

| 按键 | 功能 |
| ---- | ---- |
| M | 开关音乐（默认静音） |
| C | 开关记牌器（默认关闭） |
| P | Pass（不出） |
| H | 帮助 |
| 10 / T | 出 10 |
| B | 小王（Black Joker） |
| R | 大王（Red Joker） |
| Esc | 返回上一页 |

### 牌型一览

```
单张: 3, K, 2
对子: 33, KK
三张: 333
三带一: 3334
三带二: 33344
顺子: 34567 (5张+)
连对: 334455 (3对+)
飞机: 333444 (两个连三+)
飞机带单: 33344456
飞机带对: 3334445566
四带二: 333345
四带两对: 33334455
炸弹: 3333
王炸: 小王大王
```

## 🔧 源码编译

### 环境要求

- Go 1.25+（如系统 Go 版本较低，可调整 `go.mod` 中的版本号）
- 纯静态编译无需 CGO，不依赖任何系统库

### 编译当前架构（amd64）

```bash
# 全合一二进制（推荐）
CGO_ENABLED=0 go build -tags ci -ldflags "-s -w -X main.version=v0.5.3" -o ddz ./cmd/allinone/

# 单独编译服务端
go build -trimpath -ldflags="-w -s" -o server ./cmd/server/

# 单独编译客户端
go build -trimpath -ldflags="-w -s" -o client ./cmd/client/
```

编译参数说明：
- `CGO_ENABLED=0`：禁用 CGO，生成纯静态二进制，无 glibc 版本依赖
- `-tags ci`：使用 noop 音效实现（避免 beep/ALSA 的 CGO 依赖），音乐功能不可用但不影响游戏
- `-ldflags "-s -w"`：去除调试信息，减小二进制体积
- `-X main.version=v0.5.3`：注入版本号

### 交叉编译 ARM64

```bash
# 全合一二进制
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -tags ci -ldflags "-s -w -X main.version=v0.5.3" \
  -o ddz-linux-arm64 ./cmd/allinone/

# 单独编译服务端
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" \
  -o server-linux-arm64 ./cmd/server/

# 单独编译客户端
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -tags ci -ldflags="-s -w -X main.version=v0.5.3" \
  -o client-linux-arm64 ./cmd/client/
```

ARM64 二进制适用于树莓派、ARM 云服务器、Apple Silicon Mac（Linux 环境）等。

### 交叉编译 Windows

```bash
# 全合一二进制
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -tags ci -ldflags="-s -w -X main.version=v0.5.3" \
  -o ddz.exe ./cmd/allinone/
```

Windows 下为命令行程序，在 CMD 或 PowerShell 中运行，用法与 Linux 版一致。

### 交叉编译 macOS

```bash
# Intel Mac
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
  go build -tags ci -ldflags="-s -w" -o ddz-darwin-amd64 ./cmd/allinone/

# Apple Silicon
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
  go build -tags ci -ldflags="-s -w" -o ddz-darwin-arm64 ./cmd/allinone/
```

### 编译产物一览

| 产物 | 目标平台 | 类型 |
|------|----------|------|
| `ddz` | Linux amd64 | 全合一（推荐） |
| `ddz-linux-arm64` | Linux arm64 | 全合一 |
| `ddz.exe` | Windows amd64 | 全合一 |
| `ddz-darwin-amd64` | macOS Intel | 全合一 |
| `ddz-darwin-arm64` | macOS Apple Silicon | 全合一 |
| `server` / `client` | Linux amd64 | 独立部署 |
| `server-linux-arm64` / `client-linux-arm64` | Linux arm64 | 独立部署 |

## 🏠 局域网真人对战教程

### 准备工作

1. 确保所有玩家连接在**同一个局域网**（同一个路由器/Wi-Fi）
2. 选择一台机器作为**服务端**（建议用性能较好的机器，或一直开机的那台）
3. 将编译好的 `ddz` 二进制分发给所有玩家，或每台机器自行编译

### 第一步：启动服务端

在服务端机器上运行：

```bash
./ddz -mode server
```

启动后会看到类似输出：

```
📦 内嵌 Redis 已启动 (addr: 127.0.0.1:33599)
🤖 规则启发式机器人已启用（等待超时: 15s）
🔒 安全配置: 连接限制=10/s, 消息限制=20/s, 聊天限制=1/s, 最大连接数=10000
🎮 斗地主服务端启动中...
🚀 服务器启动在 ws://0.0.0.0:1780/ws (CPU核心数: 16)
```

记下服务端机器的局域网 IP 地址：

```bash
# Linux / macOS
hostname -I | awk '{print $1}'
# 或者
ip addr show | grep 'inet ' | grep -v 127.0.0.1

# Windows
ipconfig | findstr "IPv4"
```

例如服务端 IP 是 `192.168.3.48`。

可以用浏览器访问 `http://192.168.3.48:1780/` 确认服务端运行正常（会显示服务状态页）。

### 第二步：客户端连接

在其他玩家的机器上运行：

```bash
./ddz -mode client -server 192.168.3.48:1780
```

> 将 `192.168.3.48` 替换为你服务端的实际 IP。

客户端连接成功后会进入游戏大厅。

### 第三步：开始对战

**方式一：快速匹配（随机匹配）**

在大厅界面选择"快速匹配"，系统会自动将排队玩家组成一桌。如果等待超时（默认 15 秒）人数不足 3 人，机器人会自动填充空位。

**方式二：创建房间（好友对战）**

1. 一名玩家选择"创建房间"，获得 6 位数字房间号
2. 其他玩家选择"加入房间"，输入房间号
3. 所有玩家进入房间后点击"准备"
4. 3 人全部准备后自动开始游戏

### 局域网模式说明

| 特性 | 说明 |
|------|------|
| 机器人填充 | 默认开启，等待 15 秒后自动补位 |
| 全合一客户端 | 局域网内的客户端可以用 `-mode client` 仅启动客户端 |
| 端口 | 默认 1780，可通过 `config.yaml` 修改 |
| 防火墙 | 如果客户端无法连接，检查服务端防火墙是否放行 1780 端口 |

### 关闭机器人（纯真人）

如果想禁止机器人、仅允许真人玩家，编辑 `config.yaml`：

```yaml
bot:
  enabled: false
```

然后重启服务端。此时快速匹配需要凑满 3 个真人；建议使用房间模式手动组队。

### 防火墙设置

客户端连不上时，在服务端机器上放行端口：

```bash
# Linux (ufw)
sudo ufw allow 1780/tcp

# Linux (firewalld)
sudo firewall-cmd --add-port=1780/tcp --permanent && sudo firewall-cmd --reload

# Linux (iptables)
sudo iptables -A INPUT -p tcp --dport 1780 -j ACCEPT
```

Windows 需要在"Windows Defender 防火墙 → 高级设置 → 入站规则"中添加 1780 端口。

### 服务端设为开机自启

**Linux (systemd)**：

```bash
sudo tee /etc/systemd/system/ddz-server.service << 'EOF'
[Unit]
Description=斗地主游戏服务端
After=network.target

[Service]
Type=simple
ExecStart=/path/to/ddz -mode server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now ddz-server
```

### 常见问题

**Q: 客户端提示"连接失败"？**
- 检查服务端是否启动
- 检查 IP 地址是否正确
- 检查防火墙设置
- 用 `curl http://<服务端IP>:1780/health` 测试连通性（应返回 `OK`）

**Q: 预编译二进制提示 `GLIBC_2.34 not found`？**
- 说明系统 glibc 版本过低（常见于麒麟 V10、CentOS 7 等）
- 按上方"源码编译"步骤在本机重新编译，静态链接无此问题

**Q: 如何同时在一台机器上运行多个客户端？**
- 打开多个终端窗口，每个窗口运行 `./ddz -mode client -server 127.0.0.1:1780`

## 🐳 Docker 部署

```bash
# 下载配置文件
curl -fsSL https://raw.githubusercontent.com/new985211/fight-the-landlord/main/docker-compose.yml -o docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/new985211/fight-the-landlord/main/.env.example -o .env

# 启动
docker compose up -d

# 停止
docker compose down
```

## 🎲 配置参考

配置文件为 `config.yaml`，主要选项：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `server.port` | 1780 | WebSocket 服务端口 |
| `server.min_client_version` | v0.5.3 | 最低客户端版本要求 |
| `game.turn_timeout` | 30 | 出牌超时（秒） |
| `game.bid_timeout` | 15 | 叫地主超时（秒） |
| `game.room_timeout` | 10 | 房间空闲超时（分钟） |
| `bot.enabled` | true | 是否启用机器人填充 |
| `bot.douzero_enabled` | true | 是否使用 DouZero AI（需 Python 服务） |

## 📁 项目结构

```
cmd/
  allinone/   # 全合一二进制入口（服务端 + 客户端 + 内嵌 Redis）
  server/     # 服务端入口
  client/     # 客户端入口
internal/
  server/     # WebSocket 服务端、连接管理、安全组件
  game/       # 游戏引擎：卡牌(card)、规则(rule)、房间(room)、匹配(match)
  transport/  # 客户端 WebSocket 传输层（含断线重连）
  protocol/   # 消息类型定义、Protobuf 编解码
  bot/        # 机器人（启发式 + DouZero 神经网络）
  ui/         # Bubble Tea TUI 界面
  config/     # YAML 配置加载
  sound/      # 音效（beep 库，编译时可选禁用）
douzero/      # Python DouZero AI 推理服务
```

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

---

<div align="center">

**让斗地主回归纯粹 - 无控牌，真公平**

Forked from [palemoky/fight-the-landlord](https://github.com/palemoky/fight-the-landlord)

</div>
