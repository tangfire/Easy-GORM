# 创建插件

GORM 本身由 `Callbacks` 提供支持，因此你可以根据需要完全自定义GORM。

# 注册新的 callback

将 callback 注册进如 callbacks：

```go
func updateCreated(scope *Scope) {
    if scope.HasColumn("Created") {
        scope.SetColumn("Created", NowFunc())
    }
}

db.Callback().Create().Register("update_created_at", updateCreated)
// 注册 Create 进程的回调
```


这段代码用于在 GORM 的 **Create 操作（数据插入）** 过程中注册一个自定义回调函数，其核心功能是 **自动为 `Created` 字段设置当前时间戳**。以下是逐层解析：

---

### 一、代码功能分解
#### 1. `updateCreated` 函数
```go
func updateCreated(scope *Scope) {
    if scope.HasColumn("Created") {
        scope.SetColumn("Created", NowFunc())
    }
}
```
- **作用**：在数据插入时，若模型包含 `Created` 字段，则将其值设为当前时间。
- **关键点**：
    - **`scope *Scope`**：GORM 操作的作用域对象，包含当前操作的上下文（如模型、字段、SQL 语句等）。
    - **`HasColumn("Created")`**：检查当前模型是否定义了 `Created` 列，避免在不支持的表中误操作。
    - **`SetColumn("Created", NowFunc())`**：将 `Created` 字段的值设置为 `NowFunc()`（通常是当前时间函数，如 `time.Now()`）。

#### 2. 回调注册
```go
db.Callback().Create().Register("update_created_at", updateCreated)
```
- **作用**：将 `updateCreated` 函数注册到 GORM 的 **Create 操作回调链**中。
- **关键点**：
    - **`Create()`**：指定回调作用于 Create 操作（插入数据）。
    - **`Register`**：注册名为 `update_created_at` 的回调函数，GORM 会在每次插入数据时自动触发。

---

### 二、运行机制与场景
#### 1. **自动时间戳管理**
- 当执行 `db.Create(&user)` 时，回调会自动填充 `Created` 字段，无需手动赋值：
  ```go
  user := User{Name: "Alice"}
  db.Create(&user) // Created 字段自动设置为当前时间
  ```
- **适用场景**：记录数据创建时间（如日志、订单创建时间）。

#### 2. **条件性操作**
- 通过 `HasColumn("Created")` 确保仅当模型包含该字段时才执行操作，避免因表结构差异导致的错误。

#### 3. **回调链的执行**
- GORM 的 Create 操作包含多个预定义回调（如 `gorm:before_create`、`gorm:create`），自定义回调通过 `Register` 插入到链中，默认执行顺序取决于注册位置（可通过 `Before()` 或 `After()` 调整顺序）。

---

### 三、扩展与注意事项
#### 1. **与 GORM 内置时间追踪的区别**
- GORM 默认支持 `CreatedAt` 和 `UpdatedAt` 字段的自动更新，但此代码演示了 **自定义字段名**（如 `Created`）的扩展实现。

#### 2. **全局性影响**
- 回调注册在 `*gorm.DB` 实例上，属于 **全局配置**，所有通过该实例的 Create 操作均会触发此回调。若需会话级隔离，需创建新的 `*gorm.DB` 实例。

#### 3. **错误处理**
- 若回调返回错误（此代码未处理），GORM 会中止操作并回滚事务：
  ```go
  func updateCreated(scope *Scope) error {
      if scope.HasColumn("Created") {
          return scope.SetColumn("Created", NowFunc())
      }
      return nil
  }
  ```

---

### 四、与其他技术的结合
#### 1. **多回调协作**
- 可注册多个回调实现复杂逻辑，例如：
  ```go
  db.Callback().Create().Before("gorm:create").Register("validate_data", validateData)
  db.Callback().Create().After("gorm:create").Register("send_notification", sendNotification)
  ```

#### 2. **动态条件扩展**
- 结合 `scope` 上下文实现动态逻辑，例如根据环境变量决定是否填充时间：
  ```go
  if os.Getenv("ENV") == "prod" && scope.HasColumn("Created") {
      scope.SetColumn("Created", NowFunc())
  }
  ```

---

### 总结
此代码通过 GORM 的回调机制实现了自定义字段的自动填充，展示了 ORM 框架的灵活性和扩展性。实际开发中，类似的回调可用于审计日志、数据加密、业务规则校验等场景。





# 删除已有的 callback

从 callbacks 中删除一个 callback：

```go
db.Callback().Create().Remove("gorm:create")
// delete callback `gorm:create` from Create callbacks
```

