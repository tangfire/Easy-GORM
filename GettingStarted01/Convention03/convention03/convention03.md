# 复数表名

表名是结构体名称的复数形式

```go
type User struct {} // 默认的表名是 `users`

// 设置 `User` 的表名为 `profiles`
func (User) TableName() string {
	return "profiles"
}

func (u User) TableName() string {
	if u.Role == "admin" {
		return "admin_users"
	} else {
		return "users"
	}
}

// 如果设置禁用表名复数形式属性为 true，`User` 的表名将是 `user`
db.SingularTable(true)

```

# 指定表名

```go
// 用 `User` 结构体创建 `delete_users` 表
db.Table("deleted_users").CreateTable(&User{})

var deleted_users []User
db.Table("deleted_users").Find(&deleted_users)
//// SELECT * FROM deleted_users;

db.Table("deleted_users").Where("name = ?", "jinzhu").Delete()
//// DELETE FROM deleted_users WHERE name = 'jinzhu';
```

# 修改默认表名

你可以通过定义 `DefaultTableNameHandler` 字段来对表名使用任何规则。

```go
gorm.DefaultTableNameHandler = func (db *gorm.DB, defaultTableName string) string  {
    return "prefix_" + defaultTableName;
}
```

# 蛇形列名

列名是字段名的蛇形小写形式

```go
type User struct {
  ID        uint      // 字段名是 `id`
  Name      string    // 字段名是 `name`
  Birthday  time.Time // 字段名是 `birthday`
  CreatedAt time.Time // 字段名是 `created_at`
}

// 重写列名
type Animal struct {
    AnimalId    int64     `gorm:"column:beast_id"`         // 设置列名为 `beast_id`
    Birthday    time.Time `gorm:"column:day_of_the_beast"` // 设置列名为 `day_of_the_beast`
    Age         int64     `gorm:"column:age_of_the_beast"` // 设置列名为 `age_of_the_beast`
}
```

# 时间戳跟踪

## CreatedAt

对于有 `CreatedAt` 字段的模型，它将被设置为首次创建记录的当前时间。

```go
db.Create(&user) // 将设置 `CreatedAt` 为当前时间

// 你可以使用 `Update` 方法来更改默认时间
db.Model(&user).Update("CreatedAt", time.Now())
```

## UpdatedAt

对于有 `UpdatedAt` 字段的模型，它将被设置为记录更新时的当前时间。

```go
db.Save(&user) // 将设置 `UpdatedAt` 为当前时间

db.Model(&user).Update("name", "jinzhu") // 将设置 `UpdatedAt` 为当前时间
```

## DeletedAt

对于有 `DeletedAt` 字段的模型，当删除它们的实例时，它们并没有被从数据库中删除，只是将 `DeletedAt` 字段设置为当前时间。参考 Soft Delete