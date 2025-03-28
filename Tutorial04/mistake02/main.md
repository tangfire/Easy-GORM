
在 Go 语言中，错误处理是很重要的。

Go 语言中鼓励人们在任何 创建方法之后去检查错误。

# 错误处理

由于 GORM 的 链式 API，GORM 中的错误处理与惯用的 Go 代码不同，但它仍然相当容易。

如果发生任何错误，GORM 会将其设置为 `* gorm.DB` 的 `Error` 字段，你可以这样检查：

```go
if err := db.Where("name = ?", "jinzhu").First(&user).Error; err != nil {
    // error handling...
}
```

或者

```go
if result := db.Where("name = ?", "jinzhu").First(&user); result.Error != nil {
    // error handling...
}
```

# 错误

在处理数据期间，发生几个错误很普遍，GORM 提供了一个 API 来将所有发生的错误作为切片返回

```go
// 如果有多个错误产生，`GetErrors` 返回一个 `[]error`的切片
db.First(&user).Limit(10).Find(&users).GetErrors()

fmt.Println(len(errors))

for _, err := range errors {
  fmt.Println(err)
}
```

# RecordNotFound 错误


GORM 提供了一个处理 `RecordNotFound` 错误的快捷方式，如果发生了多个错误，它将检查每个错误，如果它们中的任何一个是`RecordNotFound` 错误。

```go
//检查是否返回 RecordNotFound 错误
db.Where("name = ?", "hello world").First(&user).RecordNotFound()

if db.Model(&user).Related(&credit_card).RecordNotFound() {
// 数据没有找到
}

if err := db.Where("name = ?", "jinzhu").First(&user).Error; gorm.IsRecordNotFoundError(err) {
// 数据没有找到
}
```



