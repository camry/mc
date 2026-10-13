---
name: mc-knowledge
description: mc（Mongo Collector）仓库知识库——架构、类型体系、解码系统、Adapter/Collector 写入链路、MongoTable 查询、protoc-gen-mc 生成代码惯例与常见陷阱。在修改或使用本仓库代码、排查 MongoDB 嵌套文档读写问题时使用。
---

# mc 仓库知识库

**项目全名**: Mongo Collector  
**仓库路径**: `D:\project\meigushi\mc`  
**模块名**: `gitea.meigushi.go/meigushi/mc`  
**Go 版本**: 1.26+（`go.mod`: `go 1.26.0`）  
**依赖**:
- `go.mongodb.org/mongo-driver/v2` v2.9.1
- `github.com/camry/g/v2`（`gerrors/gerror` 构造错误、`glog` 日志）
- `github.com/spf13/cast`
- `golang.org/x/exp`（`constraints.Float`）
- `github.com/stretchr/testify`（仅测试）

**定位**：简化 MongoDB 复杂嵌套文档操作的 Go 库。开发者像操作普通对象一样操作 MongoDB 文档，无需手动构建字段路径或 BSON 查询。适用于用户属性管理、配置存储、游戏数据存储等频繁操作嵌套文档的场景。

---

## 一、整体架构

mc 是一个 MongoDB 数据访问层框架，核心思路是：

```
MongoDB 文档
  ↓  Find（MongoDatabase，读写对象） / FindOne / FindMany（MongoTable，只读对象）
bson.Raw 字节流
  ↓  bson.NewDocumentReader → IDecode.Decode
  Go struct（MessageTop / MessageSub / ListXxx / MapXxx）
  ↑  Set / Unset（字段变更）
Adapter.buffer（channel）
  ↓  批量聚合 $set / $unset
Collector → CollectorWorker → BulkWrite
  ↓
MongoDB 集合
```

所有 struct 通过代码生成（`protoc-gen-mc`，当前 v2.0.6）产生，生成的文件以 `.pb.mc.go` 结尾。示例见 `pb/a.proto` / `pb/a.pb.mc.go`。

### 文件速查

| 文件 | 内容 |
|------|------|
| `define.go` | 类型约束、枚举（`MessageType`/`AggregateStage`）、全部接口 |
| `message_top.go` / `message_sub.go` | 顶级/子级消息 |
| `list*.go` / `map*.go` | 列表、映射容器 |
| `utils_map.go` | `MapBaseIncr/MapBaseDecr/MapMessageGetOrPut` 辅助函数 |
| `decode*.go` | 解码注册与分发 |
| `utils_decode.go` | 基础类型解码函数 |
| `utils.go` | `GetStructName`、字段路径拼接、`toMapKey`、`MToValueReader` |
| `adapter*.go` | 写入聚合层 |
| `collector*.go` | 按表批量写入层 |
| `mongo*.go` | Client/Database/Collection/Table 封装 |
| `*_z_test.go` | 测试（连真实 MongoDB，见 `mongo_z_test.go` 的 `defaultDB`） |

---

## 二、类型约束体系

```
BaseType        = BaseTypeNumber | ~string
BaseTypeNumber  = BaseTypeInteger | constraints.Float
BaseTypeInteger = ~int32 | ~int64 | ~uint32 | ~uint64
MapKeyType      = BaseTypeInteger | ~string
```

泛型参数惯例：
- `T IMessageTop`：顶级消息（根节点，对应 MongoDB collection）
- `K BaseType`：主键类型（通常 int64 或 string）
- `V BaseType` / `V IMessage`：集合/映射的值类型

`~` 约束使 proto 枚举（`type AEnum int32`）可直接作为 List/Map 的元素或键。

---

## 三、消息层级结构

### 接口（`define.go`）

| 接口 | 说明 |
|------|------|
| `IMessage` | 所有消息基础接口：`MessageType() / Key() / Set(key, val, nested...) / Unset(key, nested...)` |
| `IMessageTop` | 顶级消息，扩展 `IMessage + IMarshalBsonM`，增加 `Table() string` |
| `IMessageSub` | 子消息，扩展 `IMessage + IMarshalBsonM` |
| `IMarshalBsonM` | `Marshal() bson.M / Unmarshal(bson.M) error` |
| `IDecode` | `Decode(vr bson.ValueReader, v any) error` |
| `IAdapter[T, K]` | `Bind / GetMessageTop / GetKey / Set / Unset / Close` |
| `IAdapterBsonM` | `AggregateStage() / M() bson.M`（实现为私有 `adapterBsonM`） |
| `ICollector` | `Database() / Write(table, writeModel, opts...) / Close()` |
| `ICollectorWorker` | `Write / Run / Close` |