这段代码的作用是**从 GORM 的 `Create` 操作（数据插入）的回调链中移除名为 `gorm:create` 的内置回调函数**。以下是具体解析：

---

### 一、代码功能说明
```go
db.Callback().Create().Remove("gorm:create")
```
1. **`db.Callback()`**  
   获取 GORM 的全局回调管理器，用于操作所有数据库操作（如 Create、Update、Delete）的回调链。
2. **`.Create()`**  
   指定针对 `Create` 操作的回调链。
3. **`.Remove("gorm:create")`**  
   从回调链中移除名为 `gorm:create` 的默认回调函数。此函数是 GORM 内置的实现数据插入的核心逻辑。

---

### 二、删除 `gorm:create` 的影响
1. **破坏默认数据插入行为**  
   `gorm:create` 是 GORM 生成并执行 `INSERT` 语句的关键回调。移除后，执行 `db.Create(&model)` 时，数据将无法插入到数据库，因为核心逻辑被删除。
2. **可能影响关联操作**  
   如果模型中定义了关联（如外键约束、自动保存关联记录），移除 `gorm:create` 可能导致关联数据处理异常。
3. **事务流程中断**  
   GORM 默认在 `Create` 操作中开启事务，而事务提交/回滚的回调（如 `gorm:commit_or_rollback_transaction`）依赖 `gorm:create` 的执行结果。移除后可能导致事务无法正常关闭。

---

### 三、适用场景
1. **完全自定义插入逻辑**  
   如果需要绕过 GORM 的默认插入行为（例如手动编写原生 SQL 或通过其他方式插入数据），可移除 `gorm:create` 并注册自定义回调。
2. **性能优化实验**  
   在极端性能优化场景中，若需要绕过 ORM 的反射和动态 SQL 生成逻辑，可移除默认回调（需自行实现底层逻辑）。
3. **调试与测试**  
   临时移除默认回调以隔离问题，例如验证某个回调是否导致插入失败。

---

### 四、注意事项
1. **谨慎操作**  
   此操作会破坏 GORM 的默认功能，需确保已准备好替代逻辑。例如：
   ```go
   // 移除默认回调
   db.Callback().Create().Remove("gorm:create")
   // 注册自定义插入逻辑
   db.Callback().Create().Register("custom_create", myCustomCreateFunction)
   ```
2. **作用域问题**  
   此操作作用于全局 `*gorm.DB` 实例。若需会话级隔离，应通过 `db.Session(&gorm.Session{})` 创建新会话。
3. **恢复默认行为**  
   可通过重新注册回调或重启应用恢复默认逻辑，但更推荐在插件中动态管理回调状态。

---

### 五、扩展知识：GORM 的默认 Create 回调链
根据源码，`Create` 操作的默认回调链包括：
1. `gorm:begin_transaction`：开启事务
2. `gorm:before_create`：触发 `BeforeCreate` 钩子
3. `gorm:save_before_associations`：保存前置关联数据
4. `gorm:create`：核心插入逻辑（被此代码移除）
5. `gorm:save_after_associations`：保存后置关联数据
6. `gorm:after_create`：触发 `AfterCreate` 钩子
7. `gorm:commit_or_rollback_transaction`：提交或回滚事务

移除 `gorm:create` 后，回调链将跳过第 4 步，导致后续步骤无法执行。

---

### 总结
此代码通过移除 `gorm:create` 回调，禁用了 GORM 的默认数据插入逻辑，适用于需要完全自定义插入行为的场景。使用时需严格评估风险，并确保有完整的替代方案。




# 替换 callback

替换拥有相同名字的 callback ：


```go
db.Callback().Create().Replace("gorm:create", newCreateFunction)
// replace callback `gorm:create` with new function `newCreateFunction` for Create process
```

这段代码的作用是**将 GORM 框架中 `Create` 操作（数据插入）的默认回调函数 `gorm:create` 替换为自定义函数 `newCreateFunction`**，从而实现对核心插入逻辑的完全自定义。以下是分步解析：

---

### 一、代码功能解析
```go
db.Callback().Create().Replace("gorm:create", newCreateFunction)
```
1. **`db.Callback().Create()`**  
   获取 GORM 的 `Create` 操作回调管理器，用于操作插入相关的回调链。
2. **`Replace("gorm:create", newCreateFunction)`**
    - **`gorm:create`**：GORM 默认的插入逻辑回调名称，负责生成并执行 `INSERT` 语句。
    - **`newCreateFunction`**：用户自定义函数，将替代默认逻辑，实现自定义插入行为。

---

