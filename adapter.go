package mc

import (
    "github.com/camry/g/glog"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// Adapter 适配器结构体，用于处理Mongo数据操作
// T 代表实现了IMessageTop接口的消息类型
// K 代表可比较的键类型
type Adapter[T IMessageTop, K comparable] struct {
    t                     T                       // t 当前绑定的消息对象
    k                     K                       // k 当前绑定的键值
    collector             ICollector              // collector 收集器实例
    collectorWorkerOption []CollectorWorkerOption // collectorWorkerOption 收集器工作选项
}

// NewAdapter 创建新的适配器实例
// db Mongo数据库实例
// opts 适配器选项
func NewAdapter[T IMessageTop, K comparable](opts ...AdapterOption[T, K]) *Adapter[T, K] {
    x := &Adapter[T, K]{}
    for _, opt := range opts {
        opt(x)
    }
    return x
}

// Bind 绑定消息对象和键值
// t 顶级消息对象
// k 键值
func (x *Adapter[T, K]) Bind(t T, k K) {
    x.t = t
    x.k = k
}

// GetMessageTop 获取绑定的顶级消息对象
func (x *Adapter[T, K]) GetMessageTop() T {
    return x.t
}

// GetKey 获取绑定的键值
func (x *Adapter[T, K]) GetKey() K {
    return x.k
}

// Set 设置字段值
// nestedMessages 嵌套消息列表
// key 字段键
// val 字段值
func (x *Adapter[T, K]) Set(nestedMessages []IMessage, key, val any) {
    mk := toNestedMessageFieldKey(nestedMessages, key) // 构建嵌套消息字段键
    if x.collector == nil {
        glog.Errorf("Adapter Set Failed Collector Not Initialized For FieldKey:%s FieldValue:%v", mk, val)
        return
    }
    x.collector.Write(x.t.Table(), mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": x.GetKey()}).SetUpdate(bson.M{
        "$set": bson.M{
            mk: val, // 使用$set操作符设置字段值
        },
    }), x.collectorWorkerOption...)
    glog.Debugf("Adapter Set Table=%s _id=%v FieldKey:%s FieldValue:%v", x.t.Table(), x.GetKey(), mk, val)
}

// Unset 取消设置字段值
// nestedMessages 嵌套消息列表
// key 要取消设置的字段键
func (x *Adapter[T, K]) Unset(nestedMessages []IMessage, key any) {
    mk := toNestedMessageFieldKey(nestedMessages, key) // 构建嵌套消息字段键
    if x.collector == nil {
        glog.Errorf("Adapter Unset Failed Collector Not Initialized For FieldKey:%s", mk)
        return
    }
    x.collector.Write(x.t.Table(), mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": x.GetKey()}).SetUpdate(bson.M{
        "$unset": bson.M{
            mk: "", // 使用$unset操作符取消设置字段值
        },
    }), x.collectorWorkerOption...)
    glog.Debugf("Adapter Unset Field Key:%s", mk)
}
