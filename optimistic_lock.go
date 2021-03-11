package optimistic

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"sync/atomic"
	"time"
)

type Lock interface {
	SetVersion(version int64)
	GetVersion() int64
}
type Version struct {
	Version int64 `gorm:"column:version;default:0;NOT NULL" json:"version"` // version
}

type OverRetryError struct {
	Msg string
}

func (e *OverRetryError) Error() string {
	return e.Msg
}
func NewOverRetryError(msg string) *OverRetryError {
	return &OverRetryError{msg}
}

func UpdateWithOptimistic(db *gorm.DB, model Lock, callBack func(model Lock) Lock, retryCount, currentRetryCount int32) (err error) {
	if currentRetryCount > retryCount {
		return errors.WithStack(NewOverRetryError("Maximum number of retries exceeded:" + strconv.Itoa(int(retryCount))))
	}
	currentVersion := model.GetVersion()
	model.SetVersion(currentVersion + 1)
	column := db.Model(model).Where("version", currentVersion).UpdateColumns(model)
	affected := column.RowsAffected
	if affected == 0 {
		if callBack == nil && retryCount == 0 {
			return column.Error
		}
		time.Sleep(100 * time.Millisecond)
		db.First(model)
		bizModel := callBack(model)
		atomic.AddInt32(&currentRetryCount, 1)
		err := UpdateWithOptimistic(db, bizModel, callBack, retryCount, currentRetryCount)
		if err != nil {
			return err
		}
	}
	return column.Error

}
