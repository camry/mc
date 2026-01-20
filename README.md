# mc

Mongo Collector

## 项目概述

**mc** 是一个名为 "Mongo Collector" 的 Go 语言库，用于简化 MongoDB 数据操作，提供了面向对象的方式来处理嵌套数据结构。

## 功能特性

### 1. 泛型数据结构支持
- **Adapter**: 提供批量更新机制，使用泛型参数 T IMessageTop, K comparable
- 支持嵌套消息结构，自动处理层级关系
- 实现了批量消费模式，提高数据库写入性能

### 2. 数据类型封装
#### 列表类型
- **ListBase[T BaseType]**: 基础类型列表（int32, int64, float32, float64, string）
- **ListMessage[T IMessage]**: 消息对象列表

#### 映射类型
- **MapBase[T comparable, V BaseType]**: 基础类型映射
- **MapMessage[T comparable, V IMessage]**: 消息对象映射

### 3. 消息层次结构
- **MessageTop[T IMessageTop, K comparable]**: 顶级消息结构
- **MessageSub**: 子级消息结构
- 支持无限层级嵌套结构

### 4. 解码机制
- 解码注册: 使用 Register 函数注册结构体与解码器映射
- 字段解码: DecodeField[T any] 处理单个字段
- 消息解码: DecodeMessage 处理整个消息结构

## 使用方式

### 1. 定义数据模型
```go
type A struct {
    mc.MessageTop[*A, int64]

    _id       int64
    aUsername string
    ids       *mc.ListBase[int32]
    b         *B
}
```

### 2. 注册解码器
```go
var aDecode = mc.Register(&A{},
    mc.NewDecodePair("_id", mc.NewDecodeField(func(vr bsonrw.ValueReader, v *A) (err error) {
        v._id, err = mc.Int64Decode(vr)
        return
    })),
    // ... 其他字段解码
)
```

### 3. 创建实例并使用
```go
// 创建适配器
adapter := mc.NewAdapter[*A, int64](db)

// 创建数据对象
a := NewA(adapter, 2)

// 操作数据
a.SetAUsername("username")           // 设置字符串值
a.GetIds().Append(123)               // 添加到列表
a.GetB().SetBUsername("subname")     // 设置嵌套对象
```

### 4. MongoDB 连接管理
- **MongoClient**: MongoDB 客户端封装
- **MongoDatabase**: 数据库实例
- **MongoCollection**: 集合操作

## 技术特点

### 1. 性能优化
- **批量更新**: Adapter.Consume() 方法实现批量写入，减少数据库请求次数
- **定时刷新**: 结合批量大小和时间间隔控制数据刷新频率

### 2. 类型安全
- 充分利用 Go 泛型特性，提供类型安全的数据操作
- 编译时检查，减少运行时错误

### 3. 灵活的数据操作
- 支持 Set 和 Unset 操作
- 自动处理嵌套字段路径构建

### 4. 完整的 CRUD 操作
- 支持基本类型、列表和映射的操作
- 提供查找、过滤、遍历等常用操作方法

## 应用场景
这个库特别适用于需要频繁操作复杂嵌套 MongoDB 文档的应用场景，例如用户属性管理、配置存储、游戏数据存储等。通过面向对象的方式简化了复杂的嵌套文档操作，同时保持了高性能的数据访问能力。
