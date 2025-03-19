package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	db, err := gorm.Open("mysql", "root:8888.216@/godb?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	//自动检查 Product 结构是否变化，变化则进行迁移
	db.AutoMigrate(&Product{})

	db.Create(&Product{Code: "123456", Price: 100})

	var product Product

	db.First(&product, 1)
	db.First(&product, "code = ?", "123456")

	fmt.Println(product)

	db.Model(&product).Update("Price", 200)

	fmt.Println(product)

	db.Delete(&product)

}
