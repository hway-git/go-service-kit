package repository

import (
	"errors"
	"example.com/go-service-kit/db/mysql"
	"example.com/go-service-kit/utils/tool"

	"gorm.io/gorm"
)

type BaseRepository[T any] struct{}

// 新增
func (r *BaseRepository[T]) Create(model *T) error {
	if tx := mysql.DB.Create(model); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 编辑
func (r *BaseRepository[T]) Update(model *T) error {
	if tx := mysql.DB.Save(model); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 删除
func (r *BaseRepository[T]) Delete(model *T) error {
	if tx := mysql.DB.Delete(model); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 获取单个字段
//
//	使用：
//	var item models.Demo
//	err := repo.GetBySingleField(1, "id", &item)
//	if err != nil {
//		// handle error
//	}
func (r *BaseRepository[T]) GetBySingleField(fieldValue interface{}, fieldName string, model *T) error {
	switch v := fieldValue.(type) {
	case int, int8, int16, int32, int64:
		if !tool.IsValidInt(v) {
			return errors.New(fieldName + " is required")
		}
	case string:
		if v == "" {
			return errors.New(fieldName + " is required")
		}
	default:
		return errors.New(fieldName + " is invalid ")
	}
	if tx := mysql.DB.Where("`"+fieldName+"` = ?", fieldValue).Find(model); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 获取 By Id
//
//	使用：
//	var user models.User
//	err := repo.GetById(1, &user)
//	if err != nil {
//		// handle error
//	}
func (r *BaseRepository[T]) GetById(id int64, model *T) error {
	if id <= 0 {
		return errors.New("id must be greater than 0")
	}
	if tx := mysql.DB.Where("id = ?", id).Find(model); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 获取全部
//
//	使用：
//	var items []models.Demo
//	err := repo.GetAll(&items)
//	if err != nil {
//		// handle error
//	}
func (r *BaseRepository[T]) GetAll(models *[]T) error {
	if tx := mysql.DB.Find(models); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 获取条件查询
//
//	使用：
//	var users []models.User
//	err := repo.GetByCondition(&users, "age > ?", 18)
//	if err != nil {
//		// handle error
//	}
func (r *BaseRepository[T]) GetByCondition(result *[]T, query string, args ...interface{}) error {
	db := mysql.DB
	// db = db.Debug()
	if query != "" {
		db = db.Where(query, args...)
	}
	return db.Find(result).Error
}

// 批量创建 - 事务
func (c *BaseRepository[T]) BatchCreate(items []T) error {
	// 开始
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// 批量更新 - 事务
func (c *BaseRepository[T]) BatchUpdate(items []T) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// 批量删除 - 事务
func (c *BaseRepository[T]) BatchDelete(items []T) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Delete(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
