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


```go
// 如果单个属性被更改了，更新它
db.Model(&user).Update("name", "hello")
//// UPDATE users SET name='hello', updated_at='2013-11-17 21:34:10' WHERE id=111;

// 使用组合条件更新单个属性
db.Model(&user).Where("active = ?", true).Update("name", "hello")
//// UPDATE users SET name='hello', updated_at='2013-11-17 21:34:10' WHERE id=111 AND active=true;

// 使用 `map` 更新多个属性，只会更新那些被更改了的字段
db.Model(&user).Updates(map[string]interface{}{"name": "hello", "age": 18, "actived": false})
//// UPDATE users SET name='hello', age=18, actived=false, updated_at='2013-11-17 21:34:10' WHERE id=111;

// 使用 `struct` 更新多个属性，只会更新那些被修改了的和非空的字段
db.Model(&user).Updates(User{Name: "hello", Age: 18})
//// UPDATE users SET name='hello', age=18, updated_at = '2013-11-17 21:34:10' WHERE id = 111;

// 警告： 当使用结构体更新的时候, GORM 只会更新那些非空的字段
// 例如下面的更新，没有东西会被更新，因为像 "", 0, false 是这些字段类型的空值
db.Model(&user).Updates(User{Name: "", Age: 0, Actived: false})
```


# 更新选中的字段

如果你在执行更新操作时只想更新或者忽略某些字段，可以使用 `Select`，`Omit`方法。

```go
db.Model(&user).Select("name").Updates(map[string]interface{}{"name": "hello", "age": 18, "actived": false})
//// UPDATE users SET name='hello', updated_at='2013-11-17 21:34:10' WHERE id=111;

db.Model(&user).Omit("name").Updates(map[string]interface{}{"name": "hello", "age": 18, "actived": false})
//// UPDATE users SET age=18, actived=false, updated_at='2013-11-17 21:34:10' WHERE id=111;
```

# 更新列钩子方法

上面的更新操作更新时会执行模型的 `BeforeUpdate` 和 `AfterUpdate` 方法，来更新 `UpdatedAt` 时间戳，并且保存他的 `关联`。如果你不想执行这些操作，可以使用 `UpdateColumn`，`UpdateColumns` 方法。

```go
// Update single attribute, similar with `Update`
db.Model(&user).UpdateColumn("name", "hello")
//// UPDATE users SET name='hello' WHERE id = 111;

// Update multiple attributes, similar with `Updates`
db.Model(&user).UpdateColumns(User{Name: "hello", Age: 18})
//// UPDATE users SET name='hello', age=18 WHERE id = 111;
```

# 批量更新

批量更新时，钩子函数不会执行

```go
db.Table("users").Where("id IN (?)", []int{10, 11}).Updates(map[string]interface{}{"name": "hello", "age": 18})
//// UPDATE users SET name='hello', age=18 WHERE id IN (10, 11);

// 使用结构体更新将只适用于非零值，或者使用 map[string]interface{}
db.Model(User{}).Updates(User{Name: "hello", Age: 18})
//// UPDATE users SET name='hello', age=18;

// 使用 `RowsAffected` 获取更新影响的记录数
db.Model(User{}).Updates(User{Name: "hello", Age: 18}).RowsAffected
```

# 带有表达式的SQL更新

```go
DB.Model(&product).Update("price", gorm.Expr("price * ? + ?", 2, 100))
//// UPDATE "products" SET "price" = price * '2' + '100', "updated_at" = '2013-11-17 21:34:10' WHERE "id" = '2';

DB.Model(&product).Updates(map[string]interface{}{"price": gorm.Expr("price * ? + ?", 2, 100)})
//// UPDATE "products" SET "price" = price * '2' + '100', "updated_at" = '2013-11-17 21:34:10' WHERE "id" = '2';

DB.Model(&product).UpdateColumn("quantity", gorm.Expr("quantity - ?", 1))
//// UPDATE "products" SET "quantity" = quantity - 1 WHERE "id" = '2';

DB.Model(&product).Where("quantity > 1").UpdateColumn("quantity", gorm.Expr("quantity - ?", 1))
//// UPDATE "products" SET "quantity" = quantity - 1 WHERE "id" = '2' AND quantity > 1;

```

# 在钩子函数中更新值

如果你想使用 `BeforeUpdate`、`BeforeSave`钩子函数修改更新的值，可以使用 `scope.SetColumn`方法，例如：


```go
func (user *User) BeforeSave(scope *gorm.Scope) (err error) {
  if pw, err := bcrypt.GenerateFromPassword(user.Password, 0); err == nil {
    scope.SetColumn("EncryptedPassword", pw)
  }
}
```

