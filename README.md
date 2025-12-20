# CSCAN

**分布式网络资产扫描平台** | Go-Zero + Vue3

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 特性

- **资产发现** - Masscan + Nmap 端口扫描，Wappalyzer 指纹识别
- **漏洞检测** - 集成 Nuclei，支持自定义 POC
- **在线数据源** - FOFA / Hunter / Quake API 聚合
- **分布式架构** - Worker 节点水平扩展，Redis 任务队列
- **多工作空间** - 项目隔离，团队协作

## 快速开始

```bash
git clone https://github.com/tangxiaofeng7/cscan.git
cd cscan
docker-compose up -d --build
```

访问 `http://localhost:3000`，默认账号 `admin / 123456`

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
| 后端 | Go-Zero, gRPC |
| 存储 | MongoDB, Redis |
| 前端 | Vue 3, Element Plus |
| 扫描 | Nuclei, Nmap, Masscan |

## 本地开发

```bash
# 1. 启动依赖
docker-compose up -d redis mongodb

# 2. 启动服务
go run rpc/task/task.go -f rpc/task/etc/task.yaml
go run api/cscan.go -f api/etc/cscan.yaml
go run cmd/worker/main.go -s localhost:9000 -r localhost:6379 -n worker1

# 3. 启动前端
cd web && npm install && npm run dev
```

## 项目结构

```
├── api/          # HTTP API 服务
├── rpc/          # gRPC 服务
├── worker/       # 扫描 Worker
├── scanner/      # 扫描器实现 (nuclei/nmap/masscan)
├── model/        # 数据模型
├── onlineapi/    # FOFA/Hunter/Quake 集成
├── web/          # Vue3 前端
└── docker/       # Docker 配置
```

## 致谢

- [go-zero](https://github.com/zeromicro/go-zero) - 微服务框架
- [Nuclei](https://github.com/projectdiscovery/nuclei) - 漏洞扫描引擎
- [nemo_go](https://github.com/hanc00l/nemo_go) - 灵感来源

## License

MIT