### 具体结构体

| 类型 | 说明 |
|------|------|
| `MessageTop[T, K]` | 顶级节点，持有 `IAdapter`；`Set/Unset` 转发给 Adapter；**adapter 为 nil 时 Set/Unset 静默忽略、`Key()` 返回 nil** |
| `MessageSub` | 子节点，持有 `parent IMessage + key any`；`Set/Unset` 追加自身到 nestedMessages 链后向上冒泡 |
| `List[T]` | 通用有序列表基类：`Len/IsEmpty/Values/Contains(DeepEqual)/Find/Filter/ForEach/Append/Remove(index)/Clear/Marshal` |
| `ListBase[T BaseType]` | 基础类型列表，实现 BSON 数组解码；**`Remove(item T)` 按值删除**（遮蔽 `List.Remove(index)`） |
| `ListMessage[T IMessage]` | 消息对象列表，带 `creator` 工厂；`AppendNewMessage()` 以当前长度为 key 创建并追加 |
| `Map[T, V]` | 通用映射基类：`Get/Contains/Size/IsEmpty/Keys/Values/GetAll(副本)/Range/Clear/Marshal`（**无 Put/Delete**） |
| `MapBase[T MapKeyType, V BaseType]` | 基础类型映射：`Put(key, value)` → `$set path.key`；`Delete(key)` → `$unset path.key`；实现 BSON document 解码 |
| `MapMessage[T MapKeyType, V IMessage]` | 消息对象映射，带 `creator` 工厂：`PutNewMessage(key)` 创建、存入并 `$set` 其 Marshal 结果；`Delete(key)` |

### 辅助函数（`utils_map.go`）

| 函数 | 说明 |
|------|------|
| `MapBaseIncr(m, key, v)` / `MapBaseDecr(m, key, v)` | 数值映射增减，返回**新值**（与生成代码 `IncrXxx/DecrXxx` 一致） |
| `MapMessageGetOrPut(m, key)` | 存在则返回，不存在则 `PutNewMessage` |

### 字段变更传播路径

```
叶子.Set(key, val)
  → MessageSub / ListXxx / MapXxx.Set：nestedMessages 追加 self，调 parent.Set
  → ... 逐层向上 ...
  → MessageTop.Set → Adapter.Set(nestedMessages, key, val)
  → toNestedMessageFieldKey(nestedMessages, key) 构造 "a.b.c.key"
  → write(adapterBsonM{$set, bson.M{"a.b.c.key": val}})
```

`toNestedMessageFieldKey` / `toNestedMessageKey` 在 `utils.go` 中，从 nestedMessages 尾部向前拼接点分隔的键路径（键经 `cast.ToString`）。

各容器的写入粒度：
- `List.Append/Remove/Clear`：**整个数组** `$set x.key = Marshal()`
- `MapBase.Put` / `MapMessage.PutNewMessage`：单个键 `$set path.key`
- `Map.Clear`：整个映射 `$set x.key = {}`
- `MapXxx.Delete`：单个键 `$unset path.key`

---

## 四、解码系统

### 注册与查找（`decode.go`）

```go
// 生成代码中：
var xDecode = mc.Register(&X{}, mc.NewDecodePair("field", mc.NewDecodeField(...)), ...)

// 运行时查找：
decode, err := mc.GetDecode(obj)   // 按 GetStructName(obj)（去指针、不含包名）查 decodeMap
```

`Register` → `decodeMap[structName] = *DecodeMessage`（`decodeMap map[string]IDecode`，包级全局，无锁，仅在 init 期写入）。

> 因为以不含包名的结构体名为键，**不同包中同名 message 会互相覆盖**。

### DecodeMessage.Decode 流程

1. 若 `vr.Type() == TypeNull/Undefined` → 直接消费并返回
2. `vr.ReadDocument()` 开始读文档
3. 逐元素：找到 `fieldMap[key]` → `handler.Decode(evr, v)`；否则 `evr.Skip()`

`DecodeField[T]` 把 `v any` 断言成 `T` 后调用生成的 handler。

### 基础类型解码函数（`utils_decode.go`）

