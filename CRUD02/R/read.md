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

如果未找到记录，则使用参数初始化 struct

```go
// 未查询到
db.Where(User{Name: "non_existing"}).Attrs(User{Age: 20}).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = 'non_existing';
//// user -> User{Name: "non_existing", Age: 20}

db.Where(User{Name: "non_existing"}).Attrs("age", 20).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = 'non_existing';
//// user -> User{Name: "non_existing", Age: 20}

// 查询到
db.Where(User{Name: "Jinzhu"}).Attrs(User{Age: 30}).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = jinzhu';
//// user -> User{Id: 111, Name: "Jinzhu", Age: 20}
```

# Assign

无论是否查询到数据，都将参数赋值给 struct

```go
// 未查询到
db.Where(User{Name: "non_existing"}).Assign(User{Age: 20}).FirstOrInit(&user)
//// user -> User{Name: "non_existing", Age: 20}

// 查询到
db.Where(User{Name: "Jinzhu"}).Assign(User{Age: 30}).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = jinzhu';
//// user -> User{Id: 111, Name: "Jinzhu", Age: 30}
```

# FirstOrCreate

获取第一条匹配的记录，或者通过给定的条件创建一条记录 （仅适用与于 struct 和 map 条件）。

```go
// 未查询到
db.FirstOrCreate(&user, User{Name: "non_existing"})
//// INSERT INTO "users" (name) VALUES ("non_existing");
//// user -> User{Id: 112, Name: "non_existing"}

// 查询到
db.Where(User{Name: "Jinzhu"}).FirstOrCreate(&user)
//// user -> User{Id: 111, Name: "Jinzhu"}
```


# Attr

如果未查询到记录，通过给定的参数赋值给 struct ，然后使用这些值添加一条记录。


```go
// 未查询到
db.Where(User{Name: "non_existing"}).Attrs(User{Age: 20}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'non_existing';
//// INSERT INTO "users" (name, age) VALUES ("non_existing", 20);
//// user -> User{Id: 112, Name: "non_existing", Age: 20}

// 查询到
db.Where(User{Name: "jinzhu"}).Attrs(User{Age: 30}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'jinzhu';
//// user -> User{Id: 111, Name: "jinzhu", Age: 20}
```

# Assign

无论是否查询到，都将其分配给记录，并保存到数据库中。


```go
// 未查询到
db.Where(User{Name: "non_existing"}).Assign(User{Age: 20}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'non_existing';
//// INSERT INTO "users" (name, age) VALUES ("non_existing", 20);
//// user -> User{Id: 112, Name: "non_existing", Age: 20}

// 查询到
db.Where(User{Name: "jinzhu"}).Assign(User{Age: 30}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'jinzhu';
//// UPDATE users SET age=30 WHERE id = 111;
//// user -> User{Id: 111, Name: "jinzhu", Age: 30}
```


# FirstOrInit和FirstOrCreate的区别


在 GORM 框架中，`FirstOrInit` 和 `FirstOrCreate` 都用于根据条件查询记录并根据结果执行不同操作，但两者的核心区别在于 **是否自动持久化数据到数据库**，以及 **对后续操作的支持**。以下是具体对比：

---

### 1. **`FirstOrCreate`：查询或创建**
- **行为逻辑**：  
  根据条件查询数据库，若存在匹配记录则返回第一条；若不存在，则直接**创建新记录并保存到数据库**。
- **适用场景**：  
  需要确保记录一定存在的场景（例如初始化用户配置、快速填充默认数据）。
- **额外参数**：  
  支持通过 `Attrs()` 或 `Assign()` 指定创建时的默认值或强制覆盖值。
  ```go
  // 示例：若不存在则创建，并设置 Attrs 中的默认值
  db.Where(User{Name: "Alice"}).Attrs(User{Age: 30}).FirstOrCreate(&user)
  ```

---

### 2. **`FirstOrInit`：查询或初始化**
- **行为逻辑**：  
  根据条件查询数据库，若存在匹配记录则返回第一条；若不存在，则**初始化一个未保存的结构体对象**（仅在内存中，不写入数据库）。
- **适用场景**：  
  需要根据条件动态构建对象，但暂不确定是否需要保存的场景（例如表单预填充、条件校验）。
