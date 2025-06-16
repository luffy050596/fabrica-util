<div align="center">
  <h1>🏛️ FABRICA UTIL</h1>
  <p><em>为 go-pantheon 生态系统提供的综合工具库</em></p>
</div>

<p align="center">
<a href="https://github.com/go-pantheon/fabrica-util/actions/workflows/test.yml"><img src="https://github.com/go-pantheon/fabrica-util/workflows/Test/badge.svg" alt="Test Status"></a>
<a href="https://github.com/go-pantheon/fabrica-util/releases"><img src="https://img.shields.io/github/v/release/go-pantheon/fabrica-util" alt="Latest Release"></a>
<a href="https://pkg.go.dev/github.com/go-pantheon/fabrica-util"><img src="https://pkg.go.dev/badge/github.com/go-pantheon/fabrica-util" alt="GoDoc"></a>
<a href="https://goreportcard.com/report/github.com/go-pantheon/fabrica-util"><img src="https://goreportcard.com/badge/github.com/go-pantheon/fabrica-util" alt="Go Report Card"></a>
<a href="https://github.com/go-pantheon/fabrica-util/blob/main/LICENSE"><img src="https://img.shields.io/github/license/go-pantheon/fabrica-util" alt="License"></a>
<a href="https://deepwiki.com/go-pantheon/fabrica-util"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
</p>

> **语言**: [English](README.md) | [中文](README-zh.md)

## 关于 Fabrica Util

Fabrica Util 是为 go-pantheon 生态系统提供的综合工具库，为所有 go-pantheon 组件提供通用功能。这个库封装了可重用的代码模式、算法和辅助函数，确保游戏服务器微服务架构中的一致性，避免代码重复。