这段代码是GORM的`BeforeSave`钩子实现，用于在保存用户数据到数据库前对密码进行加密处理。以下是逐层解析：

---

### **1. 代码功能解析**
- **核心作用**：通过`bcrypt`算法对用户密码加密，并将加密后的值写入`EncryptedPassword`字段。
- **触发时机**：在GORM执行`Create`或`Save`操作前自动调用。
- **安全目标**：避免明文存储密码，防止数据泄露后被直接利用。

---

### **2. 代码逻辑拆解**
```go
func (user *User) BeforeSave(scope *gorm.Scope) (err error) {
  if pw, err := bcrypt.GenerateFromPassword(user.Password, 0); err == nil {
    scope.SetColumn("EncryptedPassword", pw)
  }
}
```

#### **(1) 方法签名**
- **接收者**：`*User`表示该钩子绑定到用户模型。
- **参数**：`scope *gorm.Scope`提供当前数据库操作的上下文，可修改字段值。
- **返回值**：返回错误时，GORM会终止操作并回滚事务。

#### **(2) 密码加密逻辑**
- **`bcrypt.GenerateFromPassword`**：
    - 第一个参数是明文字符串密码（需转为`[]byte`）。
    - 第二个参数`cost`为加密复杂度（示例中设为`0`，实际应指定如`10`，见网页5示例）。
    - 生成带随机盐的哈希值，保证同一密码每次加密结果不同。

#### **(3) 字段更新**
- **`scope.SetColumn`**：
    - 将加密后的哈希值写入`EncryptedPassword`字段。
    - 避免直接操作结构体字段，确保GORM能正确追踪变更。

#### **(4) 错误处理**
- 若加密失败（如密码为空），返回错误中断保存操作。
- 未显式返回错误时，GORM会继续执行后续流程。

---

### **3. 典型应用场景**
1. **用户注册/修改密码**：  
   在提交表单时，前端传递明文密码，后端通过此钩子自动加密存储。
2. **数据迁移**：  
   旧系统迁移到新系统时，可批量调用`Save`触发加密逻辑。
3. **安全性增强**：  
   结合`AfterFind`钩子，避免查询结果中泄露加密字段。

---

### **4. 潜在问题与优化建议**
#### **问题**
- **Cost参数问题**：代码中`cost=0`可能使用默认值（如`10`），但显式指定更安全（如`cost=12`）。
- **字段命名歧义**：若模型本身有`Password`字段，需确保与`EncryptedPassword`逻辑分离，避免混淆。

#### **优化建议**
```go
// 修改后的示例（显式指定cost，并增加错误返回）
func (user *User) BeforeSave(scope *gorm.Scope) (err error) {
  const cost = 12 // 明确加密复杂度
  hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), cost)
  if err != nil {
    return err
  }
  scope.SetColumn("EncryptedPassword", hashedPassword)
  return nil
}
```

---

### **5. 关联技术扩展**
- **密码验证**：需通过`bcrypt.CompareHashAndPassword`比对登录密码（见网页5的`CheckPasswordHash`方法）。
- **钩子跳过**：若需跳过加密（如管理员重置密码），可使用`db.Session(&gorm.Session{SkipHooks: true})`。
- **多字段处理**：如需同时更新多个字段，可在钩子中调用多次`SetColumn`或直接操作`scope.Change`方法。

--- 

通过这段代码，开发者能以最小侵入性实现密码安全存储，是GORM钩子在业务逻辑中典型的安全实践。

# 额外的更新选项

```go
// 在更新 SQL 语句中添加额外的 SQL 选项
db.Model(&user).Set("gorm:update_option", "OPTION (OPTIMIZE FOR UNKNOWN)").Update("name", "hello")
//// UPDATE users SET name='hello', updated_at = '2013-11-17 21:34:10' WHERE id=111 OPTION (OPTIMIZE FOR UNKNOWN);
```

这段代码使用 GORM（Go 的 ORM 库）执行数据库更新操作，并在生成的 SQL 语句中附加了特定于数据库的选项。以下是逐层解析：

---

### **1. 代码功能解析**
- **核心作用**：更新用户记录的 `name` 字段为 `"hello"`，并通过 `OPTION (OPTIMIZE FOR UNKNOWN)` 向 SQL Server 的查询优化器传递提示。
- **触发操作**：生成的 SQL 会更新 `name` 和自动维护的 `updated_at` 字段，同时在语句末尾添加数据库特定的优化指令。
- **适用场景**：解决参数嗅探导致的性能问题，强制生成更通用的执行计划。