- **额外参数**：  
  同样支持 `Attrs()` 或 `Assign()`，但这些值仅影响初始化对象的属性，需手动调用 `Save()` 才会持久化。
  ```go
  // 示例：若不存在则初始化对象，后续可修改并手动保存
  db.Where(User{Name: "Bob"}).Attrs(User{Age: 25}).FirstOrInit(&user)
  user.Email = "bob@example.com"
  db.Save(&user) // 显式保存
  ```

---

### 3. **关键区别总结**
| **特性**               | `FirstOrCreate`                  | `FirstOrInit`                  |
|------------------------|----------------------------------|--------------------------------|
| **数据库写入**          | 自动创建记录                     | 仅初始化对象，需手动调用 `Save` |
| **返回值状态**          | 已存在的记录或新创建的对象       | 已存在的记录或未保存的新对象   |
| **链式操作支持**        | 可直接结合 `Create` 相关逻辑      | 需后续操作（如 `Update`/`Save`）|
| **性能影响**            | 可能触发写操作，需注意并发控制   | 无写操作，仅内存操作           |

---

### 4. **如何选择？**
- **优先 `FirstOrCreate`**：  
  当逻辑明确需要“存在即返回，不存在即创建”时（如确保唯一配置），直接使用此方法简化代码。
- **优先 `FirstOrInit`**：  
  当需要根据查询结果动态修改对象后再决定是否保存时（例如表单数据填充后校验再提交），使用此方法更灵活。

通过合理选择这两个方法，可以避免重复代码并减少潜在的并发问题。


# 高级查询

## 子查询

使用 `*gorm.expr` 进行子查询

```go
db.Where("amount > ?", DB.Table("orders").Select("AVG(amount)").Where("state = ?", "paid").QueryExpr()).Find(&orders)
// SELECT * FROM "orders"  WHERE "orders"."deleted_at" IS NULL AND (amount > (SELECT AVG(amount) FROM "orders"  WHERE (state = 'paid')));
```



这段代码使用了 GORM 的 **子查询嵌套** 功能，目的是筛选出金额（`amount`）高于“已支付订单平均金额”的所有订单。其核心逻辑是通过 **子查询动态计算平均值**，并将其作为外层查询的过滤条件。以下是代码的逐层解析：

---

### **1. 代码结构分解**
```go
db.Where(
    "amount > ?", 
    DB.Table("orders").Select("AVG(amount)").Where("state = ?", "paid").QueryExpr()
).Find(&orders)
```

#### **关键方法解析**
- **`DB.Table("orders")`**  
  明确操作的表为 `orders`。

- **`.Select("AVG(amount)")`**  
  选择计算 `amount` 字段的平均值（`AVG` 是 SQL 聚合函数）。

- **`.Where("state = ?", "paid")`**  
  添加条件：仅筛选 `state` 为 `paid`（已支付）的订单。

- **`.QueryExpr()`**  
  将上述查询转换为一个 **子查询表达式**，以便嵌入到外层查询的条件中。

- **外层 `Where("amount > ?")`**  
  外层查询通过 `amount > ?` 条件，将子查询的结果作为动态参数传入。

---

### **2. 生成的 SQL 语句**
最终生成的 SQL 语句为：
```sql
SELECT * FROM "orders"  
WHERE "orders"."deleted_at" IS NULL 
AND (amount > (
    SELECT AVG(amount) FROM "orders" 
    WHERE state = 'paid'
));
```

#### **注意点**
- **软删除处理**  
  如果模型启用了 GORM 的软删除（`deleted_at` 字段），查询会自动添加 `deleted_at IS NULL` 条件。

- **子查询动态性**  
  子查询 `SELECT AVG(amount) FROM orders WHERE state = 'paid'` 的结果会实时计算，确保外层条件始终基于最新数据。

---

### **3. 实际应用场景**
此代码适用于以下业务场景：
- **筛选高价值订单**  
  例如：找出所有金额高于平均值的优质订单，用于后续分析或营销。

- **动态阈值计算**  
  避免硬编码金额阈值，通过子查询实现动态条件，适应数据变化。

---

### **4. 潜在优化点**
- **索引优化**  
  若 `state` 和 `amount` 字段频繁用于查询，可为这两个字段添加联合索引以加速子查询。

- **子查询性能**  
  当 `orders` 表数据量极大时，子查询可能成为性能瓶颈。可考虑缓存平均值或改用分库分表策略。

