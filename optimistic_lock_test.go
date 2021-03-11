package gorm_optimistic

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestUpdateWithOptimistic(t *testing.T) {
	dsn := "root:abc123@/test?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	var out Optimistic
	db.First(&out, Optimistic{Id: 1})
	out.Amount = out.Amount + 10
	err = UpdateWithOptimistic(db, &out, nil, 0)
}

func BenchmarkUpdateWithOptimistic(b *testing.B) {
	dsn := "root:abc123@/test?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	b.RunParallel(func(pb *testing.PB) {
		var out Optimistic
		db.First(&out, Optimistic{Id: 1})
		out.Amount = out.Amount + 10
		err = UpdateWithOptimistic(db, &out, func(model OptimisticLock) OptimisticLock {
			bizModel := model.(*Optimistic)
			bizModel.Amount = bizModel.Amount + 10
			return bizModel
		}, 3)
		if err != nil {
			fmt.Println(err)
		}
	})
}

type Optimistic struct {
	Id      int64   `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	UserId  string  `gorm:"column:user_id;default:0;NOT NULL" json:"user_id"` // 用户ID
	Amount  float32 `gorm:"column:amount;NOT NULL" json:"amount"`             // 金额
	Version int64   `gorm:"column:version;default:0;NOT NULL" json:"version"` // 版本
}

func (o *Optimistic) TableName() string {
	return "t_optimistic"
}

func (o *Optimistic) GetVersion() int64 {
	return o.Version
}

func (o *Optimistic) SetVersion(version int64) {
	o.Version = version
}

func TestReflect(t *testing.T) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Nanosecond, // 慢 SQL 阈值
			LogLevel:      logger.Info,     // Log level
			Colorful:      false,           // 禁用彩色打印
		},
	)
	dsn := "root:abc123@/test?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		fmt.Println(err)
		return
	}
	var out Optimistic
	db.First(&out, Optimistic{Id: 1})

	//reflect.ValueOf(&out).Elem().FieldByName("Amount").SetFloat(99.00)
	//db.Updates(&out)
	testDB(db, &out)
}

func testDB(db *gorm.DB, model OptimisticLock) {
	db.First(model)
	v := reflect.ValueOf(model)
	i := reflect.Indirect(v)
	t := i.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("field=%v \n", field)
	}
	name := reflect.ValueOf(model).Elem().FieldByName("Version")
	fmt.Println(name.Int())
}
