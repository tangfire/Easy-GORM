package main

import "github.com/jinzhu/gorm"

// gorm.Model 定义
//type Model struct {
//	ID        uint `gorm:"primary_key"`
//	CreatedAt time.Time
//	UpdatedAt time.Time
//	DeletedAt *time.Time
//}

type User struct {
	gorm.Model
	Name string
}
