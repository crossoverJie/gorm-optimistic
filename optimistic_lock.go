package gorm_optimistic

import (
	"gorm.io/gorm"
)

type OptimisticLock interface {
	SetVersion(version int64)
	GetVersion() int64
}
type Version struct {
	Version int64 `gorm:"column:version;default:0;NOT NULL" json:"version"` // version
}

func UpdateWithOptimistic(db *gorm.DB, model OptimisticLock, callBack func(model OptimisticLock) OptimisticLock) (err error) {
	currentVersion := model.GetVersion()
	model.SetVersion(currentVersion + 1)
	column := db.Model(model).Where("version", currentVersion).UpdateColumns(model)
	affected := column.RowsAffected
	if affected == 0 {
		if callBack == nil {
			return column.Error
		}
		db.First(model)
		bizModel := callBack(model)
		err := UpdateWithOptimistic(db, bizModel, callBack)
		if err != nil {
			return err
		}
	}
	return column.Error

}