- **条件扩展性**  
  若需更复杂的逻辑（如多状态平均），可调整子查询条件，例如：
  ```go
  .Where("state IN (?)", []string{"paid", "completed"})
  ```

---

### **5. 常见错误排查**
- **子查询语法错误**  
  确保 `QueryExpr()` 正确生成子查询表达式，避免手动拼接 SQL 字符串导致的错误。

- **字段名大小写敏感**  
  数据库字段名需与模型定义严格一致（如 `amount` 是否应为 `Amount`）。

- **空值处理**  
  当无 `paid` 订单时，子查询返回 `NULL`，可能导致外层条件失效。可通过 `COALESCE` 函数设置默认值：
  ```go
  .Select("COALESCE(AVG(amount), 0)")
  ```

---

通过这种方式，GORM 将复杂的嵌套查询简化为链式调用，既保持了代码可读性，又充分利用了 SQL 的动态计算能力。



# 查询

指定要从数据库检索的字段，默认情况下，将选择所有字段。

```go
db.Select("name, age").Find(&users)
//// SELECT name, age FROM users;

db.Select([]string{"name", "age"}).Find(&users)
//// SELECT name, age FROM users;

db.Table("users").Select("COALESCE(age,?)", 42).Rows()
//// SELECT COALESCE(age,'42') FROM users;
```

这段代码使用 GORM 的 **原生 SQL 函数** 和 **参数绑定** 来处理可能为 `NULL` 的字段，并为其设置默认值。它的核心目的是查询 `users` 表，并对 `age` 字段进行空值替换：**如果 `age` 为 `NULL`，则返回默认值 `42`**。以下是详细解析：

---

### **1. 代码结构分解**
```go
db.Table("users").              // 指定操作的表为 `users`
  Select("COALESCE(age, ?)", 42). // 使用 COALESCE 函数处理 `age` 字段
  Rows()                        // 执行查询并返回数据库游标（*sql.Rows）
```

#### **关键方法解析**
- **`Table("users")`**  
  明确操作的表名，避免依赖模型自动推断表名。

- **`Select("COALESCE(age, ?)", 42)`**
  - **`COALESCE`** 是 SQL 标准函数，返回第一个非 `NULL` 的参数。
  - `?` 是占位符，GORM 会自动将 `42` 作为参数绑定到查询中。
  - 逻辑等价于：如果 `age` 存在则取 `age`，否则返回 `42`。

- **`Rows()`**  
  执行查询并返回原始数据库游标，允许通过 `Next()` 和 `Scan()` 手动遍历结果（需自行处理关闭和错误）。

---

### **2. 生成的 SQL 语句**
最终生成的 SQL 语句为：
```sql
SELECT COALESCE(age, '42') FROM users;
```

#### **注意点**
- **参数类型问题**  
  代码中传入的 `42` 是整数，但生成的 SQL 中却被转换为字符串 `'42'`。  
  **原因**：GORM 默认将非结构体/非切片参数按字符串处理。  
  **解决方案**：若需保持整数类型，需明确指定参数类型（例如使用 `sql.NullInt64` 或原生 SQL）。

- **空表处理**  
  如果 `users` 表为空，查询会返回空结果，但不会报错。

---

### **3. 实际应用场景**
- **数据清洗**  
  查询时统一缺失字段的默认值（例如将 `NULL` 年龄替换为 `42`）。

- **报表统计**  
  避免聚合函数（如 `AVG`、`SUM`）因 `NULL` 值导致的计算偏差。

- **兼容性处理**  
  对可能未设置年龄的新用户提供兼容默认值。

---

### **4. 潜在优化与问题**
#### **参数类型错误**
- **问题**：生成的 SQL 中 `42` 被转换为字符串，如果 `age` 字段是整数类型，可能导致隐式类型转换或查询错误。
- **修复方法**：  
  使用 GORM 的 `Expr` 直接传递原生参数（避免自动转义）：
  ```go
  db.Table("users").
    Select("COALESCE(age, ?)", gorm.Expr("42")). // 保持整数类型
    Rows()
  ```
  生成的 SQL：
  ```sql
  SELECT COALESCE(age, 42) FROM users;
  ```

#### **性能优化**
- **索引覆盖**  
  如果仅查询 `age` 字段且表数据量大，可为 `age` 添加索引：
  ```sql
  CREATE INDEX idx_users_age ON users(age);
  ```

