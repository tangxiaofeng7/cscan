# CSCAN

**分布式网络资产扫描平台** | Go-Zero + Vue3

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 特性

- **资产发现** - 端口扫描 (Naabu/Masscan)，端口服务识别（Nmap）
- **指纹识别** - Httpx + Wappalyzer + 自定义指纹引擎，3W+自定义指纹
- **漏洞检测** - Nuclei SDK引擎，800+自定义POC
- **Web 截图** - Chromedp / HTTPX 引擎
- **在线数据源** - FOFA / Hunter / Quake API 聚合搜索与导入
- **报告管理** - 任务报告生成，支持 Excel 导出
- **分布式架构** - Worker 节点水平扩展，支持多节点并行扫描
- **多工作空间** - 项目隔离，团队协作

## 快速开始

```bash
git clone https://github.com/tangxiaofeng7/cscan.git
cd cscan
# 直接启动
docker-compose up -d --build
# 指定多个worker
docker-compose up -d --build --scale cscan-worker=3
```

访问 `http://localhost:3000`，默认账号 `admin / 123456`


## 交流群
<img src="images/cscan.jpg" alt="交流群" width="200"/>

## 截图

#### 工作台
<img src="images/index.png" alt="工作台" width="600"/>

#### 任务管理
<img src="images/task.png" alt="任务管理" width="600"/>

#### 资产管理
<img src="images/asset.png" alt="资产管理" width="600"/>

#### 扫描日志
<img src="images/process.png" alt="扫描日志" width="600"/>

#### 查看报告
<img src="images/report.png" alt="查看报告" width="600"/>

#### POC管理
<img src="images/poc.png" alt="POC管理" width="600"/>

#### 指纹管理
<img src="images/finger.png" alt="指纹管理" width="600"/>

## 架构

```
Vue3 Web ──▶ API Server ──▶ MongoDB
                │
                ▼
              Redis
                │
                ▼
            RPC Server
                │
    ┌───────────┼───────────┐
    ▼           ▼           ▼
 Worker 1   Worker 2   Worker N
```

| 组件 | 技术栈 |
|------|--------|
| 后端框架 | Go-Zero, gRPC, JWT |
| 数据存储 | MongoDB, Redis |
| 前端框架 | Vue 3, Element Plus, Vite |
| 端口扫描 | Nmap, Masscan, Naabu |
| 指纹识别 | Httpx, Wappalyzer, 自定义引擎 |
| 漏洞扫描 | Nuclei |
| 截图引擎 | Httpx, Chromedp (Chromium) |

## 本地开发

```bash
# 1. 启动依赖服务（MongoDB + Redis）
docker-compose -f docker-compose.dev.yaml up -d

# 2. 启动 RPC 服务
go run rpc/task/task.go -f rpc/task/etc/task.yaml

# 3. 启动 API 服务
go run api/cscan.go -f api/etc/cscan.yaml

# 4. 启动 Worker（可启动多个）
go run cmd/worker/main.go -s localhost:9000 -r localhost:6379

# 5. 启动前端
cd web && npm install && npm run dev
```

### Worker 参数说明

```bash
go run cmd/worker/main.go [options]
  -s string    RPC 服务地址 (默认 "localhost:9000")
  -r string    Redis 地址 (默认 "localhost:6379")
  -rp string   Redis 密码 (默认 "")
  -n string    Worker 名称 (默认 主机名)
  -c int       并发数 (默认 5)
```

### 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| cscan-web | 3000 | Web 前端 |
| cscan-api | 8080 | API 服务 |
| cscan-rpc | 9000 | RPC 服务 |
| cscan-worker | - | 扫描节点 |
| mongodb | 27017 | 数据库 |
| redis | 6379 | 缓存/消息队列 |

### 扩展 Worker 节点

```bash
docker-compose up -d --scale cscan-worker=3
```

## 项目结构

```
├── api/            # HTTP API 服务
│   ├── internal/   # 内部实现
│   │   ├── handler/  # 请求处理器
│   │   ├── logic/    # 业务逻辑
│   │   ├── svc/      # 服务上下文
│   │   └── types/    # 类型定义
│   └── etc/        # 配置文件
├── rpc/            # gRPC 服务
│   └── task/       # 任务服务
├── worker/         # 扫描 Worker
├── scanner/        # 扫描器实现
│   ├── portscan.go   # 端口扫描
│   ├── fingerprint.go # 指纹识别
│   ├── nuclei.go     # 漏洞扫描
│   └── ...
├── model/          # 数据模型
├── scheduler/      # 任务调度器
├── onlineapi/      # FOFA/Hunter/Quake 集成
├── pkg/            # 公共包
│   ├── xerr/       # 错误码定义
│   └── response/   # 响应封装
├── web/            # Vue3 前端
├── docker/         # Docker 配置
└── poc/            # POC 文件
```

## 功能模块

### 任务管理
- 创建扫描任务，支持 IP/域名/CIDR 目标
- 任务配置模板，快速复用扫描配置
- 任务状态监控，实时查看扫描进度
- 任务暂停/恢复/停止控制

### 资产管理
- 资产列表，支持语法查询和快捷筛选
- 资产统计，端口/服务/应用/标题 Top 排行
- 资产历史，追踪资产变更记录
- 批量删除，支持多选操作

### 指纹识别
- 多引擎融合：Httpx + Wappalyzer + 自定义指纹
- 自定义指纹支持 ARL 格式导入
- 指纹来源标识，可追溯识别来源
- HTTP 服务映射配置

### 漏洞扫描
- Nuclei 引擎集成
- 自定义 POC 管理
- POC 标签映射，按指纹自动匹配 POC
- 漏洞等级分类统计

### 报告管理
- 任务报告自动生成
- 资产/漏洞统计概览
- Excel 导出功能

### 在线 API
- FOFA / Hunter / Quake 搜索
- 搜索结果一键导入为扫描任务
- API 配置管理

## 参考

- [go-zero](https://github.com/zeromicro/go-zero) - 微服务框架
- [Nuclei](https://github.com/projectdiscovery/nuclei) - 漏洞扫描引擎
- [Httpx](https://github.com/projectdiscovery/httpx) - HTTP 探测工具
- [Naabu](https://github.com/projectdiscovery/naabu) - 端口扫描器
- [Wappalyzer](https://github.com/projectdiscovery/wappalyzergo) - 指纹识别
- [nemo_go](https://github.com/hanc00l/nemo_go) - 参考项目

## License

MIT
