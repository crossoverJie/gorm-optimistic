package optimistic

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Lock interface {
	SetVersion(version int64)
	GetVersion() int64
}
type Version struct {
	Version int64 `gorm:"column:version;default:0;NOT NULL" json:"version"` // version
}

type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}
func NewOptimisticError(msg string) *Error {
	return &Error{msg}
}

func UpdateWithOptimistic(db *gorm.DB, model Lock, callBack func(model Lock) Lock, retryCount, currentRetryCount int32) (err error) {
	if currentRetryCount > retryCount {
		return errors.WithStack(NewOptimisticError("Maximum number of retries exceeded:" + strconv.Itoa(int(retryCount))))
	}
	currentVersion := model.GetVersion()
	model.SetVersion(currentVersion + 1)
	column := db.Model(model).Where("version", currentVersion).UpdateColumns(model)
	if column.Error != nil {
		return errors.WithStack(column.Error)
	}
	affected := column.RowsAffected
	if affected == 0 {
		if callBack == nil && retryCount == 0 {
			return errors.WithStack(NewOptimisticError("Concurrent optimistic update error"))
		}
		time.Sleep(100 * time.Millisecond)
		first := db.First(model)
		if first.Error != nil {
			return errors.WithStack(first.Error)
		}
		bizModel := callBack(model)
		currentRetryCount++
		err := UpdateWithOptimistic(db, bizModel, callBack, retryCount, currentRetryCount)
		if err != nil {
			return err
		}
	}
	return nil

}