#### **结果处理优化**
- 使用 `Scan` 直接映射到结构体或变量（更简洁）：
  ```go
  var ages []int
  db.Table("users").
    Select("COALESCE(age, 42) as age").
    Scan(&ages) // 自动绑定到切片
  ```

---

### **5. 扩展知识：COALESCE 与 NULL 处理**
- **`COALESCE` vs `IFNULL`**
  - `COALESCE` 是标准 SQL 函数，支持多个参数（返回第一个非 `NULL` 值）。
  - `IFNULL` 是 MySQL 特有函数，仅支持两个参数。

- **与 `ISNULL` 的区别**  
  `ISNULL` 仅判断是否为 `NULL`（返回布尔值），而 `COALESCE` 直接返回替代值。

---

### **6. 完整示例代码**
```go
rows, err := db.Table("users").
  Select("COALESCE(age, ?)", 42).
  Rows()
if err != nil {
  log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
  var age int
  if err := rows.Scan(&age); err != nil {
    log.Fatal(err)
  }
  fmt.Println("User age:", age)
}
```

---

通过这段代码，你可以安全地处理数据库中可能存在的 `NULL` 值，同时避免应用程序因空值导致的逻辑异常。如果需要精确控制参数类型，建议结合 `gorm.Expr` 或原生 SQL 实现。



# Order

使用 Order 从数据库查询记录时，当第二个参数设置为 true 时，将会覆盖之前的定义条件。

```go
db.Order("age desc, name").Find(&users)
//// SELECT * FROM users ORDER BY age desc, name;

// 多个排序条件
db.Order("age desc").Order("name").Find(&users)
//// SELECT * FROM users ORDER BY age desc, name;

// 重新排序
db.Order("age desc").Find(&users1).Order("age", true).Find(&users2)
//// SELECT * FROM users ORDER BY age desc; (users1)
//// SELECT * FROM users ORDER BY age; (users2)
```




# Limit

指定要查询的最大记录数

```go
db.Limit(3).Find(&users)
//// SELECT * FROM users LIMIT 3;

// 用 -1 取消 LIMIT 限制条件
db.Limit(10).Find(&users1).Limit(-1).Find(&users2)
//// SELECT * FROM users LIMIT 10; (users1)
//// SELECT * FROM users; (users2)
```


这段代码通过 GORM 的链式调用实现了 **分批次查询**，但利用了 `Limit(-1)` 的特殊语法来 **取消前序的 `LIMIT` 限制**，最终生成两个不同的 SQL 查询。以下是逐层解析：

---

### **1. 代码结构与执行逻辑**
```go
db.Limit(10).Find(&users1).Limit(-1).Find(&users2)
```
- **第一段查询**：`db.Limit(10).Find(&users1)`  
  生成 SQL：`SELECT * FROM users LIMIT 10`，将前 10 条记录存入 `users1`。

- **第二段查询**：`.Limit(-1).Find(&users2)`  
  生成 SQL：`SELECT * FROM users`，取消 `LIMIT 10` 的限制，查询全部记录并存入 `users2`。

---

### **2. 关键机制解析**
#### **(1) GORM 的链式调用与条件继承**
- GORM 的链式调用会 **继承并合并条件**，例如：
  ```go
  db.Where("age > 20").Limit(10).Find(&users1).Find(&users2)
  ```
  这里 `users2` 的查询会同时继承 `WHERE age > 20` 和 `LIMIT 10` 条件。

- **问题**：直接链式调用会导致第二个查询也包含 `LIMIT 10`，与预期不符。

#### **(2) `Limit(-1)` 的作用**
- `Limit(-1)` 是 GORM 提供的 **特殊语法**，用于清除前序的 `LIMIT` 限制。
- 在上述代码中，通过 `.Limit(-1)` 显式重置了条件，使第二个查询不再受 `LIMIT 10` 影响。

---

### **3. 生成的 SQL 语句**
```sql
-- users1 的查询（带限制）
SELECT * FROM users LIMIT 10;

-- users2 的查询（无限制）
SELECT * FROM users;
```

---

### **4. 应用场景**
- **分批次处理数据**：  
  例如先获取部分数据快速展示（`users1`），再加载全部数据用于后台处理（`users2`）。
