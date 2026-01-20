package mc

// MessageTop 顶级消息结构体，用于表示消息树的根节点
// T 顶级消息类型，必须实现 IMessageTop 接口
// K 键类型，必须是可比较的
type MessageTop[T IMessageTop, K comparable] struct {
    a IAdapter[T, K] // a 适配器，用于管理消息对象和键值对
}

// MessageType 返回消息类型
func (x *MessageTop[T, K]) MessageType() MessageType {
    return MessageTypeTop
}

// Set 设置字段值
// self 当前消息对象
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *MessageTop[T, K]) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    x.a.Set(x.a.GetMessageTop(), nestedMessages, key, val) // 通过适配器设置值
}

// Unset 取消设置字段值
// self 当前消息对象
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *MessageTop[T, K]) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    x.a.Unset(x.a.GetMessageTop(), nestedMessages, key) // 通过适配器取消设置值
}

// Bind 绑定适配器、消息对象和键
// adapter 适配器
// t 顶级消息对象
// k 键值
func (x *MessageTop[T, K]) Bind(adapter IAdapter[T, K], t T, k K) {
    x.a = adapter  // 设置适配器
    x.a.Bind(t, k) // 绑定消息对象和键
}

// Key 返回当前消息的键
func (x *MessageTop[T, K]) Key() any {
    return x.a.GetKey() // 通过适配器获取键
}
