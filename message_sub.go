package mc

// MessageSub 子级消息结构体，用于表示嵌套在其他消息中的子消息
type MessageSub struct {
    parent IMessage // parent 父级消息对象
    key    any      // key 当前消息在父级中的键
}

// MessageType 返回消息类型
func (x *MessageSub) MessageType() MessageType {
    return MessageTypeSub
}

// Key 返回当前消息的键
func (x *MessageSub) Key() any {
    return x.key
}

// Set 设置字段值，将当前消息添加到嵌套消息链
// self 当前消息对象
// key 要设置的字段键
// val 要设置的值
// nestedMessages 嵌套消息列表
func (x *MessageSub) Set(self IMessage, key, val any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)          // 将当前消息添加到嵌套消息链
    x.parent.Set(x.parent, key, val, nestedMessages...) // 委托给父级设置
}

// Unset 取消设置字段值，将当前消息添加到嵌套消息链
// self 当前消息对象
// key 要取消设置的字段键
// nestedMessages 嵌套消息列表
func (x *MessageSub) Unset(self IMessage, key any, nestedMessages ...IMessage) {
    nestedMessages = append(nestedMessages, x)       // 将当前消息添加到嵌套消息链
    x.parent.Unset(x.parent, key, nestedMessages...) // 委托给父级取消设置
}

// Bind 绑定父级消息对象和键
// parent 父级消息对象
// key 当前消息在父级中的键
func (x *MessageSub) Bind(parent IMessage, key any) {
    x.parent = parent // 设置父级消息对象
    x.key = key       // 设置当前消息的键
}
