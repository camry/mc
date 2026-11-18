# CLAUDE.md

本文件为 Claude Code（claude.ai/code）在本仓库中工作时提供指引。

## 项目概述

`github.com/camry/mc`（Mongo Collector）是一个 Go 库：调用方通过生成的 Go 结构体直接操作深层嵌套的 MongoDB 文档，无需手写 BSON 字段路径或查询。所有对外类型（`MessageTop`/`MessageSub`、`ListBase`/`ListMessage`、`MapBase`/`MapMessage`）均由外部插件 `protoc-gen-mc` 生成为 `*.pb.mc.go` 文件；`pb/a.proto` 与 `pb/a.pb.mc.go` 是参考示例。

`go.mod` 要求 Go 1.26。依赖：mongo-driver v2、`github.com/camry/g/v2`（`gerror`、`glog`）、`spf13/cast`、`x/exp`、testify（仅测试）。

## 常用命令

```bash
go build ./...
go vet ./...                 # 同时检查 copylocks（容器内嵌了互斥锁）

go test ./...                # 注意：*_z_test.go 会连接真实 MongoDB（见下）
go test -run 'TestMap' ./... # 跑一组测试
go test -run 'TestMapBaseIncrConcurrent' -v ./...   # 跑单个测试

protoc --mc_out=./pb pb/*.proto   # 重新生成 pb/*.pb.mc.go
```

- 以 `_z_test.go` 结尾的文件（`mongo_z_test.go`、`adapter_z_test.go`、`mongo_z_table_test.go`）需要可用的 MongoDB；测试 URI 硬编码在 `mongo_z_test.go` 的 `defaultDB()` 中。不依赖数据库的纯测试：`utils_z_test.go`、`map_z_test.go`。
- `-race` 需要 cgo 和 C 编译器（gcc）；当前 Windows 环境未安装 mingw 工具链时不可用。

## 格式化 —— 不要跑 gofmt

`.editorconfig` 规定 **4 空格缩进**（`indent_size = 4`、`ij_go_run_go_fmt_on_reformat = false`）。`gofmt -l .` 会列出所有文件，这是仓库的既有状态，并非需要修复的问题。编辑时与周围代码的 4 空格风格保持一致，行宽上限 160。

## 架构

写入链路（库的核心）：

```
生成对象的 Set/Unset
  → 沿 nestedMessages 链向上冒泡（MessageSub/ListXxx/MapXxx 先把自身加入链）
  → MessageTop.Set → IAdapter.Set → 由嵌套键拼出字段路径（"b.c.d.username"）
  → Adapter.buffer（channel）→ 在每个 Adapter 内合并批量 $set/$unset
  → Collector.Write → 按表的 CollectorWorker buffer
  → BulkWrite（每个文档一条 UpdateOne，filter {_id: key}）
```

读取链路：`MongoDatabase.Find(top)` / `MongoTable.FindOne/FindMany` 读出 `bson.Raw`，经 `bson.NewDocumentReader` → `IDecode.Decode` → 回填结构体字段。解码器在 init 期用 `mc.Register` 按结构体名注册到包级全局映射，键是**不含包名的裸结构体名**（不同包的同名 message 会互相覆盖）。

关键组件：
- `define.go` —— 类型约束（`BaseType`、`MapKeyType`）、枚举、全部接口。
- `adapter*.go` —— 一个 Adapter 只绑定一个文档（顶级对象 + key）；聚合 `$set`/`$unset`，遇到父子键冲突（`"a"` 与 `"a.b"`）先 flush 再拆分。
- `collector*.go` —— 每张表懒创建一个 `CollectorWorker`（二次检查）；BulkWrite 带重试（`retryBaseDelay*2^(n-1)` 封顶 + 抖动）和按 worker 的批量刷新。
- `mongo*.go` —— client/database/collection 封装，以及只读的 `MongoTable[T,K]`（其中创建的对象**没有 Adapter**：`Set` 只改内存，`Key()` 返回 nil）。
- 容器写入粒度：列表操作整体重写数组（`$set x.key = Marshal()`）；Map 的 Put/Delete 只操作单个键（`$set`/`$unset path.key`），`Map.Clear` 整体重写映射。
- `utils_map.go` —— 原子的 `MapBaseIncr/MapBaseDecr`、`MapMessageGetOrPut`（均在锁内完成）。
- `Map` 内嵌 `sync.RWMutex`（并发读写安全；`Range` 遍历快照）。因此 `MapBase`/`MapMessage` 绝不可按值拷贝 —— 始终使用指针。

生命周期规则：每个文档需要独立的 `Adapter`；必须**先关闭所有 Adapter，再关闭 Collector**（`Adapter.Close()` → `Collector.Close()`），二者都会阻塞直到缓冲写入刷出。BulkWrite 失败会重试，重试耗尽后仅记录日志并丢弃 —— `Close` 成功不代表数据已落库。

深度参考：`.claude/skills/mc-knowledge/SKILL.md`（与代码保持同步；修改解码、Adapter 或容器行为前先查阅）。

## 工作约定

- 生成文件（`*.pb.mc.go`）不要手改；修改 `pb/*.proto` 后重新生成。
- `decodeMap` 的键是裸结构体名；在不同包新增同名 message 会导致解码错乱。
- 返回错误时的泛型零值写法：`var zero T; return zero, err`。
- `ListMessage` 元素的 key 是创建时的数组下标；`Remove(index)` 不会重排后续元素，之后再对它们 `Set` 会写到错误的下标路径。
- 测试位于外部包 `mc_test`，使用 testify。
- **plan 归档**：Claude Code 的 plan 文件固定在配置目录（`~/.claude/plans/`，平铺、不按项目分，且无法通过 settings 配置）。plan 获批后、开始改代码前，把 plan 正文归档到本仓库 `.claude/plans/<kebab-slug>.md`，用有意义的文件名（如 `map-concurrent-read-write.md`），不要用随机生成的名字。该目录被 `.claude/plans/.gitignore` 忽略，**plan 不提交到仓库**，仅供本地参考。