| 函数 | 支持的 BSON 类型 |
|------|----------------|
| `Int64Decode` | Int32 / Int64 / Boolean / Double / **Null** / **Undefined** |
| `Int32Decode` | 委托 Int64Decode |
| `UInt32Decode` | 委托 Int64Decode |
| `UInt64Decode` | 委托 Int64Decode |
| `EnumDecode[T ~int32]` | 委托 Int64Decode |
| `Float64Decode` | Int32 / Int64 / Boolean / Double / **Null** / **Undefined** |
| `Float32Decode` | 委托 Float64Decode |
| `BoolDecode` | Boolean / **Null** / **Undefined**（null→false） |
| `StringDecode` | String / **Null** / **Undefined**（null→""） |
| `BinaryDecode` | Binary / String(base64) / **Null** / **Undefined** |
| `BaseDecode[T BaseType]` | 泛型解码入口：按 `T` 的 Kind 分发到上述各函数，最后 `reflect.Convert` 转成 T（支持枚举等派生类型） |
| `canBaseDecode[T BaseType](t)` | 判断源 bson 类型是否可解码：**仅 Int32 / Int64 / Double / String**（不含 Boolean） |

> **注意**：`canBaseDecode` 门只放行 Int32/Int64/Double/String 源类型，Boolean/null/undefined 等源值在容器解码时被 `Skip` 跳过。`BaseDecode` 的 `BoolDecode` 分支实际只服务于 struct 的 bool 字段（`bool` 不在 `BaseType` 约束内，无法作为 List/Map 的值类型）。

### 容器解码

- `ListBase.Decode`：非数组 → `vr.Skip()`；元素源类型不在 `canBaseDecode[T]` 放行集内则 `Skip` 跳过；否则用 `BaseDecode[T]` 宽松转换（数值类型互相兼容，支持 uint32/uint64/bool/string 及枚举等派生类型）。
- `ListMessage.Decode`：非数组 → `vr.Skip()`；只处理 `EmbeddedDocument` 元素（其他 `Skip` 跳过）；元素 key 为当前下标。
- `MapBase.Decode` / `MapMessage.Decode`：非 `EmbeddedDocument` → `vr.Skip()`；值源类型不可解码或为 null/undefined → `Skip` 并 `continue`（**不再中止整个 map 解码**）；键经 `toMapKey[T](k, reflect.TypeFor[T]().Kind())` 转换（支持 int32/int64/uint32/uint64/string 及其派生类型）。

---

## 五、Adapter（写入聚合层）

```go
adapter := mc.NewAdapter[*MyMessage, int64](
    mc.WithCollector[*MyMessage, int64](collector),
    mc.WithCollectorWorkerOption[*MyMessage, int64](opts...),
    mc.WithContext[*MyMessage, int64](ctx),
)
obj := pb.NewMyMessage(adapter, id) // 内部 Bind(adapter, obj, id)
```

**一个 Adapter 只绑定一个顶级对象**（`Bind` 保存 t、k），每个文档需要各自的 Adapter。

**内部机制**：
- `buffer chan IAdapterBsonM`（默认容量 1000）；仅在 collector 非 nil 时创建并启动 goroutine
- goroutine `writeCollectorWorker`：ticker（默认 10ms）或 batchSize（默认 10）触发 flush；缓冲区通道关闭且剩余数据 flush 完才退出（`defer close(done)`）
- `doBatchAdapterBsonM`：按顺序合并同阶段的 `bson.M`；阶段切换（$set↔$unset）先 flush；检测父子键冲突（`HasPrefix(k+".")`），冲突则先 flush 再合并；相同键直接覆盖
- 最终生成 `mongo.UpdateOneModel(filter {_id: key}, $set/$unset)` 交给 `Collector.Write(table, model, collectorWorkerOption...)`
- Adapter 的 bufferSize/batchSize/flushInterval **没有对应 Option**，只能用默认值

**并发与关闭**：
- `mu sync.RWMutex` 保护 `closed` 与 `close(buffer)`：`write` 持读锁、`shutdown` 持写锁，因此不会向已关闭的 channel 发送（`write` 先判断 `closed || ctx.Err()`，命中则告警丢弃）
- `context.AfterFunc(ctx, shutdown)`：外部 context 取消时同样停止接收写入并关闭 buffer
- `adapter.Close()`：`cancel()` → `shutdown()`（置 closed + close(buffer)）→ `<-done` **阻塞等待**剩余数据全部交给 Collector；可重复调用；须在 `collector.Close()` 之前调用

---

## 六、Collector / CollectorWorker（批量写入层）

```go
collector := mc.NewCollector(ctx, db)
```