### 二、运行机制与影响
#### 1. **原默认逻辑的覆盖**
- **原 `gorm:create` 的作用**：  
  默认回调会基于模型结构生成 `INSERT` 语句，处理字段映射、关联数据保存等核心逻辑。
- **替换后的行为**：  
  所有 `Create` 操作（如 `db.Create(&user)`）将执行 `newCreateFunction` 而非默认逻辑。例如，用户可实现以下自定义行为：
    - 修改 SQL 生成规则（如插入前加密字段）；
    - 添加日志记录或审计功能；
    - 绕过 ORM 反射，直接调用原生 SQL。

#### 2. **回调链的完整性**
- **默认回调链顺序**：  
  GORM 的 `Create` 操作默认回调链包括：事务开启 → `BeforeCreate` 钩子 → 前置关联保存 → `gorm:create` → 后置关联保存 → `AfterCreate` 钩子 → 事务提交/回滚。
- **替换后的依赖关系**：  
  若 `newCreateFunction` 未正确处理关联保存或事务逻辑（如未调用 `SaveAfterAssociations`），可能导致数据不一致。

---

### 三、适用场景
#### 1. **深度定制插入逻辑**
- **性能优化**：  
  通过原生 SQL 替代 ORM 的动态反射，减少生成 SQL 的开销。
- **数据加密**：  
  在插入前自动加密敏感字段（如密码、手机号）。
- **日志与审计**：  
  记录每次插入操作的元数据（如操作人、时间戳）。

#### 2. **兼容性扩展**
- **支持特殊数据库特性**：  
  如 MySQL 的 `ON DUPLICATE KEY UPDATE` 或 PostgreSQL 的 `RETURNING` 子句。
- **多租户分片**：  
  根据业务规则动态选择插入的目标分片表。

#### 3. **调试与测试**
- **Mock 数据库行为**：  
  在单元测试中模拟插入结果，避免真实数据库依赖。

---

### 四、注意事项
#### 1. **全局性影响**
- 替换操作作用于全局 `*gorm.DB` 实例，所有通过该实例的 `Create` 操作均受影响。若需会话级隔离，需通过 `db.Session(&gorm.Session{})` 创建新会话。

#### 2. **逻辑完整性**
- **保留必要步骤**：  
  若自定义函数未实现默认回调的关键逻辑（如事务管理），需手动补充。例如：
  ```go
  func newCreateFunction(db *gorm.DB) {
      // 自定义插入逻辑
      db.Exec("INSERT INTO users ...")
      // 手动触发后置关联保存
      if err := SaveAfterAssociations(true)(db); err != nil {
          db.AddError(err)
      }
  }
  ```

#### 3. **错误处理**
- **继承错误传递机制**：  
  GORM 通过 `db.Error` 传递错误，自定义函数需通过 `db.AddError()` 报告错误以确保后续回调（如事务回滚）正常执行。

---

### 五、示例扩展
#### 自定义插入函数实现日志记录
```go
func newCreateFunction(db *gorm.DB) {
    // 记录插入前数据状态
    log.Printf("Inserting data: %+v", db.Statement.Dest)
    // 调用默认逻辑（可选）
    // callOriginalCreate(db)
    // 或完全自定义逻辑
    db.Exec("INSERT ...")
    // 记录插入结果
    log.Printf("Inserted rows affected: %d", db.RowsAffected)
}
```

---

### 总结
此代码通过替换 `gorm:create` 回调，实现了对 GORM 数据插入逻辑的深度定制，适用于需要突破 ORM 默认行为的场景。开发时需谨慎处理关联数据、事务和错误传递，避免破坏数据一致性。


# 注册 callback 的顺序

在注册 callbacks 时设置顺序：

```go
db.Callback().Create().Before("gorm:create").Register("update_created_at", updateCreated)
db.Callback().Create().After("gorm:create").Register("update_created_at", updateCreated)
db.Callback().Query().After("gorm:query").Register("my_plugin:after_query", afterQuery)
db.Callback().Delete().After("gorm:delete").Register("my_plugin:after_delete", afterDelete)
db.Callback().Update().Before("gorm:update").Register("my_plugin:before_update", beforeUpdate)
db.Callback().Create().Before("gorm:create").After("gorm:before_create").Register("my_plugin:before_create", beforeCreate)
```

# 自带的 Callbacks

GORM 在处理 CRUD 操作时自带了一些 Callback，建议你在写插件前先熟悉这些 Callback：

- Create callbacks(略)

- Update callbacks(略)

- Query callbacks(略)

- Delete callbacks(略)

- RowQuery callbacks - 默认没有注册的 Callbacks

