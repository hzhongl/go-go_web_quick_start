# Go Web Quick Start 脚手架

这是一个Go Web项目脚手架，用于快速构建基于Gin的Web应用程序，类似于SpringBoot的脚手架。

## 特性

- **Web框架**: 使用Gin快速搭建基础RESTful风格API
- **数据库支持**: 支持MySQL、PostgreSQL、SQLite、Oracle，使用GORM实现对数据库的基本操作
- **缓存**: 使用Redis实现缓存功能
- **配置文件**: 使用fsnotify和viper实现yaml格式的配置文件，支持热更新
- **日志**: 使用zap实现高性能日志记录
- **依赖注入**: 使用wire作为依赖注入工具
- **代码生成**: 内置数据表代码生成器
- **基础CRUD**: 提供base_service.go和base_dao.go实现通用CRUD操作

## 项目结构

```
.
├── cmd                 # 主要应用程序入口
├── config             # 配置文件目录
├── internal           # 私有应用程序和库代码
│   ├── api            # API层，处理HTTP请求
│   ├── dao            # 数据访问层
│   ├── middleware     # HTTP中间件
│   ├── model          # 数据模型
│   └── service        # 业务逻辑层
├── pkg                # 公共库代码
│   ├── cache          # Redis缓存实现
│   ├── config         # 配置加载
│   ├── database       # 数据库连接
│   ├── logger         # 日志实现
│   └── utils          # 工具函数
├── scripts            # 脚本，包括代码生成器
└── test               # 测试文件
```

## 快速开始

1. 克隆项目
2. 修改`config/config.yaml`配置文件
3. 运行`go mod tidy`安装依赖
4. 运行`go run cmd/main.go`启动应用

## 代码生成

使用内置的代码生成器可以快速生成模型对应的CRUD代码：

```bash
go run scripts/generator/main.go -model=User
```

## 依赖注入

项目使用Wire进行依赖注入，生成依赖关系：

```bash
wire ./...ß
```