- 每张表懒初始化一个 `CollectorWorker`（首次 `Write` 时经 `getOrCreateWorker` 创建并 `go worker.Run()`；**二次检查**避免并发首写同表创建多个 worker）
- `CollectorWorkerOption` 仅在该表 worker **首次创建时生效**，之后传入的选项被忽略
- `CollectorWorker`：buffer（默认 10000）、batchSize（默认 100）、flushInterval（默认 100ms）、writeTimeout（默认 10s）、maxRetries（默认 3）、retryBaseDelay（默认 100ms）、retryMaxDelay（默认 2s）；可用 `WithCollectorWorkerBufferSize/BatchSize/FlushInterval/WriteTimeout/MaxRetries/RetryBaseDelay/RetryMaxDelay` 调整
- `Run()` goroutine：ticker 或 batchSize 触发 `BulkWrite`；缓冲区通道关闭且剩余数据写完才退出（`defer close(done)`）
- `doBatchWriteModel`：`BulkWrite` 用 `context.WithoutCancel(db.Context())` + 可选 `writeTimeout` 超时（<=0 不限制）；**失败即重试**（任何错误都重试）：指数退避 `retryBaseDelay*2^(n-1)` 封顶 `retryMaxDelay`、再加抖动 `[d/2, d]`，退避等待可被 `ctx.Done()` 打断；重试耗尽后逐条 `%+v` 记录被丢弃的模型（不静默丢）
- `Close()`：置 closed、清空 worker 映射，并发调用各 worker 的 `Close()`（`sync.WaitGroup.Go`，Go 1.26）并 `wg.Wait()` **阻塞等待**所有剩余数据写完；关闭后 `Write` 被丢弃并告警
- `CollectorWorker.Close()`：`cancel()` → `shutdown()`（close buffer）→ `<-done` 阻塞等待，可重复调用

---

## 七、MongoClient / MongoDatabase / MongoCollection

三层封装，主要是注入 `context.Context`：

```
mc.NewMongoClient(ctx, opts...)   // Connect + Ping(Primary)
MongoClient.Database(name) → MongoDatabase
MongoDatabase.Collection(name) → MongoCollection
MongoClient.Close()               // Disconnect
```

`MongoDatabase.Find(top IMessageTop)`（用于已通过 `NewX(adapter, k)` 创建、带 Adapter 的对象）：
- `FindOne(_id=top.Key())`
- 若 `ErrNoDocuments` → `InsertOne(top.Marshal())`（自动创建文档）
- 否则用 `decode.Decode` 反序列化到 top

---

## 八、MongoTable[T, K]（只读查询）

```go
table := mc.NewMongoTable[*MyMessage, int64](db)

obj, err := table.FindOne(id)       // 查单条；无文档时返回 mongo.ErrNoDocuments
list, err := table.FindMany(ids)    // 查多条（$in 查询，batchSize=2048）
```

- 用 `reflect.TypeFor[T]().Elem()` + `reflect.New(t).Interface().(T)` 创建对象，**不绑定 Adapter**：得到的对象调用 Setter 只改内存，不会写库；`Key()` 返回 nil（用 `GetXId()` 取主键）。
- **FindOne**：失败时用 `var zero T` 返回泛型零值；反序列化用 `bson.NewDocumentReader(bytes.NewReader(bsonRaw))`。
- **FindMany**：游标逐条解码，每次循环新建对象；返回顺序为 Mongo 游标顺序（不保证与 ids 一致），不存在的 id 直接缺失。

---

## 九、代码生成惯例（`.pb.mc.go`）

命名规则：
- 表名 `Table()` = message 名（如 `"A"`）
- BSON 字段名 = proto 字段转 lowerCamel（`a_nickname_alias` → `aNicknameAlias`，`d_list` → `dList`）
- `_id` 字段 Getter 为 `GetXId()`，无 Setter
- proto enum → `type AEnum int32` + 常量 `AEnum_A_ENUM_0`，解码用 `mc.EnumDecode[AEnum]`
- `repeated 基础类型/枚举` → `*mc.ListBase[T]`；`repeated Message` → `*mc.ListMessage[*M]`
- `map<K, 基础类型/枚举>` → `*mc.MapBase[K, V]`；`map<K, Message>` → `*mc.MapMessage[K, *M]`
- `bytes` → `[]byte` + `mc.BinaryDecode`

生成文件的固定模式：

```go
// 1. 定义 struct（内嵌 mc.MessageTop[*T, K] 或 mc.MessageSub）
// 2. 注册解码器
var xDecode = mc.Register(&X{},
    mc.NewDecodePair("fieldName", mc.NewDecodeField(func(vr bson.ValueReader, v *X) (err error) {
        v.field, err = mc.Int64Decode(vr)  // 或其他 XxxDecode / 子消息 yDecode.Decode / 容器 .Decode
        return
    })),
    ...
)
// 3. 构造函数 NewX(adapter, k)：设置 _id 并 Bind(adapter, x, k)
// 4. Table() string
// 5. Marshal() bson.M
// 6. Unmarshal(bson.M) error（MToValueReader + decode）
// 7. Getter/Setter（数值额外有 Incr/Decr，返回新值）
// 8. 子消息用 newX(parent, key)，MessageSub 嵌入；子消息字段有 UnsetX()
```

