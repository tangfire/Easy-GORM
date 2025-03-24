# 运行原生SQL

执行原生 SQL时不能通过链式调用其他方法

```go
db.Exec("DROP TABLE users;")
db.Exec("UPDATE orders SET shipped_at=? WHERE id IN (?)", time.Now(), []int64{11,22,33})

// Scan
type Result struct {
    Name string
    Age  int
}

var result Result
db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
```


这段代码包含三个独立的数据库操作，分别涉及 **表删除**、**数据更新** 和 **原生 SQL 查询**。以下是逐段解析：

---

### 一、删除表 `users`
```go
db.Exec("DROP TABLE users;")
```
- **作用**：直接执行原生 SQL，删除数据库中的 `users` 表及其所有数据。
- **关键点**：
    - `Exec` 方法用于执行不返回数据的 SQL 语句（如 `DROP`, `UPDATE`, `INSERT`）。
    - **高危操作**：`DROP TABLE` 会永久删除表结构和数据，需谨慎使用（尤其在生产环境）。
    - **依赖关系**：若其他表通过外键关联到 `users`，可能因约束导致删除失败，需先处理依赖。

---

### 二、批量更新 `orders` 表
```go
db.Exec(
    "UPDATE orders SET shipped_at=? WHERE id IN (?)", 
    time.Now(), 
    []int64{11,22,33},
)
```
- **作用**：将 `orders` 表中 `id` 为 11、22、33 的记录的 `shipped_at` 字段更新为当前时间。
- **关键点**：
    - **参数化查询**：`?` 是占位符，GORM 会自动转义参数，防止 SQL 注入。
    - **切片展开**：`[]int64{11,22,33}` 会被转换为 `(11,22,33)`，适应 `IN` 子句的语法。
    - **事务性**：此操作是独立执行的，若需保证多个操作的原子性，需显式启用事务（此处未使用）。

---

### 三、原生 SQL 查询与结果映射
```go
type Result struct {
    Name string
    Age  int
}

var result Result
db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
```
- **作用**：执行原生 SQL 查询，将 `users` 表中 `name` 为 3 的记录的 `name` 和 `age` 字段映射到 `Result` 结构体。
- **关键点**：
    - **`Raw` + `Scan`**：`Raw` 用于编写原生 SQL，`Scan` 将查询结果映射到结构体或变量。
    - **字段匹配**：结构体 `Result` 的字段名（`Name`, `Age`）需与查询结果的列名（`name`, `age`）匹配（不区分大小写）。
    - **单条记录**：若查询返回多条记录，需使用 `Scan` 到切片（如 `var results []Result`）或结合 `Rows` 逐行处理。

---

### 四、注意事项与风险
1. **直接执行 DDL 的风险**  
   `DROP TABLE` 应避免在应用代码中硬编码，通常通过数据库迁移工具（如 Goose、GORM Migrate）管理表结构变更。
2. **原生 SQL 的维护性**  
   虽然灵活，但原生 SQL 可能降低代码可读性，且与数据库类型耦合（如 MySQL 和 PostgreSQL 语法差异）。优先使用 GORM 的链式方法（如 `Where`, `Updates`）。
3. **错误处理缺失**  
   代码中未检查 `Exec` 或 `Raw` 的返回错误，实际使用时需添加错误处理逻辑：
   ```go
   if err := db.Exec(...).Error; err != nil {
       // 处理错误
   }
   ```

---

### 五、替代方案示例
1. **使用 GORM 方法更新数据**（避免原生 SQL）：
   ```go
   db.Model(&Order{}).
     Where("id IN ?", []int64{11,22,33}).
     Update("shipped_at", time.Now())
   ```
2. **查询单条记录**：
   ```go
   var user User
   db.Select("name", "age").Where("name = ?", 3).First(&user)
   ```

---

### 总结
这段代码演示了直接操作数据库的三种典型场景，但缺乏错误处理和事务管理。在实际项目中，建议：
- 对写操作（如 `DROP`, `UPDATE`）添加事务包裹。
- 优先使用 ORM 方法而非原生 SQL，除非有明确的性能或复杂查询需求。
- 始终检查并处理数据库操作的错误。


# `sql.Row` 和 `sql.Rows`

使用 `*sql.Row` 或者 `*sql.Rows` 获得查询结果

```go
row := db.Table("users").Where("name = ?", "jinzhu").Select("name, age").Row() // (*sql.Row)
row.Scan(&name, &age)

rows, err := db.Model(&User{}).Where("name = ?", "jinzhu").Select("name, age, email").Rows() // (*sql.Rows, error)
defer rows.Close()
for rows.Next() {
    ...
    rows.Scan(&name, &age, &email)
    ...
}

// 原生SQL
rows, err := db.Raw("select name, age, email from users where name = ?", "jinzhu").Rows() // (*sql.Rows, error)
defer rows.Close()
for rows.Next() {
    ...
    rows.Scan(&name, &age, &email)
    ...
}
```