- **条件动态调整**：  
  在复杂链式调用中灵活切换条件，避免继承冗余限制。

---

### **5. 潜在风险与优化建议**
#### **(1) 风险**
- **条件污染**：  
  若未正确使用 `Limit(-1)`，后续查询可能意外继承前序条件（如 `WHERE`、`ORDER BY`）。

- **性能问题**：  
  若 `users` 表数据量极大，无限制查询可能导致内存溢出或响应延迟。

#### **(2) 优化方案**
- **显式重置会话**：  
  使用 `Session(&gorm.Session{})` 创建独立会话，避免条件继承：
  ```go
  db.Limit(10).Find(&users1)
  db.Session(&gorm.Session{}).Find(&users2) // 完全独立的条件
  ```
- **分页控制**：  
  对大数据集使用分页（`Limit` + `Offset`）而非无限制查询：
  ```go
  db.Limit(10).Offset(0).Find(&users1)  // 第一页
  db.Limit(10).Offset(10).Find(&users2) // 第二页
  ```

---

### **6. 总结**
| **操作**                | **作用**                          | **适用场景**                     |
|-------------------------|-----------------------------------|----------------------------------|
| `Limit(10)`             | 添加 `LIMIT 10` 条件              | 限制查询结果数量                 |
| `Limit(-1)`             | 清除前序 `LIMIT` 条件              | 动态调整查询限制                 |
| `Session()`             | 创建独立会话，重置所有条件         | 复杂查询链中避免条件污染         |

通过合理使用 `Limit(-1)` 或 `Session()`，可以在链式调用中灵活控制查询逻辑，同时规避潜在风险。



# Offset

指定在开始返回记录之前要跳过的记录数。

```go
db.Offset(3).Find(&users)
//// SELECT * FROM users OFFSET 3;

// 用 -1 取消 OFFSET 限制条件
db.Offset(10).Find(&users1).Offset(-1).Find(&users2)
//// SELECT * FROM users OFFSET 10; (users1)
//// SELECT * FROM users; (users2)
```

这段代码利用 GORM 的链式操作特性，通过 `Offset(-1)` 实现了 **分页查询中偏移条件的动态取消**。其核心目的是在同一链式调用中，先查询部分数据（跳过前 10 条），再取消偏移限制以查询全部数据。以下是具体解析：

---

### **1. 代码逻辑分解**
```go
db.Offset(10).Find(&users1).Offset(-1).Find(&users2)
```
- **第一段查询** `db.Offset(10).Find(&users1)`  
  生成 SQL：`SELECT * FROM users OFFSET 10;`  
  作用：跳过前 10 条记录，将剩余数据存入 `users1`（例如用于分页第二页）。

- **第二段查询** `.Offset(-1).Find(&users2)`  
  生成 SQL：`SELECT * FROM users;`  
  作用：`Offset(-1)` 清除前序的 `OFFSET 10` 条件，查询所有记录存入 `users2`。

---

### **2. 关键机制**
#### **(1) GORM 的条件继承与重置**
- GORM 的链式调用会 **继承前序条件**（如 `WHERE`、`ORDER BY`、`LIMIT/OFFSET`）。
- `Offset(-1)` 是 GORM 的 **特殊语法**，用于清除当前链式调用中累积的 `OFFSET` 条件。

#### **(2) 应用场景**
- **分页与全量混合查询**：例如先展示分页数据（`users1`），再后台处理全量数据（`users2`）。
- **动态调整查询范围**：在复杂业务逻辑中灵活切换偏移量，避免重复定义查询条件。

---

### **3. 潜在问题与优化**
#### **(1) 条件污染风险**
- 若链式调用中存在其他条件（如 `WHERE`），第二个查询会 **继承这些条件**。例如：
  ```go
  db.Where("age > 20").Offset(10).Find(&users1).Offset(-1).Find(&users2)
  ```
  第二个查询的 SQL 会是 `SELECT * FROM users WHERE age > 20;`（仍保留 `WHERE` 条件）。

- **解决方案**：使用 `Session(&gorm.Session{})` 创建独立会话以重置所有条件：
  ```go
  db.Offset(10).Find(&users1)
  db.Session(&gorm.Session{}).Find(&users2) // 完全独立的条件
  ```

