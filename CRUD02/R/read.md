# 查询

```go
// 获取第一条记录，按主键排序
db.First(&user)
//// SELECT * FROM users ORDER BY id LIMIT 1;

// 获取一条记录，不指定排序
db.Take(&user)
//// SELECT * FROM users LIMIT 1;

// 获取最后一条记录，按主键排序
db.Last(&user)
//// SELECT * FROM users ORDER BY id DESC LIMIT 1;

// 获取所有的记录
db.Find(&users)
//// SELECT * FROM users;

// 通过主键进行查询 (仅适用于主键是数字类型)
db.First(&user, 10)
//// SELECT * FROM users WHERE id = 10;
```

# Where


# 原生 SQL

```go
// 获取第一条匹配的记录
db.Where("name = ?", "jinzhu").First(&user)
//// SELECT * FROM users WHERE name = 'jinzhu' limit 1;

// 获取所有匹配的记录
db.Where("name = ?", "jinzhu").Find(&users)
//// SELECT * FROM users WHERE name = 'jinzhu';

// <>
db.Where("name <> ?", "jinzhu").Find(&users)

// IN
db.Where("name in (?)", []string{"jinzhu", "jinzhu 2"}).Find(&users)

// LIKE
db.Where("name LIKE ?", "%jin%").Find(&users)

// AND
db.Where("name = ? AND age >= ?", "jinzhu", "22").Find(&users)

// Time
db.Where("updated_at > ?", lastWeek).Find(&users)

// BETWEEN
db.Where("created_at BETWEEN ? AND ?", lastWeek, today).Find(&users)
```

# Struct & Map

```go
// Struct
db.Where(&User{Name: "jinzhu", Age: 20}).First(&user)
//// SELECT * FROM users WHERE name = "jinzhu" AND age = 20 LIMIT 1;

// Map
db.Where(map[string]interface{}{"name": "jinzhu", "age": 20}).Find(&users)
//// SELECT * FROM users WHERE name = "jinzhu" AND age = 20;

// 多主键 slice 查询
db.Where([]int64{20, 21, 22}).Find(&users)
//// SELECT * FROM users WHERE id IN (20, 21, 22);
```

NOTE 当通过struct进行查询的时候，GORM 将会查询这些字段的非零值， 意味着你的字段包含 `0`， `''`， `false` 或者其他 零值, 将不会出现在查询语句中， 例如:

```go
db.Where(&User{Name: "jinzhu", Age: 0}).Find(&users)
//// SELECT * FROM users WHERE name = "jinzhu";
```

你可以考虑适用指针类型或者 scanner/valuer 来避免这种情况。

```go
// 使用指针类型
type User struct {
  gorm.Model
  Name string
  Age  *int
}

// 使用 scanner/valuer
type User struct {
  gorm.Model
  Name string
  Age  sql.NullInt64
}
```

# Not

和 `Where`查询类似

```go
db.Not("name", "jinzhu").First(&user)
//// SELECT * FROM users WHERE name <> "jinzhu" LIMIT 1;

// 不包含
db.Not("name", []string{"jinzhu", "jinzhu 2"}).Find(&users)
//// SELECT * FROM users WHERE name NOT IN ("jinzhu", "jinzhu 2");

//不在主键 slice 中
db.Not([]int64{1,2,3}).First(&user)
//// SELECT * FROM users WHERE id NOT IN (1,2,3);

db.Not([]int64{}).First(&user)
//// SELECT * FROM users;

// 原生 SQL
db.Not("name = ?", "jinzhu").First(&user)
//// SELECT * FROM users WHERE NOT(name = "jinzhu");

// Struct
db.Not(User{Name: "jinzhu"}).First(&user)
//// SELECT * FROM users WHERE name <> "jinzhu";
```

# Or

```go
db.Where("role = ?", "admin").Or("role = ?", "super_admin").Find(&users)
//// SELECT * FROM users WHERE role = 'admin' OR role = 'super_admin';

// Struct
db.Where("name = 'jinzhu'").Or(User{Name: "jinzhu 2"}).Find(&users)
//// SELECT * FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2';

// Map
db.Where("name = 'jinzhu'").Or(map[string]interface{}{"name": "jinzhu 2"}).Find(&users)
//// SELECT * FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2';
```

# 行内条件查询

和 `Where` 查询类似。

需要注意的是，当使用链式调用传入行内条件查询时，这些查询不会被传参给后续的中间方法。

