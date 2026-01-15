package mc

import (
    "strings"

    "github.com/camry/g/glog"
    "github.com/spf13/cast"
)

func toNestedMessageFieldKey(nestedMessages []IMessage, key any) string {
    if len(nestedMessages) > 0 {
        var builder strings.Builder
        builder.WriteString(toNestedMessageKey(nestedMessages))
        builder.WriteString(".")
        builder.WriteString(cast.ToString(key))
        return builder.String()
    }
    return cast.ToString(key)
}

func toNestedMessageKey(nestedMessages []IMessage) string {
    if len(nestedMessages) == 0 {
        return ""
    }
    var builder strings.Builder
    for i := len(nestedMessages) - 1; i >= 0; i-- {
        message := nestedMessages[i]
        if builder.Len() > 0 {
            builder.WriteString(".")
        }
        switch message.MessageType() {
        case MessageTypeSub:
            builder.WriteString(cast.ToString(message.Key()))
        case MessageTypeMapBase, MessageTypeMapNested:
            glog.Errorf("MessageTypeMapBase or MessageTypeMapNested error: %v", message.MessageType())
        default:
            glog.Errorf("message type error: %v", message.MessageType())
        }
    }
    return builder.String()
}
