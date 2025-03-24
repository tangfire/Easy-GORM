GORM 默认在事务中执行单个 `create`， `update`， `delete` 操作，以确保数据库数据完整性。

如果你想将多个 `create`， `update`， `delete` 当成一个原子性操作，`Transaction` 就是为了这个而创造的。

# 事务

要在事务中执行一组操作，正常的流程如下所示。

```go
// 开启事务
tx := db.Begin()

// 在事务中执行一些数据库操作 （从这里开始使用 'tx'，而不是 'db'）
tx.Create(...)

// ...

// 发生错误回滚事务
tx.Rollback()

// 或者提交这个事务
tx.Commit()
```

# 具体例子

```go
func CreateAnimals(db *gorm.DB) err {
  // 注意在事务中要使用 tx 作为数据库句柄
  tx := db.Begin()
  defer func() {
    if r := recover(); r != nil {
      tx.Rollback()
    }
  }()

  if tx.Error != nil {
    return err
  }

  if err := tx.Create(&Animal{Name: "Giraffe"}).Error; err != nil {
     tx.Rollback()
     return err
  }

  if err := tx.Create(&Animal{Name: "Lion"}).Error; err != nil {
     tx.Rollback()
     return err
  }

  return tx.Commit().Error
}
```


这段代码是使用 GORM 实现数据库事务的典型示例，主要目的是保证「创建两个 Animal 记录」操作的原子性。以下是逐段解析：

---

### 一、事务初始化与异常捕获
```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
```
- `db.Begin()` 启动一个新事务，返回事务句柄 `tx`（后续操作必须基于 `tx` 而非原 `db`）。
- `defer` 语句通过 `recover()` 捕获代码执行过程中的 panic（如空指针等意外错误），若发生异常则自动回滚事务，防止数据处于中间状态。

---

### 二、事务启动错误检查
```go
if tx.Error != nil {
    return err
}
```
- 检查事务是否成功启动。例如：数据库连接中断、权限不足等情况会导致 `tx.Error` 非空，此时直接终止流程。

---

### 三、原子性操作实现
```go
if err := tx.Create(&Animal{Name: "Giraffe"}).Error; err != nil {
    tx.Rollback()
    return err
}
// 同理创建 Lion...
```
- **事务内操作**：通过 `tx.Create()` 插入数据（而非直接使用 `db.Create()`），保证操作在事务上下文中执行。
- **错误回滚**：若任意一步插入失败（如违反唯一约束、字段类型不匹配等），立即调用 `Rollback()` 撤销当前事务的所有操作。
- 此处 `Animal` 结构体需满足 GORM 模型定义约定（例如主键字段 `ID` 或标记 `gorm:"primary_key"`）。

---

### 四、事务提交
```go
return tx.Commit().Error
```
- 若所有操作成功，调用 `Commit()` 提交事务，使所有变更永久生效。提交时仍需检查错误（例如网络波动导致提交失败）。

---

### 关键设计解读
1. **原子性保证**：通过「要么全部成功，要么全部回滚」机制，避免出现只创建 Giraffe 而 Lion 失败的中间状态。
2. **防御性编程**：通过 `defer` 和 `recover()` 处理 panic，增强了代码鲁棒性。
3. **模型依赖**：代码中的 `Animal` 结构体需遵循 GORM 的模型定义规则（例如字段标签 `gorm:"column:beast_id"` 或时间戳字段 `CreatedAt` 的自动跟踪）。

---

### 扩展建议
- 若需要记录事务日志，可在 `Commit()` 前添加审计逻辑。
- 对于高频事务，可结合 `context.WithTimeout` 添加超时控制，防止长事务阻塞数据库连接池。