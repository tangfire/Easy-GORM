package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	gorm.Model
	CreditCard CreditCard // has one 关系
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID uint // 外键字段（自动关联 User.ID）
}

func main() {
	db, err := gorm.Open("mysql", "root:8888.216@/godb?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	//自动检查 Product 结构是否变化，变化则进行迁移
	db.AutoMigrate(&User{})
	db.AutoMigrate(&CreditCard{})

}
