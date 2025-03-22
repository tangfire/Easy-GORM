# 更新所有字段

`Save` 方法在执行 SQL 更新操作时将包含所有字段，即使这些字段没有被修改。

```go
db.First(&user)

user.Name = "jinzhu 2"
user.Age = 100
db.Save(&user)

//// UPDATE users SET name='jinzhu 2', age=100, birthday='2016-01-01', updated_at = '2013-11-17 21:34:10' WHERE id=111;
```

# 更新已更改的字段

如果你只想更新已经修改了的字段，可以使用 `Update`，`Updates` 方法。