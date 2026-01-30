# mc

Mongo Collector

## 项目概述

**mc** 是一个名为 "Mongo Collector" 的 Go 语言库，专为简化 MongoDB 复杂文档操作而设计。该库采用面向对象的方式处理嵌套数据结构，通过泛型和接口设计实现类型安全的 MongoDB 文档操作，并能够自动处理复杂的嵌套结构。项目通过 Adapter 模式提供数据更新机制，结合 Collector 的批量异步处理能力，实现了高效的数据持久化方案，特别适合处理复杂的嵌套文档结构。

该项目的核心在于提供了一种优雅的方式来处理 MongoDB 中的嵌套文档，将复杂的 BSON 操作抽象为简单的对象操作。通过消息结构（MessageTop、MessageSub）和数据容器（ListBase、ListMessage、MapBase、MapMessage）的组合，开发者可以轻松地操作深层嵌套的文档结构，而无需手动构建复杂的字段路径。其独特的批量处理机制允许将多个更新操作合并为单个批量请求，显著提高了数据库写入性能。

## 功能特性

### 1. 泛型数据结构支持

- **Adapter**: 提供数据更新机制，使用泛型参数 T IMessageTop, K BaseType
- 支持嵌套消息结构，自动处理层级关系
- 实现了批量消费模式，提高数据库写入性能

### 2. 数据类型封装

#### 列表类型

- **ListBase[T BaseType]**: 基础类型列表（int32, int64, float32, float64, string）
- **ListMessage[T IMessage]**: 消息对象列表

#### 映射类型

- **MapBase[T MapKeyType, V BaseType]**: 基础类型映射
- **MapMessage[T MapKeyType, V IMessage]**: 消息对象映射

### 3. 消息层次结构

- **MessageTop[T IMessageTop, K BaseType]**: 顶级消息结构
- **MessageSub**: 子级消息结构
- 支持无限层级嵌套结构

### 4. 解码机制

- 解码注册: 使用 Register 函数注册结构体与解码器映射
- 字段解码: DecodeField[T any] 处理单个字段
- 消息解码: DecodeMessage 处理整个消息结构

## 技术特点

### 1. 性能优化

- **批量更新**: Collector.Write() 方法实现批量写入，减少数据库请求次数
- **定时刷新**: 结合批量大小和时间间隔控制数据刷新频率
- **异步处理**: CollectorWorker 使用缓冲通道和定时器实现异步数据写入

### 2. 类型安全

- 充分利用 Go 泛型特性，提供类型安全的数据操作
- 编译时检查，减少运行时错误
- 严格的类型约束(BaseType, MapKeyType等)

### 3. 灵活的数据操作

- 支持 Set 和 Unset 操作
- 自动处理嵌套字段路径构建
- 支持基础类型(int32/64, uint32/64, float32/64, string, bool, []byte)和复杂对象的混合操作

### 4. 完整的 CRUD 操作

- 支持基本类型、列表和映射的操作
- 提供查找、过滤、遍历等常用操作方法
- 支持嵌套结构的增删改查

### 5. 高效的编解码

- 支持 BSON 自动编解码
- 通过 Protobuf 代码生成工具快速创建数据结构
- 自定义解码器支持灵活的数据格式转换

## 应用场景

这个库特别适用于需要频繁操作复杂嵌套 MongoDB 文档的应用场景，例如用户属性管理、配置存储、游戏数据存储等。通过面向对象的方式简化了复杂的嵌套文档操作，同时保持了高性能的数据访问能力。该库的设计使得开发者可以像操作普通对象一样操作 MongoDB 文档，而无需关心底层的 BSON 转换和查询构造。

## 使用方式

```bash
protoc --mc_out=./pb pb/*.proto
```

## 示例代码

以下是一个典型的使用示例，展示了如何使用 mc 库来操作嵌套的 MongoDB 文档：

```go
package main

import (
    "context"
    "math"
    "time"

    "github.com/camry/g/glog"
    "github.com/camry/mc"
    "github.com/camry/mc/pb"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // 初始化 MongoDB 连接
    client, err := mc.NewMongoClient(
        context.Background(),
        options.Client().ApplyURI("mongodb://localhost:27017"),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // 获取数据库实例
    db := client.Database("mydb")

    // 创建收集器
    collector := mc.NewCollector(context.Background(), db)

    // 创建适配器
    adapter := mc.NewAdapter[*pb.A, int64](
        mc.WithCollector[*pb.A, int64](collector),
        mc.WithCollectorWorkerOption[*pb.A, int64](mc.WithCollectorWorkerFlushInterval(100*time.Millisecond)),
    )

    // 创建 A 对象实例
    a := pb.NewA(adapter, 2)

    // 更新嵌套字段
    a.SetAUsername("camry")
    a.SetANicknameAlias("Camry.Chen")

    // 操作嵌套的 B->C->D 结构
    a.GetB().GetC().GetD().SetId(math.MaxInt32)
    a.GetB().GetC().GetD().SetIid(math.MaxUint32)
    a.GetB().GetC().GetD().SetUid(math.MaxInt64)
    a.GetB().GetC().GetD().SetDUsername("camry")
    a.GetB().GetC().GetD().SetIsBool(true)
    a.GetB().GetC().GetD().SetFltV32(1234.1234)
    a.GetB().GetC().GetD().SetFltV64(1234.123456789)
    a.GetB().GetC().GetD().SetBytes([]byte("hello"))

    // 操作嵌套列表
    a.GetB().GetC().GetD().GetTags().Append("tag3")

    // 等待批量写入完成
    time.Sleep(5 * time.Second)

    // 关闭收集器
    _ = collector.Close()

    glog.Infof("Final object: %v", a.Marshal())
}
```

## 设计原理

mc 库的核心设计原理是通过接口和泛型实现类型安全的 MongoDB 文档操作。它将复杂的嵌套文档操作分解为以下几个核心组件：

1. **IMessage 接口族**：定义了消息对象的基本行为，包括类型标识、键值操作、字段设置等。
2. **Collector 体系**：负责批量收集和异步处理数据库操作，提升性能。
3. **Adapter 模式**：将业务对象与数据库操作解耦，提供统一的数据访问接口。
4. **自动路径构建**：通过嵌套消息链自动构建 MongoDB 的字段路径，如 "b.c.d.username"。
5. **编解码框架**：支持从 Protobuf 定义自动生成代码，并提供 BSON 编解码功能。
