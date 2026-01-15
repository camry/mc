package mc

import "github.com/camry/g/glog"

type Adapter[K comparable] struct {
}

func NewAdapter[K comparable](k K) *Adapter[K] {
    return &Adapter[K]{}
}

func (a *Adapter[K]) Collect(k K, nestedMessages []IMessage, key, val any) {
    glog.Infof("Adapter Collect Field Key:%s Value:%v", toNestedMessageFieldKey(nestedMessages, key), val)
}
