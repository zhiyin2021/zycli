package db

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/zhiyin2021/zycli/cmd"
	"gorm.io/gorm"
)

var (
	_db *gorm.DB
)

type IEntities interface {
	GetId() int
}

func SetDB(db *gorm.DB) {
	_db = db
}
func GetDB() *gorm.DB {
	if cmd.DEBUG {
		return _db.Debug()
	}
	return _db
}

type Option func(*gorm.DB) *gorm.DB

func WithPage(pageIndex int, pageSize int) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((pageIndex - 1) * pageSize).Limit(pageSize)
	}
}
func WithOrder(order string) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(order)
	}
}

func WithID(id int) Option {
	if id > 0 {
		return Equal(id, "id")
	}
	return noneOpt()
}

// 等于
func IfCall(flag bool, call Option) Option {
	if flag {
		return call
	}
	return noneOpt()
}

// 等于
func Equal(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if !checkParam(val, fields...) {
			return db
		}
		f, v := genKeyVal("=", val, fields...)
		return db.Where(f, v...)
	}
}
func Between(val1, val2 any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val1 == "" || val2 == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal(" between ", val1, fields...)
		f += " and ?"
		v = append(v, val2)
		return db.Where(f, v...)
	}
}

// 不等于
func NEqual(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if !checkParam(val, fields...) {
			return db
		}
		f, v := genKeyVal("!=", val, fields...)
		return db.Where(f, v...)
	}
}
func checkParam(val any, fields ...string) bool {
	if len(fields) == 0 {
		return false
	}
	ele := reflect.ValueOf(val)
	switch v := val.(type) {
	case string:
		if v == "" {
			return false
		}
	case int, int32, int64, int8, int16:
		if ele.Int() < 0 {
			return false
		}
	case uint, uint8, uint16, uint32, uint64:
		if ele.Uint() == 0 {
			return false
		}
	case float32, float64:
		if ele.Float() < 0 {
			return false
		}
	}
	return true
}

// 小于
func LT(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if !checkParam(val, fields...) {
			return db
		}
		f, v := genKeyVal("<", val, fields...)
		return db.Where(f, v...)
	}
}

// 大于
func GT(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if !checkParam(val, fields...) {
			return db
		}
		f, v := genKeyVal(">", val, fields...)
		return db.Where(f, v...)
	}
}

// 小于等于
func LTE(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if !checkParam(val, fields...) {
			return db
		}
		f, v := genKeyVal("<=", val, fields...)
		return db.Where(f, v...)
	}
}

// 大于等于
func GTE(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if !checkParam(val, fields...) {
			return db
		}
		f, v := genKeyVal(">=", val, fields...)
		return db.Where(f, v...)
	}
}

// 起始包含
func StartWith(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		vv := fmt.Sprintf("%v%s", val, "%")
		f, v := genKeyVal(" like", vv, fields...)
		return db.Where(f, v...)
	}
}

// 结束包含
func EndWith(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		vv := fmt.Sprintf("%s%v", "%", val)
		f, v := genKeyVal(" like ", vv, fields...)
		return db.Where(f, v...)
	}
}

// 包含
func Contains(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		vv := fmt.Sprintf("%s%v%s", "%", val, "%")
		f, v := genKeyVal(" like ", vv, fields...)
		return db.Where(f, v...)
	}
}
func noneOpt() Option {
	return func(db *gorm.DB) *gorm.DB {
		return db
	}
}
func genKeyVal(expression string, val any, fields ...string) (where string, vals []any) {
	vals = make([]any, len(fields))
	for i, v := range fields {
		if where != "" {
			where += " or "
		}
		where += v + expression + "?"
		vals[i] = val
	}
	return
}
func Count[T IEntities](options ...Option) (total int64) {
	db := GetQuery(options...)
	var s T
	db.Model(&s).Count(&total)
	return
}

func ToList[T IEntities](options ...Option) (o []T, err error) {
	err = GetQuery(options...).Find(&o).Error
	return
}
func ToPageList[T IEntities](page, limit int, options ...Option) (o []T, total int64, err error) {
	qry := GetQuery(options...)
	var obj T
	if page < 1 {
		page = 1
	}
	page--
	err = qry.Model(obj).Count(&total).Offset(page * limit).Limit(limit).Find(&o).Error
	return
}

func Get[T IEntities](id int) (T, error) {
	var o *T
	err := GetDB().Where("id=?", id).First(&o).Error
	return *o, err
}
func GetBy[T IEntities](options ...Option) (T, error) {
	var o *T
	err := GetQuery(options...).First(&o).Error
	return *o, err
}

func Add(model IEntities) (err error) {
	log.Println("AddByEntities", model)
	err = GetDB().Create(model).Error
	return
}

type Repository[T IEntities] struct {
	runOne sync.Once
}

func (r *Repository[T]) ToPageList(page, limit int, options ...Option) (o []T, total int64, err error) {
	return ToPageList[T](page, limit, options...)
}

func GetQuery(options ...Option) *gorm.DB {
	db := GetDB()
	for _, opt := range options {
		db = opt(db)
	}
	return db
}

func (r *Repository[T]) ToList(options ...Option) (o []T, err error) {
	return ToList[T](options...)
}

func (r *Repository[T]) Get(id int) (T, error) {
	return Get[T](id)
}

func (r *Repository[T]) GetBy(options ...Option) (T, error) {
	return GetBy[T](options...)
}
func (r *Repository[T]) Add(model IEntities) (err error) {
	err = GetDB().Create(model).Error
	return
}

func (r *Repository[T]) AddMap(model map[string]any) (err error) {
	var m T
	err = GetDB().Model(&m).Create(model).Error
	return
}
func (r *Repository[T]) Update(model IEntities) error {
	var m T
	return GetDB().Model(&m).Where("id=?", model.GetId()).Updates(model).Error
}
func (r *Repository[T]) AddOrUpdate(model IEntities) error {
	if model.GetId() > 0 {
		return r.Update(model)
	} else {
		return r.Add(model)
	}
}
func (r *Repository[T]) Delete(id int) error {
	var m T
	return GetDB().Delete(&m, "id=?", id).Error
}
func (r *Repository[T]) DeleteBy(options ...Option) error {
	var m T
	return GetQuery(options...).Delete(&m).Error
}

func (r *Repository[T]) Count(options ...Option) (total int64) {
	db := GetQuery(options...)
	var s T
	db.Model(&s).Count(&total)
	return
}