#### **(2) 性能优化建议**
- **分页控制**：对大数据集使用 `LIMIT` + `OFFSET` 组合，避免全量查询：
  ```go
  db.Limit(10).Offset(10).Find(&users1) // 分页第二页
  ```
- **索引优化**：为 `ORDER BY` 和 `WHERE` 涉及的字段添加索引，加速分页查询。

---

### **4. 对比其他分页方法**
| **方法**             | 优点                          | 缺点                     |
|----------------------|-------------------------------|--------------------------|
| `Offset(-1)` 重置    | 灵活取消偏移量，无需重复定义条件 | 需注意条件继承问题       |
| `Session()` 独立会话  | 彻底隔离查询条件               | 代码冗余，需显式调用     |
| 原生 SQL 分页        | 直接控制 SQL 语句              | 失去 ORM 的链式操作便利性 |

---

### **5. 总结**
- **`Offset(-1)` 的作用**：清除前序 `OFFSET` 条件，适用于动态调整分页范围的场景。
- **注意事项**：需警惕链式调用中其他条件的继承问题，必要时使用 `Session()` 隔离查询。
- **性能优先**：结合 `LIMIT` 和索引优化，避免全表扫描对数据库造成压力。

通过合理运用 GORM 的条件重置机制，可以在保证代码简洁性的同时，灵活应对复杂的分页需求。


# Count

获取模型记录数

```go
db.Where("name = ?", "jinzhu").Or("name = ?", "jinzhu 2").Find(&users).Count(&count)
//// SELECT * from USERS WHERE name = 'jinzhu' OR name = 'jinzhu 2'; (users)
//// SELECT count(*) FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2'; (count)

db.Model(&User{}).Where("name = ?", "jinzhu").Count(&count)
//// SELECT count(*) FROM users WHERE name = 'jinzhu'; (count)

db.Table("deleted_users").Count(&count)
//// SELECT count(*) FROM deleted_users;
```

注意： 在查询链中使用 `Count` 时，必须放在最后一个位置，因为它会覆盖 `SELECT` 查询条件。


# Group和Having

```go
rows, err := db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Rows()
for rows.Next() {
    ...
}

rows, err := db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Having("sum(amount) > ?", 100).Rows()
for rows.Next() {
    ...
}

type Result struct {
    Date  time.Time
    Total int64
}
db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Having("sum(amount) > ?", 100).Scan(&results)

```


这段代码使用 GORM 的链式查询方法对订单表进行分组统计和条件过滤，主要目的是按日期汇总订单金额并筛选出满足条件的记录。以下是逐段解析：

---

### **1. 第一段代码：基础分组统计**
```go
rows, err := db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Rows()
for rows.Next() {
    ...
}
```
- **作用**：按订单创建日期分组，统计每日销售总额。
- **关键方法**：
  - `Select("date(created_at) as date, sum(amount) as total")`：提取日期部分（如 `2025-03-22`）并计算金额总和。
  - `Group("date(created_at)")`：按日期分组，生成每日汇总数据。
  - `Rows()`：执行查询并返回原始结果集，需手动遍历`rows.Next()`逐行读取数据。
- **生成 SQL**：
  ```sql
  SELECT DATE(created_at) AS date, SUM(amount) AS total 
  FROM orders 
  GROUP BY DATE(created_at);
  ```

---

### **2. 第二段代码：添加 HAVING 过滤**
```go
rows, err := db.Table("orders").Select(...).Group(...).Having("sum(amount) > ?", 100).Rows()
for rows.Next() {
    ...
}
```
- **作用**：在分组统计的基础上，筛选出单日总销售额超过 100 的记录。
- **新增方法**：
  - `Having("sum(amount) > ?", 100)`：在分组后过滤结果，类似 SQL 的 `HAVING` 子句，仅保留符合条件的分组。
- **生成 SQL**：
  ```sql
  SELECT DATE(created_at) AS date, SUM(amount) AS total 
  FROM orders 
  GROUP BY DATE(created_at)
  HAVING SUM(amount) > 100;
  ```

---

### **3. 第三段代码：结果映射到结构体**
```go
type Result struct {
    Date  time.Time
    Total int64
}
var results []Result
db.Table("orders").Select(...).Group(...).Having(...).Scan(&results)
```
- **作用**：将查询结果直接映射到 `Result` 结构体切片，无需手动遍历 `Rows()`。
- **关键方法**：
  - `Scan(&results)`：自动将查询结果填充到结构体切片中，字段名需与 `Select` 中的别名匹配（如 `date` 和 `total`）。