子消息 / 容器字段的 Getter 带懒初始化：
```go
func (x *A) GetB() *B {
    if x.b == nil { x.b = newB(x, "b") }
    return x.b
}
```

---

## 十、完整使用示例

代码生成命令：
```bash
protoc --mc_out=./pb pb/*.proto
```

典型用法：
```go
// 1. 连接
client, _ := mc.NewMongoClient(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
defer client.Close()

// 2. Collector + Adapter
db := client.Database("mydb")
collector := mc.NewCollector(context.Background(), db)
adapter := mc.NewAdapter[*pb.A, int64](
    mc.WithCollector[*pb.A, int64](collector),
    mc.WithCollectorWorkerOption[*pb.A, int64](mc.WithCollectorWorkerFlushInterval(100*time.Millisecond)),
)

// 3. 创建对象（可选：db.Find(a) 加载已有文档，不存在则插入）
a := pb.NewA(adapter, 2)
_ = db.Find(a)

// 4. 操作嵌套字段
a.SetAUsername("camry")
a.GetB().GetC().GetD().SetId(math.MaxInt32)
a.GetB().GetC().GetD().SetBytes([]byte("hello"))
a.GetB().GetC().GetD().GetTags().Append("tag3")
a.GetB().GetC().GetTagMap().Put("k", pb.AEnum_A_ENUM_1)
mc.MapMessageGetOrPut(a.GetB().GetC().GetDMap(), 1).SetUid(100)

// 5. 先关 Adapter 再关 Collector（两者均阻塞等待剩余数据刷出）
_ = adapter.Close()
_ = collector.Close()
```

---

## 十一、常见陷阱与规则

1. **泛型零值**：`FindOne/FindMany` 的 error 返回必须用 `var zero T; return zero, err`，不能 `return nil, err`。

2. **reflect.TypeFor[T]().Elem()**：比 `reflect.TypeOf(new(T)).Elem().Elem()` 更简洁，二者等价（需 Go 1.22+），仓库已统一使用 `reflect.TypeFor`。

3. **Adapter 无 Collector 时静默失败**：`Set/Unset` 若 `collector == nil` 只打 error 日志，不 panic；MessageTop 无 Adapter（如 MongoTable 查出的对象）时 Set/Unset 直接忽略。

4. **父子键冲突**：同一 batch 中若出现 `"a"` 和 `"a.b"` 两个字段，Adapter 会自动 flush 拆分，避免 MongoDB `Conflicting mods` 错误。

5. **关闭顺序与阻塞语义**：`Adapter.Close()` / `Collector.Close()` / `CollectorWorker.Close()` 现在会**阻塞等待**剩余数据刷出（`<-done`），可重复调用。须**先关所有 Adapter 再关 Collector**——反之 Adapter 最后一批会因 Collector 已关闭被丢弃（日志 `Write After Closed`）。`BulkWrite` 失败会按 `maxRetries` 重试，重试耗尽仅记录丢弃模型、不保证落库，故 Close 返回不代表数据已落库。

6. **ListBase/MapBase 值解码先过 `canBaseDecode` 源类型门**：仅源 bson 类型为 Int32/Int64/Double/String 才解码，其余（Boolean、null/undefined 等）被 `Skip` 跳过；数值目标类型间经 `BaseDecode` 宽松转换（int/uint/float 跨类型兼容），uint32/uint64 现已支持。

7. **MapBase/MapMessage null 值**：遇到单个 value 为 null/undefined（或不可解码类型）时 `Skip` 并 `continue`，**不再中止整个 map 解码**——该键在结果 map 中缺失（而非报错或丢弃后续键）。

8. **ListMessage 元素 key 固定为创建时下标**：`Remove(index)` 后，后续元素的 key 不会重排，再对它们 `Set` 会写到错误的数组下标路径。删除后需重建列表或整体重写。

9. **Collector 懒加载现用二次检查**：`getOrCreateWorker` 在写锁内二次查表，并发首次写同一张表只会创建一个 worker（此前无二次检查会泄漏 goroutine）。worker 被并发关闭后，`Write` 会被丢弃并告警，不会 panic。

10. **同名结构体冲突**：`decodeMap` 以不含包名的结构体名为键，跨包同名 message 会覆盖彼此的解码器。

11. **MapBaseIncr/Decr 返回新值**，与生成代码 `IncrXxx/DecrXxx` 一致（此前返回旧值，已修复）。