更多信息请查看：[deepwiki/go-pantheon/fabrica-util](https://deepwiki.com/go-pantheon/fabrica-util)

## 关于 go-pantheon 生态系统

**go-pantheon** 是一个开箱即用的游戏服务器框架，基于 [go-kratos](https://github.com/go-kratos/kratos) 微服务架构提供高性能、高可用的游戏服务器集群解决方案。Fabrica Util 作为基础工具库，支持以下核心组件：

- **Roma**：游戏核心逻辑服务
- **Janus**：网关服务，处理客户端连接和请求转发
- **Lares**：账户服务，用于用户认证和账户管理
- **Senate**：后台管理服务，提供运营管理接口

### 核心特性

- 🕒 **时间工具**：高级时间处理，支持多语言、时区管理和时间段计算
- 🔄 **并发处理**：线程安全的同步原语，包括延迟器、Future 和协程管理
- 🔐 **安全模块**：全面的加密工具（AES-GCM、RSA、ECDH），用于安全数据传输
- 🆔 **ID 管理**：分布式 ID 生成，支持区域编码和 HashID 混淆
- 🎲 **随机化**：安全的随机数生成和字符串创建工具
- 📊 **数据结构**：高性能实现（布隆过滤器、位图、一致性哈希）
- 🧠 **内存管理**：多池内存管理，优化资源利用
- 🔤 **字符串处理**：大小写转换和文本处理工具
- ⚠️ **错误处理**：增强的错误处理，支持上下文和堆栈跟踪

## 工具包

### 时间工具 (`xtime/`)
支持多语言的高级时间处理：
- 可配置的时区和语言支持
- 本地化时间格式转换
- 日/周/月周期计算
- 时区转换工具

### 同步工具 (`xsync/`)
线程安全的同步原语：
- **Delayer**：基于时间的任务调度和过期管理
- **Future**：异步计算结果
- **Closure**：线程安全的函数执行包装器
- **Routines**：协程生命周期管理

### ID 生成 (`xid/`)
分布式 ID 管理系统：
- 基于区域的 ID 组合，支持多区域
- HashID 编码/解码，用于前端显示
- ID 混淆，提升安全性

### 安全模块 (`security/`)
全面的加密操作：
- **AES**：AES-GCM 加密/解密，安全的随机数生成
- **RSA**：公钥/私钥操作
- **ECDH**：椭圆曲线 Diffie-Hellman 密钥交换
- **Certificate**：X.509 证书处理工具

### 数据结构
- **布隆过滤器** (`bloom/`)：内存高效的集合成员测试
- **位图** (`bitmap/`)：位级操作，紧凑数据存储
- **一致性哈希** (`consistenthash/`)：负载均衡的分布式哈希环

### 其他工具
- **随机数** (`xrand/`)：加密安全的随机数生成
- **压缩** (`compress/`)：数据压缩工具
- **驼峰命名** (`camelcase/`)：字符串大小写转换工具
- **内存池** (`multipool/`)：内存池管理
- **错误处理** (`errors/`)：增强的错误处理与上下文

## 技术栈

| 技术/组件         | 用途                         | 版本    |
| ----------------- | ---------------------------- | ------- |
| Go                | 主要开发语言                 | 1.23+   |
| crypto            | 加密操作                     | v0.39.0 |
| go-redis          | Redis 客户端，用于分布式操作 | v9.10.0 |
| PostgreSQL Driver | 数据库连接                   | v5.7.5  |
| MongoDB Driver    | NoSQL 数据库操作             | v2.2.2  |
| HashIDs           | ID 混淆库                    | v2.0.1  |
| Murmur3           | 快速哈希算法                 | v1.1.0  |

## 系统要求

- Go 1.23+

## 快速开始

### 安装

```bash
go get github.com/go-pantheon/fabrica-util
```

### 初始化开发环境

```bash
make init
```

### 运行测试

```bash
make test
```

## 使用示例

### 多语言时间处理

```go
package main

import (
    "fmt"
    "time"

    "github.com/go-pantheon/fabrica-util/xtime"
)

func main() {
    // 初始化配置
    err := xtime.Init(xtime.Config{
        Language: "zh",
        Timezone: "Asia/Shanghai",
    })
    if err != nil {
        panic(err)
    }

    // 格式化当前时间
    fmt.Println(xtime.Format(time.Now()))

    // 计算下次每日重置时间（上午5点重置）
    nextReset := xtime.NextDailyTime(time.Now(), 5*time.Hour)
    fmt.Println("下次每日重置:", nextReset)

    // 获取本周开始时间
    weekStart := xtime.StartOfWeek(time.Now())
    fmt.Println("本周开始:", weekStart)
}
```

### AES-GCM 加密

```go
package main

import (
    "fmt"

    "github.com/go-pantheon/fabrica-util/security/aes"
)

func main() {
    // 创建 AES 加密器，使用 32 字节密钥
    key := []byte("0123456789abcdef0123456789abcdef")
    cipher, err := aes.NewAESCipher(key)
    if err != nil {
        panic(err)
    }

    data := []byte("敏感的游戏数据")

    // 加密数据
    encrypted, err := cipher.Encrypt(data)
    if err != nil {
        panic(err)
    }

    // 解密数据
    decrypted, err := cipher.Decrypt(encrypted)
    if err != nil {
        panic(err)
    }

    fmt.Printf("原始数据: %s\n", data)
    fmt.Printf("解密数据: %s\n", decrypted)
}
```

### 基于区域的 ID 管理

```go
package main

import (
    "fmt"

    "github.com/go-pantheon/fabrica-util/xid"
)

func main() {
    // 组合区域 ID 和区域编号
    playerID := int64(12345)
    zoneNum := uint8(3)
    combinedID := xid.CombineZoneID(playerID, zoneNum)

    // 编码 ID 用于前端显示
    encodedID, err := xid.EncodeID(combinedID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("编码 ID: %s\n", encodedID)

    // 解码 ID
    decodedID, err := xid.DecodeID(encodedID)
    if err != nil {
        panic(err)
    }

    // 分割 ID 回到原始组件
    originalPlayerID, originalZone := xid.SplitID(decodedID)
    fmt.Printf("玩家 ID: %d, 区域: %d\n", originalPlayerID, originalZone)
}
```

### 延迟器同步

```go
package main

import (
    "fmt"
    "time"

    "github.com/go-pantheon/fabrica-util/xsync"
)

func main() {
    // 创建延迟器
    delayer := xsync.NewDelayer()
    defer delayer.Close()

    // 设置过期时间（5秒后）
    expiryTime := time.Now().Add(5 * time.Second)
    delayer.SetExpiryTime(expiryTime)

    fmt.Println("等待延迟器过期...")

    // 等待过期
    select {
    case <-delayer.Wait():
        fmt.Println("延迟器已过期！")
    case <-time.After(10 * time.Second):
        fmt.Println("等待延迟器超时")
    }
}
```

## 项目结构

```
.
├── xtime/              # 支持本地化的时间工具
├── xsync/              # 同步原语
│   ├── delayer.go      # 基于时间的任务调度
│   ├── future.go       # 异步计算
│   ├── closure.go      # 线程安全函数包装器
│   └── routines.go     # 协程管理
├── xrand/              # 安全随机数生成
├── xid/                # ID 生成和混淆
├── security/           # 加密操作
│   ├── aes/            # AES-GCM 加密
│   ├── rsa/            # RSA 加密
│   ├── ecdh/           # 椭圆曲线 Diffie-Hellman
│   └── certificate/    # X.509 证书工具
├── consistenthash/     # 一致性哈希实现
├── multipool/          # 内存池管理
├── errors/             # 增强错误处理
├── bloom/              # 布隆过滤器实现
├── compress/           # 数据压缩工具
├── bitmap/             # 位图数据结构
└── camelcase/          # 字符串大小写转换
```

## 与 go-pantheon 组件集成

Fabrica Util 设计用于被其他 go-pantheon 组件导入：

```go
import (
    // Lares 中用于令牌生成的安全工具
    "github.com/go-pantheon/fabrica-util/security/aes"

    // Roma 中用于游戏逻辑的时间工具
    "github.com/go-pantheon/fabrica-util/xtime"

    // Janus 中用于连接处理的同步工具
    "github.com/go-pantheon/fabrica-util/xsync"

    // 分布式玩家身份识别的 ID 管理
    "github.com/go-pantheon/fabrica-util/xid"
)
```

## 开发指南

### 许可证合规

项目对所有依赖项强制执行许可证合规。我们只允许以下许可证：
- MIT
- Apache-2.0
- BSD-2-Clause
- BSD-3-Clause
- ISC
- MPL-2.0

许可证检查在以下情况执行：
- CI/CD 流水线中自动执行
- 通过预提交钩子在本地执行
- 使用 `make license-check` 手动执行

### 测试

运行完整的测试套件：

```bash
# 运行所有测试并生成覆盖率报告
make test

# 运行代码检查
make lint

# 运行 go vet
make vet
```

### 添加新工具

添加新的工具函数时：

1. 根据功能创建新包或添加到现有包中
2. 实现具有适当错误处理的工具函数
3. 编写涵盖边界情况的全面单元测试
4. 用清晰的示例记录使用方法
5. 确保适用时的线程安全
6. 运行测试：`make test`
7. 如需要，更新文档

### 贡献指南

1. Fork 这个仓库
2. 从 `main` 分支创建功能分支
3. 实现带有全面测试的更改
4. 确保所有测试通过且代码检查清洁
5. 更新任何 API 更改的文档
6. 提交带有清晰描述的 Pull Request

## 性能考虑

- **内存池**：对于高频对象分配，使用 `multipool`
- **加密**：AES-GCM 操作针对吞吐量进行了优化
- **ID 生成**：HashID 编码针对重复操作进行了缓存
- **时间操作**：时区加载被缓存和重用
- **同步**：所有同步原语都设计为低竞争

## 许可证

本项目根据 LICENSE 文件中指定的条款获得许可。