- **优势**：相比 `Rows()`，代码更简洁，且自动处理数据类型转换（如 `time.Time` 与日期字符串的映射）。

---

### **4. 代码对比与适用场景**
| **方法**         | **优点**                           | **缺点**                     | **适用场景**               |
|------------------|-----------------------------------|-----------------------------|--------------------------|
| `Rows()` + 循环   | 灵活处理复杂结果，适合大数据量逐行处理   | 需手动管理游标和错误，代码冗余     | 需要自定义解析或流式处理数据 |
| `Scan(&struct)`  | 自动映射结果，代码简洁               | 需预定义结构体，字段类型需严格匹配 | 快速获取结构化数据         |

---

### **5. 潜在问题与优化建议**
1. **日期格式处理**：
  - `DATE(created_at)` 假设数据库支持日期函数（如 MySQL），其他数据库可能需要调整语法（如 PostgreSQL 的 `DATE_TRUNC`）。
  - 若 `created_at` 包含时区，需统一时区设置（可在连接字符串中添加 `loc=Local`）。

2. **数据类型匹配**：
  - 确保 `amount` 字段类型与 `Total int64` 兼容。若 `amount` 为浮点数，应使用 `float64` 类型。

3. **性能优化**：
  - 为 `created_at` 和 `amount` 添加索引以加速分组和聚合操作。
  - 大数据量时，考虑分页查询或限制结果集（如 `.Limit(1000)`）。

4. **代码健壮性**：
  - 检查 `err` 并处理可能的错误（如数据库连接失败或 SQL 语法错误）。
  - 使用 `defer rows.Close()` 防止资源泄漏（仅 `Rows()` 需要）。

---

### **6. 完整示例与扩展**
```go
// 定义接收结果的结构体
type Result struct {
    Date  time.Time `gorm:"column:date"`
    Total int64     `gorm:"column:total"`
}

// 执行查询并映射到结构体
var results []Result
err := db.Table("orders").
    Select("DATE(created_at) as date, SUM(amount) as total").
    Group("DATE(created_at)").
    Having("SUM(amount) > ?", 100).
    Scan(&results).Error

if err != nil {
    log.Fatal("查询失败:", err)
}

// 打印结果
for _, r := range results {
    fmt.Printf("日期: %s, 总销售额: %d\n", r.Date.Format("2006-01-02"), r.Total)
}
```

---

### **总结**
- **核心功能**：通过分组和聚合统计每日销售额，支持条件过滤和结果自动映射。
- **最佳实践**：优先使用 `Scan` 简化代码，复杂场景结合 `Rows()` 手动处理。
- **扩展学习**：可结合 `Order` 排序、`Joins` 多表关联实现更复杂分析。


# Joins

指定关联条件

```go
rows, err := db.Table("users").Select("users.name, emails.email").Joins("left join emails on emails.user_id = users.id").Rows()
for rows.Next() {
    ...
}

db.Table("users").Select("users.name, emails.email").Joins("left join emails on emails.user_id = users.id").Scan(&results)

// 多个关联查询
db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Joins("JOIN credit_cards ON credit_cards.user_id = users.id").Where("credit_cards.number = ?", "411111111111").Find(&user)
```

以下是对代码的逐段解析及作用说明：

### 一、左连接查询与手动遍历结果
```go
rows, err := db.Table("users").
    Select("users.name, emails.email").
    Joins("left join emails on emails.user_id = users.id").
    Rows()
for rows.Next() {
    // 手动逐行读取数据
}
```
#### 作用解析
1. **多表联合查询**  
   通过 `Joins("left join emails...")` 实现 `users` 表与 `emails` 表的左连接，确保即使某些用户没有关联邮箱记录，仍会返回用户基本信息。
2. **字段筛选**  
   `Select("users.name, emails.email")` 指定仅返回用户名和邮箱，避免全字段查询带来的性能损耗。
3. **原生结果集处理**  
   `Rows()` 返回 `*sql.Rows` 对象，需配合 `rows.Next()` 和 `rows.Scan()` 手动遍历数据，适合大数据量或需要自定义解析的场景。