```go
// 通过主键进行查询 (仅适用于主键是数字类型)
db.First(&user, 23)
//// SELECT * FROM users WHERE id = 23 LIMIT 1;
// 非数字类型的主键查询
db.First(&user, "id = ?", "string_primary_key")
//// SELECT * FROM users WHERE id = 'string_primary_key' LIMIT 1;

// 原生 SQL
db.Find(&user, "name = ?", "jinzhu")
//// SELECT * FROM users WHERE name = "jinzhu";

db.Find(&users, "name <> ? AND age > ?", "jinzhu", 20)
//// SELECT * FROM users WHERE name <> "jinzhu" AND age > 20;

// Struct
db.Find(&users, User{Age: 20})
//// SELECT * FROM users WHERE age = 20;

// Map
db.Find(&users, map[string]interface{}{"age": 20})
//// SELECT * FROM users WHERE age = 20;
```

# 额外的查询选项

```go
// 为查询 SQL 添加额外的选项
db.Set("gorm:query_option", "FOR UPDATE").First(&user, 10)
//// SELECT * FROM users WHERE id = 10 FOR UPDATE;

```
这段代码演示了在GORM中为查询操作添加数据库锁（`FOR UPDATE`）的实现方式，主要用于事务中锁定查询结果以防止并发修改。以下是对其原理和作用的详细解释：

---

### **1. 代码功能解析**
```go
db.Set("gorm:query_option", "FOR UPDATE").First(&user, 10)
// 生成SQL：SELECT * FROM users WHERE id = 10 FOR UPDATE;
```
- **核心机制**：
    - `Set("gorm:query_option", "FOR UPDATE")`：通过GORM的保留参数`gorm:query_option`，向生成的SQL语句末尾注入自定义片段`FOR UPDATE`。
    - `First(&user, 10)`：查询`id=10`的第一条记录，并将结果映射到`user`结构体。
    - **最终SQL**：`SELECT * FROM users WHERE id = 10 FOR UPDATE;`，表示在查询时对目标行加锁。

---

### **2. `FOR UPDATE`的作用**
这是数据库的**行级锁机制**，常见于事务场景：
- **并发控制**：锁定查询到的行，阻止其他事务对这些行的修改或加锁（如`UPDATE/DELETE`），直到当前事务提交或回滚。
- **数据一致性**：确保在事务中读取的数据在后续操作中不被其他事务干扰，适用于需要原子性操作的场景（如库存扣减、金额转账）。

---

### **3. 适用场景**
- **悲观锁实现**：在事务中先锁定数据，再进行更新操作。
  ```go
  tx := db.Begin()
  defer tx.Rollback()
  
  // 锁定用户记录
  tx.Set("gorm:query_option", "FOR UPDATE").First(&user, 10)
  
  // 更新用户余额（此时其他事务无法修改该行）
  user.Balance -= 100
  tx.Save(&user)
  
  tx.Commit()
  ```
- **数据库兼容性**：支持`FOR UPDATE`的数据库包括PostgreSQL、MySQL（需使用InnoDB引擎）等。

---

### **4. 注意事项**
1. **事务依赖**：`FOR UPDATE`需在事务中生效，否则锁会立即释放，失去意义。
2. **性能影响**：过度使用可能导致锁竞争和性能下降，需根据业务需求权衡。
3. **替代写法**：GORM推荐通过`Clauses`方法设置锁，更符合现代用法：
   ```go
   db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, 10)
   ```
   此方式生成的SQL与`FOR UPDATE`等价，但代码可读性更高。

---

### **5. 扩展对比**
| **方法**                | **优点**                  | **适用场景**             |
|-------------------------|---------------------------|--------------------------|
| `Set("gorm:query_option")` | 灵活，支持任意SQL片段      | 需快速注入自定义选项      |
| `Clauses(clause.Locking)` | 结构化，支持多种锁类型     | 规范化的锁操作（推荐写法） |

---

### **总结**
通过`Set("gorm:query_option", "FOR UPDATE")`，GORM允许开发者在查询时直接附加数据库锁指令，从而在高并发场景下保障数据一致性。但需注意事务的合理使用和性能优化，对于长期维护的代码，建议优先采用`Clauses`语法。


# FirstOrInit

获取第一条匹配的记录，或者通过给定的条件下初始一条新的记录（仅适用与于 struct 和 map 条件）。

```go
// 未查询到
db.FirstOrInit(&user, User{Name: "non_existing"})
//// user -> User{Name: "non_existing"}

// 查询到
db.Where(User{Name: "Jinzhu"}).FirstOrInit(&user)
//// user -> User{Id: 111, Name: "Jinzhu", Age: 20}
db.FirstOrInit(&user, map[string]interface{}{"name": "jinzhu"})
//// user -> User{Id: 111, Name: "Jinzhu", Age: 20}
```


# Attrs

