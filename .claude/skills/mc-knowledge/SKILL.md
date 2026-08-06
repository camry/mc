# mc 仓库知识库

**项目全名**: Mongo Collector  
**仓库路径**: `D:\project\meigushi\mc`  
**模块名**: `gitea.meigushi.go/meigushi/mc`  
**Go 版本**: 1.25+  
**依赖**: `go.mongodb.org/mongo-driver/v2`、`github.com/samber/lo`、`github.com/spf13/cast`

**定位**：简化 MongoDB 复杂嵌套文档操作的 Go 库。开发者像操作普通对象一样操作 MongoDB 文档，无需手动构建字段路径或 BSON 查询。适用于用户属性管理、配置存储、游戏数据存储等频繁操作嵌套文档的场景。

---

## 一、整体架构

mc 是一个 MongoDB 数据访问层框架，核心思路是：

```
MongoDB 文档
  ↓  FindOne / FindMany（MongoTable / MongoDatabase）
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

所有 struct 通过代码生成（`protoc-gen-mc`）产生，生成的文件以 `.pb.mc.go` 结尾。

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

---

## 三、消息层级结构

### 接口

| 接口 | 说明 |
|------|------|
| `IMessage` | 所有消息基础接口：`MessageType() / Key() / Set() / Unset()` |
| `IMessageTop` | 顶级消息，扩展 `IMessage + IMarshalBsonM`，增加 `Table() string` |
| `IMessageSub` | 子消息，扩展 `IMessage + IMarshalBsonM` |
| `IMarshalBsonM` | `Marshal() bson.M / Unmarshal(bson.M) error` |
| `IDecode` | `Decode(vr bson.ValueReader, v any) error` |

### 具体结构体

| 类型 | 说明 |
|------|------|
| `MessageTop[T, K]` | 顶级节点，持有 `IAdapter`；`Set/Unset` 转发给 Adapter |
| `MessageSub` | 子节点，持有 `parent IMessage + key any`；`Set/Unset` 追加自身到 nestedMessages 链后向上冒泡 |
| `List[T]` | 通用有序列表基类，实现 `Append/Remove/Clear/Contains/Find/Filter/ForEach/Values/Marshal` |
| `ListBase[T BaseType]` | 基础类型列表，实现 BSON 数组解码（`TypeArray`） |
| `ListMessage[T IMessage]` | 消息对象列表，带 `creator` 工厂；`AppendNewMessage()` 追加新对象 |
| `Map[T, V]` | 通用映射基类，实现 `Get/Put/Delete/Contains/Size/Keys/Values/GetAll/Range/Clear/Marshal` |
| `MapBase[T MapKeyType, V BaseType]` | 基础类型映射，实现 BSON document 解码 |
| `MapMessage[T MapKeyType, V IMessage]` | 消息对象映射，带 `creator` 工厂；`PutNewMessage(key)` 创建并存入 |

### 字段变更传播路径

```
叶子.Set(key, val)
  → MessageSub.Set：nestedMessages 追加 self，调 parent.Set
  → ... 逐层向上 ...
  → MessageTop.Set → Adapter.Set(nestedMessages, key, val)
  → toNestedMessageFieldKey(nestedMessages, key) 构造 "a.b.c.key"
  → write(adapterBsonM{$set, bson.M{"a.b.c.key": val}})
```

`toNestedMessageFieldKey` / `toNestedMessageKey` 在 `utils.go` 中，从 nestedMessages 尾部向前拼接点分隔的键路径。

---

## 四、解码系统

### 注册与查找

```go
// 生成代码中：
var xDecode = mc.Register(&X{}, mc.NewDecodePair("field", mc.NewDecodeField(...)), ...)

