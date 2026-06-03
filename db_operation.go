package common

import (
	"context"

	"gorm.io/gorm"
)

type DbOperation struct {
	TableName string
	// Module, when non-nil, is passed to GORM Model() for QueryCV/Query instead of Table(TableName).
	Module  any
	Db      *gorm.DB
	Context context.Context
}

func (op *DbOperation) gormForTable() *gorm.DB {
	db := op.Db.WithContext(op.Context)
	if op.Module != nil {
		return db.Model(op.Module)
	}
	return db.Table(op.TableName)
}

// 带条件查询，
//
// 1. option是查询条件；
// 2. result是查询结果的存放地；
func (op *DbOperation) Query(option *SqlQueryOptions, result any) (err error) {
	return op.QueryCV(option, nil, result)
}

// 带条件查询，
//
// 1. option是查询条件；
// 2. total是查询结果的总数；
// 3. result是查询结果的存放地；
func (op *DbOperation) QueryCV(option *SqlQueryOptions, total *int64, result any) (err error) {
	tbOrg := op.gormForTable()
	GenMysqlWhere(tbOrg, option.QueryFields)
	// for _, where := range option.QueryFields {
	// 	tbOrg.Where(where.Query, where.Args...)
	// }

	for _, join := range option.Joins {
		tbOrg.Joins(join.Query, join.Args...)
	}
	// total!=nil 表示我想要总数；
	// 如果此时Limit为0，表示如果此时的sql语句会查询所有行，就没有必要去count了；
	// 此时如果rsult为nil，表示我不想查结果，那就单纯查count。所以又要count了；
	if total != nil && (option.Limit > 0 || result == nil) {
		tbOrg.Count(total)
		if *total == 0 {
			return tbOrg.Error
		}
	}
	if result == nil {
		return nil
	}
	for _, group := range option.GroupBy {
		tbOrg.Group(group)
	}

	tbOrg.Select(option.SelectFields)
	if option.Limit > 0 {
		tbOrg.Limit(option.Limit)
	}
	if option.Offset > 0 { // offset 0 写不写效果相同，就少写一句，减少gorm拼接sql和数据库服务器的解析工作
		tbOrg.Offset(option.Offset)
	}
	if len(option.OrderFields) > 0 {
		for _, order := range option.OrderFields {
			if order.Direction == ASCStr {
				tbOrg.Order(order.Field + " ASC")
			} else {
				tbOrg.Order(order.Field + " DESC")
			}
		}
	}
	tbOrg.Find(result)
	err = tbOrg.Error
	return
}
func (op *DbOperation) Update(option *SqlUpdateOptions) (err error) {
	tb := op.gormForTable()
	GenMysqlWhere(tb, option.QueryFields)
	tb.Updates(option.Updates)
	return tb.Error
}

func (op *DbOperation) Delete(option []Optioner) (err error) {
	tb := op.gormForTable()
	GenMysqlWhere(tb, option)
	tb.Delete(nil)
	return tb.Error
}

// Create 新增
func (op *DbOperation) Create(obj any) (err error) {
	tb := op.gormForTable()
	tb.Create(obj)
	return tb.Error
}
