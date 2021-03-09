package gorm_optimistic

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
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
	err = UpdateWithOptimistic(db, &out)
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