你可以用以下的方法来注册你的 Callback：

```go
func updateTableName(scope *gorm.Scope) {
  scope.Search.Table(scope.TableName() + "_draft") // append `_draft` to table name
}

db.Callback().RowQuery().Register("publish:update_table_name", updateTableName)
```



这段代码的作用是**在 GORM 执行原生 SQL 查询时，动态修改操作的表名，为其添加 `_draft` 后缀**。以下是逐层解析：

---

### 一、代码功能分解
#### 1. 回调函数 `updateTableName`
```go
func updateTableName(scope *gorm.Scope) {
    scope.Search.Table(scope.TableName() + "_draft") // 为表名添加 _draft 后缀
}
```
- **`scope.TableName()`**：  
  获取当前模型对应的默认表名（例如 `User` 结构体默认对应表名 `users`）。
- **`scope.Search.Table(...)`**：  
  动态修改当前查询的目标表名（例如将 `users` 改为 `users_draft`）。

#### 2. 回调注册
```go
db.Callback().RowQuery().Register("publish:update_table_name", updateTableName)
```
- **`RowQuery` 回调**：  
  当通过 `Raw` 或 `Exec` 执行原生 SQL 时触发。
- **`Register`**：  
  将 `updateTableName` 注册到 `RowQuery` 回调链中，命名为 `publish:update_table_name`。

---

### 二、运行机制与效果
#### 1. 触发场景
以下操作会触发此回调：
```go
// 场景 1：Raw 查询
db.Raw("SELECT * FROM users WHERE id = ?", 1).Scan(&result)
// 实际执行：SELECT * FROM users_draft WHERE id = 1

// 场景 2：Exec 更新
db.Exec("UPDATE users SET name = ?", "Bob")
// 实际执行：UPDATE users_draft SET name = 'Bob'
```

#### 2. 表名修改逻辑
- **自动后缀追加**：  
  无论原始 SQL 中的表名是什么，都会被替换为 `<原表名>_draft`。
- **作用域隔离**：  
  修改仅在当前查询生命周期内有效，不会影响其他操作。

---

### 三、适用场景
#### 1. 草稿数据隔离
- 将正式数据（`users`）与草稿数据（`users_draft`）分离，例如 CMS 系统的内容编辑暂存功能。
- 应用场景：
  ```go
  // 编辑文章时保存到草稿表
  db.Exec("UPDATE articles SET title = ? WHERE id = ?", "New Title", 100)
  // 实际执行：UPDATE articles_draft ...
  ```

#### 2. 多环境数据隔离
- 在测试环境中自动操作 `_draft` 表，避免污染正式数据。
- 例如：
  ```go
  // 测试脚本清理数据
  db.Exec("DELETE FROM users") 
  // 实际执行：DELETE FROM users_draft（测试环境专用）
  ```

#### 3. 动态分表
- 根据业务规则动态切换表名（如按日期分表），但此处固定添加 `_draft` 后缀。

---

### 四、注意事项
#### 1. 原生 SQL 的匹配风险
- 若 SQL 中表名通过别名或字符串拼接方式编写，可能无法被正确替换：
  ```go
  db.Exec("SELECT * FROM users AS u") // 替换后：users_draft AS u（可能不符合预期）
  db.Exec("SELECT * FROM 'users'")     // 替换后：'users_draft'（需注意引号）
  ```

#### 2. 回调作用范围
- **全局性**：  
  注册到 `db` 实例的回调会影响所有通过该实例的原生查询。
- **会话级隔离**：  
  若需仅在特定场景启用，可通过新会话临时注册：
  ```go
  newDB := db.Session(&gorm.Session{})
  newDB.Callback().RowQuery().Register("draft_table", updateTableName)
  ```

#### 3. 与其他回调的冲突
- 若其他回调也修改表名（如分表插件），需确保执行顺序或逻辑兼容性。

---

### 五、扩展用法
#### 1. 条件性表名修改
```go
func updateTableName(scope *gorm.Scope) {
    if isDraftMode { // 根据业务条件判断
        scope.Search.Table(scope.TableName() + "_draft")
    }
}
```

#### 2. 多后缀支持
```go
func updateTableName(scope *gorm.Scope) {
    scope.Search.Table(fmt.Sprintf("%s_%s", scope.TableName(), env)) 
    // 例如：users_dev、users_prod
}
```

---

### 总结
这段代码通过 GORM 的 `RowQuery` 回调，在原生 SQL 执行阶段动态修改表名，实现了 **数据表名的透明化切换**。适用于草稿隔离、多环境数据管理等场景，但需注意 SQL 兼容性和作用域隔离。



