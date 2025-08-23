package common

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

type Direction string

const (
	ASCStr  Direction = "ASC"
	DESCStr Direction = "DESC"
)

const (
	ASC  = 1
	DESC = 2
)

type OrderByParam struct {
	Field     string
	Direction Direction
}

type OrderByParams []OrderByParam
type QueryOptions struct {
	SelectFields []string
	OmitFields   []string
	OrderFields  OrderByParams
	Limit        int
	Offset       int
}

type mysqlOption struct {
	Query string //缩小范围，仅支持字符串
	Args  []any
}

func (o *mysqlOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query, o.Args...)
}
func (o *mysqlOption) GenMongoOption(m bson.M) {
	// m[o.Query] = o.Args[0]
}

type Option = Optioner

func O(query string, args ...any) *mysqlOption {
	return &mysqlOption{Query: query, Args: args}
}
func W(query string, args ...any) *mysqlOption {
	return &mysqlOption{Query: query, Args: args}
}

type SqlQueryOptions struct {
	QueryFields  []Optioner
	SelectFields []string
	// OmitFields   []string
	Joins       []*mysqlOption
	GroupBy     []string
	OrderFields OrderByParams
	Limit       int
	Offset      int
}

type SqlUpdateOptions struct {
	QueryFields []Optioner
	Updates     any
}

type MongoQueryOptions struct {
	QueryFields  bson.M
	SelectFields bson.M
	OrderFields  bson.M
	Limit        int
	Offset       int
}

type DbOperation struct {
	TableName string
	Db        *gorm.DB
	Context   context.Context
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
	tbOrg := op.Db.WithContext(op.Context).Table(op.TableName)
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
	tb := op.Db.WithContext(op.Context).Table(op.TableName)
	GenMysqlWhere(tb, option.QueryFields)
	tb.Updates(option.Updates)
	return tb.Error
}

func (op *DbOperation) Delete(option []Optioner) (err error) {
	tb := op.Db.WithContext(op.Context).Table(op.TableName)
	GenMysqlWhere(tb, option)
	tb.Delete(nil)
	return tb.Error
}

// Create 新增
func (op *DbOperation) Create(obj any) (err error) {
	tb := op.Db.WithContext(op.Context).Table(op.TableName)
	tb.Create(obj)
	return tb.Error
}

type OptionType int

const (
	OptionTypeEq OptionType = iota
	OptionTypeNe
)

type TypeOption struct {
	Query string //缩小范围，仅支持字符串
	Args  any
}

type EqOption TypeOption
type NeOption TypeOption
type InOption TypeOption
type GtOption TypeOption
type GteOption TypeOption
type LtOption TypeOption
type LteOption TypeOption

// ExistOption 表示MongoDB的$exists操作符，用于检查字段是否存在
type ExistOption TypeOption

// LikeOption 表示模糊匹配操作符，用于SQL的LIKE查询和MongoDB的正则查询
type LikeOption TypeOption

func Eq(field string, value any) EqOption {
	return EqOption{Query: field, Args: value}
}

func Ne(field string, value any) NeOption {
	return NeOption{Query: field, Args: value}
}

func In(field string, value any) InOption {
	return InOption{Query: field, Args: value}
}

func Gt(field string, value any) GtOption {
	return GtOption{Query: field, Args: value}
}

func Gte(field string, value any) GteOption {
	return GteOption{Query: field, Args: value}
}

func Lt(field string, value any) LtOption {
	return LtOption{Query: field, Args: value}
}

func Lte(field string, value any) LteOption {
	return LteOption{Query: field, Args: value}
}

// Exist 创建一个检查字段是否存在的条件
// field: 字段名
// exists: true表示字段必须存在，false表示字段必须不存在
func Exist(field string, exists bool) ExistOption {
	return ExistOption{Query: field, Args: exists}
}

// Like 创建一个模糊匹配条件
// field: 字段名
// value: 匹配值（应包含适当的通配符，如%或_）
func Like(field string, value any) LikeOption {
	return LikeOption{Query: field, Args: value}
}

type MongoOptioner interface {
	GenMongoOption(m bson.M)
}
type Optioner interface {
	MongoOptioner
	MysqlOptioner
}

// GenMongoOption 实现MongoOptioner接口，将EqOption转换为MongoDB的等于查询
func (o EqOption) GenMongoOption(m bson.M) {
	m[o.Query] = o.Args
}

