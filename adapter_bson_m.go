package mc

import "go.mongodb.org/mongo-driver/v2/bson"

// newAdapterBsonM 创建新的适配器BSON实例
// aggregateStage 聚合阶段
// m 聚合字段
func newAdapterBsonM(aggregateStage AggregateStage, m bson.M) IAdapterBsonM {
    return &adapterBsonM{
        aggregateStage: aggregateStage,
        m:              m,
    }
}

// adapterBsonM 适配器结构体，用于处理Mongo数据操作
type adapterBsonM struct {
    aggregateStage AggregateStage // aggregateStage 聚合阶段
    m              bson.M         // m 聚合字段
}

// AggregateStage 聚合阶段
func (a *adapterBsonM) AggregateStage() AggregateStage {
    return a.aggregateStage
}

// M 聚合字段
func (a *adapterBsonM) M() bson.M {
    return a.m
}