---

### **2. 代码逻辑拆解**
```go
db.Model(&user)                     // 指定操作模型（关联到 users 表）
  .Set("gorm:update_option", "OPTION (OPTIMIZE FOR UNKNOWN)")  // 附加自定义 SQL 选项
  .Update("name", "hello")           // 更新 name 字段
```

#### **(1) 模型绑定 `db.Model(&user)`**
- 确定操作的表为 `users`（与 `User` 结构体映射）。
- 若 `user` 实例包含 `id` 字段（如 `111`），GORM 会自动添加 `WHERE id=111` 条件。

#### **(2) 设置更新选项 `.Set("gorm:update_option", ...)`**
- **`gorm:update_option`**：GORM 的保留键，用于向 UPDATE 语句末尾注入自定义 SQL 片段。
- **`OPTION (OPTIMIZE FOR UNKNOWN)`**：SQL Server 的查询提示，指示优化器在编译查询时假定所有局部变量的值为未知，避免因参数嗅探（Parameter Sniffing）生成不稳定的执行计划。

#### **(3) 执行更新 `.Update("name", "hello")`**
- 更新 `name` 字段为 `"hello"`。
- 若 `User` 模型定义了 `UpdatedAt` 字段，GORM 会自动设置当前时间到 `updated_at`。

---

### **3. 生成的 SQL 语句**
```sql
UPDATE users
SET name = 'hello', updated_at = '2013-11-17 21:34:10'
WHERE id = 111
OPTION (OPTIMIZE FOR UNKNOWN);  -- SQL Server 的查询提示
```

#### **关键点**
- **自动维护时间戳**：GORM 默认更新 `updated_at` 字段（需在模型中定义）。
- **条件推导**：通过 `user` 实例的 `id` 值自动添加 `WHERE id=111`。
- **自定义选项附加**：通过 `Set` 注入的文本直接拼接到 SQL 末尾。

---

### **4. 适用场景与注意事项**
#### **何时使用？**
- **参数嗅探问题**：当 SQL Server 因缓存参数值生成次优执行计划时，强制通用计划。
- **复杂查询优化**：需手动干预查询优化器的决策（如索引选择、连接顺序）。
- **数据库特定功能**：使用非标准 SQL 语法扩展功能（如 SQL Server 的 `OPTION` 子句）。

#### **注意事项**
1. **数据库兼容性**：
    - `OPTION (OPTIMIZE FOR UNKNOWN)` 仅适用于 **SQL Server**。
    - 其他数据库（如 MySQL、PostgreSQL）需使用其特定的提示语法（如 `FORCE INDEX`）。

2. **GORM 版本**：
    - 确保使用的 GORM 版本支持 `gorm:update_option`（v1 和 v2 均支持）。

3. **SQL 注入风险**：
    - 若动态生成 `OPTION` 内容，需严格验证输入，避免拼接恶意 SQL。

---

### **5. 扩展应用**
#### **(1) 其他 SQL 操作的自定义选项**
- **INSERT**：通过 `gorm:insert_option` 添加选项。
  ```go
  db.Set("gorm:insert_option", "ON CONFLICT (name) DO NOTHING")
  ```
- **DELETE**：通过 `gorm:delete_option` 添加选项。
  ```go
  db.Set("gorm:delete_option", "RETURNING id")
  ```

#### **(2) 动态生成优化提示**
根据条件动态附加不同提示：
```go
hint := "OPTION (MAXDOP 1)"
if isComplexQuery {
  hint = "OPTION (OPTIMIZE FOR UNKNOWN)"
}
db.Model(&data).Set("gorm:update_option", hint).Update(...)
```

---

### **6. 替代方案**
若需跨数据库兼容，可通过 **原生 SQL** 或 **回调（Hook）** 实现：
```go
// 原生 SQL（明确控制语句）
db.Exec("UPDATE users SET name=? WHERE id=? OPTION (OPTIMIZE FOR UNKNOWN)", "hello", user.ID)

// 通过 BeforeUpdate 钩子追加选项（需操作 *gorm.DB 的 Statement）
func (u *User) BeforeUpdate(db *gorm.DB) error {
  db.Statement.AddClause(clause.Expr{SQL: "OPTION (OPTIMIZE FOR UNKNOWN)"})
  return nil
}
```

---

### **总结**
这段代码演示了 GORM 如何通过 `Set` 方法注入数据库特定的 SQL 选项，适用于需要精细控制查询行为的场景。开发者在使用时需权衡数据库兼容性和功能需求，确保生成的 SQL 语法正确且安全。