#### 生成 SQL
```sql
SELECT users.name, emails.email 
FROM users 
LEFT JOIN emails ON emails.user_id = users.id;
```

---

### 二、自动映射结果到结构体
```go
db.Table("users").
    Select("users.name, emails.email").
    Joins("left join emails on emails.user_id = users.id").
    Scan(&results)
```
#### 作用解析
1. **自动数据映射**  
   `Scan(&results)` 直接将查询结果映射到结构体切片，要求结构体字段名与查询结果的列名匹配（如 `Name` 对应 `users.name`，`Email` 对应 `emails.email`）。
2. **代码简洁性**  
   相比手动遍历，减少了错误处理和数据解析的代码量，适合快速获取结构化数据。

#### 结构体定义示例
```go
type Result struct {
    Name  string `gorm:"column:name"` // 映射 users.name
    Email string `gorm:"column:email"` // 映射 emails.email
}
```

---

### 三、多表关联查询与条件过滤
```go
db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").
    Joins("JOIN credit_cards ON credit_cards.user_id = users.id").
    Where("credit_cards.number = ?", "411111111111").
    Find(&user)
```
#### 作用解析
1. **多表联合查询**  
   通过链式调用 `Joins` 实现 `users` 表与 `emails`、`credit_cards` 表的内连接，仅返回同时满足所有关联条件的记录。
2. **条件过滤**
  - `emails.email = "jinzhu@example.org"`：直接嵌入到 JOIN 条件中，减少后续过滤的数据量。
  - `Where("credit_cards.number = ...")`：在最终结果集上过滤信用卡号，实现精确匹配。
3. **链式调用顺序**  
   GORM 会按调用顺序拼接 SQL 子句，最终生成的查询逻辑为：先关联邮箱和信用卡表，再应用全局过滤条件。

#### 生成 SQL
```sql
SELECT * FROM users
JOIN emails ON emails.user_id = users.id AND emails.email = 'jinzhu@example.org'
JOIN credit_cards ON credit_cards.user_id = users.id
WHERE credit_cards.number = '411111111111';
```

---

### 四、关键机制与优化建议
#### 1. **JOIN 类型选择**
- **LEFT JOIN**：保留主表（如 `users`）所有记录，即使关联表无匹配数据。
- **INNER JOIN**（默认）：仅返回两表完全匹配的记录，适合必须满足关联条件的场景。

#### 2. 性能优化
- **索引设计**：为 `emails.user_id` 和 `credit_cards.user_id` 添加索引，加速 JOIN 操作。
- **字段筛选**：避免 `SELECT *`，明确指定必要字段以减少数据传输量。

#### 3. 错误处理
- 检查 `err` 变量，处理数据库连接失败或 SQL 语法错误。
- 使用 `defer rows.Close()` 防止资源泄漏（仅 `Rows()` 需要）。

---

### 五、适用场景对比
| **方法**          | **适用场景**                               | **优点**                          | **缺点**                     |
|-------------------|------------------------------------------|----------------------------------|-----------------------------|
| `Rows()` + 手动遍历 | 大数据量分页、自定义解析逻辑               | 灵活控制内存占用和数据处理逻辑       | 代码冗余，需手动管理游标和错误 |
| `Scan(&struct)`   | 快速获取结构化数据                         | 自动类型转换，代码简洁              | 需预定义匹配的结构体           |
| 链式 `Joins`       | 复杂多表关联查询                           | 条件可动态组合，支持复杂过滤逻辑      | 需注意条件叠加顺序             |

---

通过合理选择 JOIN 类型和查询方法，可以高效实现从简单到复杂的多表关联查询需求。如需进一步优化复杂嵌套查询，可参考 GORM 的 `Preload` 预加载机制。


# Pluck

使用 Pluck 从模型中查询单个列作为集合。如果想查询多个列，应该使用 `Scan` 代替。

```go
var ages []int64
db.Find(&users).Pluck("age", &ages)

var names []string
db.Model(&User{}).Pluck("name", &names)

db.Table("deleted_users").Pluck("name", &names)

// Requesting more than one column? Do it like this:
db.Select("name, age").Find(&users)
```

# Scan

将 Scan 查询结果放入另一个结构体中。


```go
type Result struct {
    Name string
    Age  int
}

var result Result
db.Table("users").Select("name, age").Where("name = ?", 3).Scan(&result)

// Raw SQL
db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
```