package PrimaryKey02

// User GORM 默认使用 ID 作为主键名。
type User struct {
	ID string // 字段名 `ID` 将被作为默认的主键名
}

// Animal 设置字段 `AnimalID` 为默认主键
type Animal struct {
	AnimalID int64 `gorm:"primary_key"`
	Name     string
	Age      int64
}
