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
func Equal(val interface{}, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("=", val, fields...)
		return db.Where(f, v...)
	}
}

// 不等于
func NEqual(val interface{}, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("!=", val, fields...)
		return db.Where(f, v...)
	}
}

// 小于
func LT(val interface{}, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("<", val, fields...)
		return db.Where(f, v...)
	}
}

// 大于
func GT(val interface{}, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal(">", val, fields...)
		return db.Where(f, v...)
	}
}

// 小于等于
func LTE(val interface{}, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal("<=", val, fields...)
		return db.Where(f, v...)
	}
}

// 大于等于
func GTE(val interface{}, fields ...string) Option {
	return func(db *gorm.DB) *gorm.DB {
		if val == "" || len(fields) == 0 {
			return db
		}
		f, v := genKeyVal(">=", val, fields...)
		return db.Where(f, v...)
	}
}

// 起始包含
func StartWith(val interface{}, fields ...string) Option {
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
func EndWith(val interface{}, fields ...string) Option {
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
func Contains(val interface{}, fields ...string) Option {
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
func genKeyVal(expression string, val interface{}, fields ...string) (where string, vals []interface{}) {
	vals = make([]interface{}, len(fields))
	for i, v := range fields {
		if where != "" {
			where += " or "
		}
		where += v + expression + " ? "
		vals[i] = val
	}
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

func GetQuery(options ...Option) *gorm.DB {
	db := GetDB()
	for _, opt := range options {
		db = opt(db)
	}
	return db
}

func (r *Repository[T]) ToList(options ...Option) (o []T) {
	db := GetQuery(options...)
	if err := db.Find(&o).Error; err != nil {
		logrus.Errorln("")
	}
	return
}

func (r *Repository[T]) FindById(id int) *T {
	var o T
	if err := GetDB().Where("id=?", id).First(&o).Error; err != nil {
		logrus.Errorln("FindById err", err)
		return nil
	}
	return &o
}

func (r *Repository[T]) FindBy(options ...Option) *T {
	db := GetQuery(options...)
	var o T
	if err := db.First(&o).Error; err != nil {
		logrus.Errorln("FindBy err", err)
		return nil
	}
	return &o
}

func (r *Repository[T]) FindByModel(model interface{}, options ...Option) error {
	db := GetQuery(options...)
	if err := db.First(model).Error; err != nil {
		logrus.Errorln("FindByModel err", err)
		return err
	}
	return nil
}
func (r *Repository[T]) AddByEntities(model interface{}) (err error) {
	log.Println("AddByEntities", model)
	err = GetDB().Create(model).Error
	return
}
func (r *Repository[T]) Add(model map[string]interface{}) (err error) {
	var m T
	err = GetDB().Model(&m).Create(model).Error
	return
}

func (r *Repository[T]) Update(id int, model map[string]interface{}) *gorm.DB {
	var m T
	return GetDB().Model(&m).Where("id=?", id).Updates(model)
}

func (r *Repository[T]) Delete(id int) error {
	var m T
	return GetDB().Delete(&m, "id=?", id).Error
}

func (r *Repository[T]) Count(options ...Option) (total int64) {
	db := GetQuery(options...)
	var s T
	db.Model(&s).Count(&total)
	return
}
