
```go
user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}

db.NewRecord(user) // => 返回 `true` ，因为主键为空

db.Create(&user)

db.NewRecord(user) // => 在 `user` 之后创建返回 `false`
```


这段代码演示了使用GORM创建数据库记录的过程，并通过`NewRecord()`方法判断主键状态。具体解释如下：

---

### 1. **初始化结构体**
```go
user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
```
- 创建了一个`User`结构体实例，但未显式赋值主键字段（假设模型中的主键是`ID`）。

---

### 2. **首次调用 `NewRecord()`**
```go
db.NewRecord(user) // 返回 true
```
- `NewRecord()`的作用是**检查主键是否为空**。由于此时`user.ID`未赋值（默认零值，如`0`），因此返回`true`。
- GORM通过模型定义的主键字段（如`ID`）判断记录是否已存在数据库中。

---

### 3. **调用 `Create()` 插入记录**
```go
db.Create(&user)
```
- `Create()`会将`user`插入数据库，并**自动回填主键值**（例如自增ID）到`user.ID`字段。
- **必须传递指针**（`&user`），否则GORM无法修改原结构体的`ID`字段。

---

### 4. **再次调用 `NewRecord()`**
```go
db.NewRecord(user) // 返回 false
```
- 插入后，`user.ID`已被数据库赋予非零值（如自增生成的整数值），因此`NewRecord()`返回`false`。
- 这表示该记录已存在于数据库中，后续操作（如更新）需依赖此主键。

---

### 关键点总结
| 操作              | 逻辑说明                                                                 |
|-------------------|------------------------------------------------------------------------|
| `NewRecord()`     | 通过主键零值判断是否为“新记录”，常用于区分创建/更新操作。            |
| `Create(&user)`   | 插入记录并回填主键，需注意传递指针以保证数据同步。                       |
| 主键回填机制       | 依赖数据库的自增主键或默认值实现，回填后`ID`字段不再为空。         |

---

### 补充说明
- 若模型未定义主键，GORM会默认将`ID`字段作为主键。
- 如果手动赋值主键（如`user.ID = 100`），则首次`NewRecord()`会直接返回`false`。



# 默认值

你可以通过标签定义字段的默认值，例如：

```go
type Animal struct {
    ID   int64
    Name string `gorm:"default:'galeone'"`
    Age  int64
}
```

然后 SQL 会排除那些没有值或者有 零值 的字段，在记录插入数据库之后，gorm将从数据库中加载这些字段的值。


```go
var animal = Animal{Age: 99, Name: ""}
db.Create(&animal)
// INSERT INTO animals("age") values('99');
// SELECT name from animals WHERE ID=111; // 返回的主键是 111
// animal.Name => 'galeone'
```

注意 所有包含零值的字段，像 `0`，`''`，`false` 或者其他的 零值 不会被保存到数据库中，但会使用这个字段的默认值。你应该考虑使用指针类型或者其他的值来避免这种情况:

```go
// Use pointer value
type User struct {
  gorm.Model
  Name string
  Age  *int `gorm:"default:18"`
}

// Use scanner/valuer
type User struct {
  gorm.Model
  Name string
  Age  sql.NullInt64 `gorm:"default:18"`
}
```


# 在钩子中设置字段值

如果你想在 `BeforeCreate` 函数中更新字段的值，应该使用 `scope.SetColumn`，例如：

```go
func (user *User) BeforeCreate(scope *gorm.Scope) error {
  scope.SetColumn("ID", uuid.New())
  return nil
}
```




# 创建额外选项

```go
// 为插入 SQL 语句添加额外选项
db.Set("gorm:insert_option", "ON CONFLICT").Create(&product)
// INSERT INTO products (name, code) VALUES ("name", "code") ON CONFLICT;
```


这段代码是GORM框架中用于实现**自定义插入选项**的典型用法，主要用于在插入数据库时添加特定数据库的扩展语法（例如PostgreSQL的`ON CONFLICT`冲突处理）。以下是分步解析：

---

### 1. **代码功能解释**
```go
db.Set("gorm:insert_option", "ON CONFLICT").Create(&product)
```
- **作用**：通过`Set("gorm:insert_option", "ON CONFLICT")`在生成的INSERT语句末尾附加自定义SQL片段`ON CONFLICT`，最终生成的SQL语句为：
  ```sql
  INSERT INTO products (name, code) VALUES ("name", "code") ON CONFLICT;
  ```
- **核心机制**：
    - `gorm:insert_option`是GORM的保留关键字参数，允许开发者在执行`Create()`时向INSERT语句注入自定义内容。
    - `ON CONFLICT`是PostgreSQL特有的语法，用于处理插入时的唯一性约束冲突（例如：冲突时执行更新或忽略操作）。

---

### 2. **适用场景**
- **合并插入（Upsert）**：当需要实现“存在则更新，不存在则插入”的逻辑时，PostgreSQL可通过`ON CONFLICT ... DO UPDATE`实现，而MySQL则使用`ON DUPLICATE KEY UPDATE`。
- **扩展数据库功能**：某些数据库支持在INSERT语句后附加特定指令（如SQLite的`RETURNING`子句返回插入数据），此时可通过此方法注入。

---

### 3. **注意事项**
- **数据库兼容性**：不同数据库的冲突处理语法不同，需根据实际数据库类型调整`insert_option`的值（例如PostgreSQL用`ON CONFLICT`，MySQL用`ON DUPLICATE KEY`）。
- **链式调用顺序**：`Set()`需在`Create()`之前调用才能生效，且作用于当前链式操作的作用域。
- **参数传递**：若需动态传递变量（如冲突时的更新字段），需将完整的SQL片段写入`insert_option`中，例如：
  ```go
  db.Set("gorm:insert_option", "ON CONFLICT (name) DO UPDATE SET age = EXCLUDED.age")
  ```

---

### 4. **实际应用示例（PostgreSQL）**
```go
// 冲突时更新特定字段
db.Set("gorm:insert_option", "ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name").
   Create(&user)
```
生成的SQL：
```sql
INSERT INTO users (email, name) VALUES ('test@example.com', 'John') 
ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name;
```
- `EXCLUDED`是PostgreSQL的关键字，指代冲突时尝试插入的数据。

---

### 总结
通过`Set("gorm:insert_option", ...)`，GORM提供了灵活的方式扩展原生SQL功能，尤其适用于需要依赖数据库特性的高级插入操作。但需注意不同数据库的语法差异，并确保注入的SQL片段符合当前数据库的规范。