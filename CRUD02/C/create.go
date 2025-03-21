package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type User struct {
	gorm.Model
	Name     string
	Age      int
	Birthday time.Time
}

func main() {
	db, err := gorm.Open("mysql", "root:8888.216@/godb?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{})

	user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}

	db.NewRecord(user) // => 返回 `true` ，因为主键为空

	db.Create(&user)

	db.NewRecord(user) // => 在 `user` 之后创建返回 `false`

	defer db.Close()
}
