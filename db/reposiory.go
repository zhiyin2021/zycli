package db

import (
	"fmt"
	"log"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/zhiyin2021/zycli/cmd"
	"gorm.io/gorm"
)

var (
	_db *gorm.DB
)

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
	if id > -1 {
		return Equal(id, "id")
	}
	return noneOpt()
}

func WithStatus(status int) Option {
	if status > -1 {
		return Equal(status, "status")
	}
	return noneOpt()
}

func WithUserId(userId int) Option {
	if userId > 0 {
		return Equal(userId, "userId")
	}
	return noneOpt()
}

func WithPreId(preId string) Option {
	if preId != "" {
		return StartWith(preId, "preId")
	}
	return noneOpt()
}

// 等于
func IfCall(flag bool, call func() Option) Option {
	if flag {
		return call()
	}
	return noneOpt()
}

// 等于
func Equal(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("=", val, fields...)
		return db.Where(f, v...)
	}
}

// 不等于
func NEqual(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("!=", val, fields...)
		return db.Where(f, v...)
	}
}

// 小于
func LT(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("<", val, fields...)
		return db.Where(f, v...)
	}
}

// 大于
func GT(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal(">", val, fields...)
		return db.Where(f, v...)
	}
}

// 小于等于
func LTE(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("<=", val, fields...)
		return db.Where(f, v...)
	}
}

// 大于等于
func GTE(val any, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
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
		f, v := genKeyVal("like", vv, fields...)
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
		where += v + expression + " ? "
		vals[i] = val
	}
	return
}
func List[T any](options ...Option) (o []T, err error) {
	err = GetQuery(options...).Find(&o).Error
	return
}

func Detail[T any](options ...Option) (o T, err error) {
	err = GetQuery(options...).First(&o).Error
	return
}

func Add(model any) (err error) {
	log.Println("AddByEntities", model)
	err = GetDB().Create(model).Error
	return
}

type Repository[T any] struct {
	runOne sync.Once
	log    *logrus.Entry
}

func (r *Repository[T]) Log() *logrus.Entry {
	r.runOne.Do(func() {
		r.log = logrus.WithFields(logrus.Fields{})
	})
	return r.log //.Debug()
}
func (r *Repository[T]) ToPageList(page, limit int, options ...Option) (o []*T, total int64, err error) {
	qry := GetQuery(options...)
	var obj T
	qry.Model(&obj).Count(&total)
	if page < 1 {
		page = 1
	}
	page--
	err = qry.Offset(page * limit).Limit(limit).Find(&o).Error
	return
}

func GetQuery(options ...Option) *gorm.DB {
	db := GetDB()
	for _, opt := range options {
		db = opt(db)
	}
	return db
}

func (r *Repository[T]) ToList(options ...Option) (o []*T, err error) {
	err = GetQuery(options...).Find(&o).Error
	return
}

func (r *Repository[T]) Get(id int) (o T, err error) {
	err = GetDB().Where("id=?", id).First(&o).Error
	return
}

func (r *Repository[T]) GetBy(options ...Option) (o T, err error) {
	err = GetQuery(options...).First(&o).Error
	return
}
func (r *Repository[T]) Add(model map[string]any) (err error) {
	var m T
	err = GetDB().Model(&m).Create(model).Error
	return
}

func (r *Repository[T]) Update(id int, model map[string]any) *gorm.DB {
	var m T
	return GetDB().Model(&m).Where("id=?", id).Updates(model)
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