// 运行时查找：
decode, err := mc.GetDecode(obj)   // 按 struct 名称查 decodeMap
```

`Register` → `decodeMap[structName] = *DecodeMessage`

### DecodeMessage.Decode 流程

1. 若 `vr.Type() == TypeNull/Undefined` → 直接消费并返回
2. `vr.ReadDocument()` 开始读文档
3. 逐元素：找到 `fieldMap[key]` → `handler.Decode(evr, v)`；否则 `evr.Skip()`

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
| `BoolDecode` | Boolean |
| `StringDecode` | String |
| `BinaryDecode` | Binary / String(base64) / **Null** / **Undefined** |

> **注意**：`BoolDecode` 和 `StringDecode` 未处理 Null/Undefined——若字段在文档中为 null 会报错。`BinaryDecode` 已修复支持 Null/Undefined（返回 nil, nil）。

---

## 五、Adapter（写入聚合层）

```go
adapter := mc.NewAdapter[*MyMessage, int64](
    mc.WithCollector[*MyMessage, int64](collector),
    mc.WithCollectorWorkerOption[*MyMessage, int64](opts...),
    mc.WithContext[*MyMessage, int64](ctx),
)
```

**内部机制**：
- `buffer chan IAdapterBsonM`（默认容量 1000）
- goroutine `writeCollectorWorker`：ticker（默认 10ms）或 batchSize（默认 10）触发 flush
- `doBatchAdapterBsonM`：合并同阶段的 `bson.M`，检测父子键冲突（`HasPrefix`），冲突则先 flush 再合并
- 最终生成 `mongo.UpdateOneModel($set/$unset)` 发送给 `Collector`

**关闭顺序**：`adapter.Close()` → cancel context → goroutine 消费剩余 buffer → `collector.Write`

---

## 六、Collector / CollectorWorker（批量写入层）

```go
collector := mc.NewCollector(ctx, db)
```

- 每张表懒初始化一个 `CollectorWorker`（加读写锁保护）
- `CollectorWorker`：buffer（默认 10000）、batchSize（默认 100）、flushInterval（默认 100ms）
- `Run()` goroutine：ticker 或 batchSize 触发 `BulkWrite`
- `Close()` 并发关闭所有 worker

---

## 七、MongoClient / MongoDatabase / MongoCollection

三层封装，主要是注入 `context.Context`：

```
MongoClient.Database(name) → MongoDatabase
MongoDatabase.Collection(name) → MongoCollection
```

`MongoDatabase.Find(top IMessageTop)`：
- `FindOne(_id=top.Key())`
- 若 `ErrNoDocuments` → `InsertOne(top.Marshal())`（自动创建文档）
- 否则用 `decode.Decode` 反序列化

---

## 八、MongoTable[T, K]（只读查询）

```go
table := mc.NewMongoTable[*MyMessage, int64](db)

obj, err := table.FindOne(id)       // 查单条
list, err := table.FindMany(ids)    // 查多条（$in 查询，batchSize=2048）
```

**FindOne**：用 `var zero T` 作为泛型零值返回（不能用 nil）；反序列化用 `bson.NewDocumentReader(bytes.NewReader(bsonRaw))`。

**FindMany**：同样通过游标逐条解码；每次循环用 `reflect.New(t).Interface().(T)` 创建新对象。

---

## 九、代码生成惯例（`.pb.mc.go`）

生成文件的固定模式：

```go
// 1. 定义 struct（内嵌 mc.MessageTop[*T, K] 或 mc.MessageSub）
// 2. 注册解码器
var xDecode = mc.Register(&X{},
    mc.NewDecodePair("fieldName", mc.NewDecodeField(func(vr bson.ValueReader, v *X) (err error) {
        v.field, err = mc.Int64Decode(vr)  // 或其他 XxxDecode
        return
    })),
    ...
)
// 3. 构造函数 NewX(adapter, k)
// 4. Table() string
// 5. Marshal() bson.M
// 6. Unmarshal(bson.M) error
// 7. Getter/Setter（含 Incr/Decr for 数值）
// 8. 子消息用 newX(parent, key)，MessageSub 嵌入
```

子消息字段的 Getter 带懒初始化：
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

典型用法（来自 README）：
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

// 3. 创建对象并操作嵌套字段
a := pb.NewA(adapter, 2)
a.SetAUsername("camry")
a.GetB().GetC().GetD().SetId(math.MaxInt32)
a.GetB().GetC().GetD().SetBytes([]byte("hello"))
a.GetB().GetC().GetD().GetTags().Append("tag3")

// 4. 等待异步写入完成后关闭
time.Sleep(5 * time.Second)
collector.Close()
```

---

## 十一、常见陷阱与规则

1. **泛型零值**：`FindOne/FindMany` 的 error 返回必须用 `var zero T; return zero, err`，不能 `return nil, err`。

2. **BinaryDecode null**：`[]byte` 字段在 MongoDB 中可能为 null，`BinaryDecode` 已增加 `TypeNull/TypeUndefined` 处理。

3. **reflect.TypeFor[T]().Elem()**：比 `reflect.TypeOf(new(T)).Elem().Elem()` 更简洁，二者等价（需 Go 1.22+）。

4. **Adapter 无 Collector 时静默失败**：`Set/Unset` 若 `collector == nil` 只打 error 日志，不 panic。

5. **父子键冲突**：同一 batch 中若出现 `"a"` 和 `"a.b"` 两个字段，Adapter 会自动 flush 拆分，避免 MongoDB `Conflicting mods` 错误。

6. **ListBase 枚举解码**：`ListBase[AEnum]` 使用 `reflect.TypeFor[[]T]().Elem()` 获取底层 Kind，再 `reflect.ValueOf(val).Convert(targetType)` 完成类型转换。

7. **MapBase/MapMessage null 元素**：遇到单个 value 为 null/undefined 时直接 return（整个 map 解码中止），不是跳过——如需容忍可改为 `continue`。