// GenMongoOption 实现MongoOptioner接口，将NeOption转换为MongoDB的不等于查询
func (o NeOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$ne": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将InOption转换为MongoDB的in查询
func (o InOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$in": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将GtOption转换为MongoDB的大于查询
func (o GtOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$gt": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将GteOption转换为MongoDB的大于等于查询
func (o GteOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$gte": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将LtOption转换为MongoDB的小于查询
func (o LtOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$lt": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将LteOption转换为MongoDB的小于等于查询
func (o LteOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$lte": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将ExistOption转换为MongoDB的$exists查询
func (o ExistOption) GenMongoOption(m bson.M) {
	m[o.Query] = bson.M{"$exists": o.Args}
}

// GenMongoOption 实现MongoOptioner接口，将LikeOption转换为MongoDB的正则查询
func (o LikeOption) GenMongoOption(m bson.M) {
	// 在MongoDB中使用正则表达式实现LIKE功能
	// 自动添加前后通配符
	if pattern, ok := o.Args.(string); ok {
		// 自动添加前后通配符，除非已经包含
		regexPattern := pattern
		if !strings.HasPrefix(pattern, ".*") && !strings.HasPrefix(pattern, "%") {
			regexPattern = ".*" + regexPattern
		}
		if !strings.HasSuffix(pattern, ".*") && !strings.HasSuffix(pattern, "%") {
			regexPattern = regexPattern + ".*"
		}
		// 将SQL LIKE通配符转换为JavaScript正则表达式格式
		regexPattern = bsonRegexEscapeForLike(regexPattern)
		m[o.Query] = bson.M{"$regex": regexPattern, "$options": "i"}
	} else {
		// 如果不是字符串，直接使用原始值并添加通配符
		regexPattern := fmt.Sprintf(".*%v.*", o.Args)
		m[o.Query] = bson.M{"$regex": regexPattern, "$options": "i"}
	}
}

// bsonRegexEscapeForLike 转义MongoDB正则表达式中的特殊字符并将SQL通配符转换为正则表达式
func bsonRegexEscapeForLike(pattern string) string {
	// 先转义正则特殊字符
	escaped := pattern
	// 转义正则表达式特殊字符
	specialChars := []string{`\`, `^`, `$`, `.`, `+`, `*`, `?`, `(`, `)`, `[`, `]`, `{`, `}`, `|`}
	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, `\`+char)
	}

	// 将SQL LIKE通配符转换为正则表达式（如果有的话）
	escaped = strings.ReplaceAll(escaped, `%`, `.*`)
	escaped = strings.ReplaceAll(escaped, `_`, `.`)

	return escaped
}

// OrOption 表示MongoDB的$or操作符，用于组合多个查询条件（任一条件满足即可）
type OrOption struct {
	Options []Optioner
}

// AndOption 表示MongoDB的$and操作符，用于组合多个查询条件（所有条件都必须满足）
type AndOption struct {
	Options []Optioner
}

// Or 创建一个OR条件，满足任一条件即可
func Or(options ...Optioner) OrOption {
	return OrOption{Options: options}
}

// And 创建一个AND条件，必须满足所有条件
func And(options ...Optioner) AndOption {
	return AndOption{Options: options}
}

// GenMongoOption 实现MongoOptioner接口，将OrOption转换为MongoDB的$or查询
func (o OrOption) GenMongoOption(m bson.M) {
	var conditions []bson.M
	for _, option := range o.Options {
		condition := make(bson.M)
		option.GenMongoOption(condition)
		conditions = append(conditions, condition)
	}
	m["$or"] = conditions
}

// GenMongoOption 实现MongoOptioner接口，将AndOption转换为MongoDB的$and查询
func (o AndOption) GenMongoOption(m bson.M) {
	var conditions []bson.M
	for _, option := range o.Options {
		condition := make(bson.M)
		option.GenMongoOption(condition)
		conditions = append(conditions, condition)
	}
	m["$and"] = conditions
}

func GenMongoOption(option []Optioner) bson.M {
	m := make(bson.M)
	for _, v := range option {
		v.GenMongoOption(m)
	}
	return m
}

type MysqlOptioner interface {
	GenMysqlWhere(db *gorm.DB)
}

// GenMysqlWhere 实现MysqlOptioner接口，将EqOption转换为MySQL的等于查询
func (o EqOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" = ?", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将NeOption转换为MySQL的不等于查询
func (o NeOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" <> ?", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将InOption转换为MySQL的IN查询
func (o InOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" IN (?)", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将GtOption转换为MySQL的大于查询
func (o GtOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" > ?", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将GteOption转换为MySQL的大于等于查询
func (o GteOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" >= ?", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将LtOption转换为MySQL的小于查询
func (o LtOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" < ?", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将LteOption转换为MySQL的小于等于查询
func (o LteOption) GenMysqlWhere(db *gorm.DB) {
	db.Where(o.Query+" <= ?", o.Args)
}

// GenMysqlWhere 实现MysqlOptioner接口，将ExistOption转换为MySQL的IS NULL/IS NOT NULL查询
func (o ExistOption) GenMysqlWhere(db *gorm.DB) {
	if o.Args.(bool) {
		db.Where(o.Query + " IS NOT NULL")
	} else {
		db.Where(o.Query + " IS NULL")
	}
}

// GenMysqlWhere 实现MysqlOptioner接口，将LikeOption转换为MySQL的LIKE查询
func (o LikeOption) GenMysqlWhere(db *gorm.DB) {
	// 自动添加前后%通配符
	if pattern, ok := o.Args.(string); ok {
		// 如果用户没有自己添加通配符，则自动添加前后%
		if !strings.ContainsAny(pattern, "%_") {
			db.Where(o.Query+" LIKE ?", "%"+pattern+"%")
		} else {
			// 如果用户已经添加了通配符，则按用户指定的执行
			db.Where(o.Query+" LIKE ?", pattern)
		}
	} else {
		// 如果不是字符串，直接添加前后通配符
		db.Where(o.Query+" LIKE ?", fmt.Sprintf("%%%v%%", o.Args))
	}
}

// GenMysqlWhere 实现MysqlOptioner接口，将OrOption转换为MySQL的OR查询
func (o OrOption) GenMysqlWhere(db *gorm.DB) {
	for i, option := range o.Options {
		if i == 0 {
			option.GenMysqlWhere(db)
		} else {
			db.Or(func(orDb *gorm.DB) {
				option.GenMysqlWhere(orDb)
			})
		}
	}
}

// GenMysqlWhere 实现MysqlOptioner接口，将AndOption转换为MySQL的AND查询
func (o AndOption) GenMysqlWhere(db *gorm.DB) {
	for _, option := range o.Options {
		option.GenMysqlWhere(db)
	}
}

// 添加一个辅助函数，用于生成MySQL查询条件
func GenMysqlWhere(db *gorm.DB, options []Optioner) *gorm.DB {
	for _, option := range options {
		option.GenMysqlWhere(db)
	}
	return db
}
