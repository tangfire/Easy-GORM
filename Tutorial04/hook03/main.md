# 对象的生命周期

钩子是一个在 插入/查询/更新/删除 之前或之后被调用的方法。

如果你在一个模型中定义了特殊的方法，它将会在插入，更新，查询，删除的时候被自动调用，如果任何的回调抛出错误，GORM 将会停止将要执行的操作并且回滚当前的改变。

# 钩子

## 创建一个对象

可用于创建的钩子

```go
// 开启事务
BeforeSave
BeforeCreate
// 连表前的保存
// 更新时间戳 `CreatedAt`, `UpdatedAt`
// 保存自己
// 重载哪些有默认值和空的字段
// 链表后的保存
AfterCreate
AfterSave
// 提交或回滚事务
```

代码例子:

```go
func (u *User) BeforeSave() (err error) {
    if u.IsValid() {
        err = errors.New("can't save invalid data")
    }
    return
}

func (u *User) AfterCreate(scope *gorm.Scope) (err error) {
    if u.ID == 1 {
    scope.DB().Model(u).Update("role", "admin")
  }
    return
}
```

这段代码实现了 GORM 中的两个钩子方法，分别用于在保存（插入/更新）用户记录前进行数据验证，以及在创建用户记录后自动设置管理员角色。以下是对代码的逐层解析：

---

### 一、BeforeSave 钩子解析
```go
func (u *User) BeforeSave() (err error) {
    if u.IsValid() {
        err = errors.New("can't save invalid data")
    }
    return
}
```
**触发时机**：在用户记录被保存（插入或更新）到数据库之前自动调用。  
**核心逻辑**：
1. 通过 `u.IsValid()` 检查用户数据是否有效（具体校验逻辑需查看 `IsValid` 方法实现）。
2. 如果数据有效（`IsValid` 返回 `true`），则返回错误 `can't save invalid data`，阻止保存操作。
3. 若返回错误，GORM 将回滚事务，记录不会被保存到数据库。

**潜在问题**：  
代码逻辑存在矛盾。根据常见设计模式，`IsValid()` 通常返回 `true` 表示数据有效，此时应允许保存而非报错。开发者可能需要检查 `IsValid` 方法的实现逻辑，或调整条件为 `if !u.IsValid()`。

---

### 二、AfterCreate 钩子解析
```go
func (u *User) AfterCreate(scope *gorm.Scope) (err error) {
    if u.ID == 1 {
        scope.DB().Model(u).Update("role", "admin")
    }
    return
}
```
**触发时机**：在用户记录成功插入数据库后自动调用。  
**核心逻辑**：
1. 检查用户 ID 是否为 `1`（通常为第一条记录的主键）。
2. 如果是，则通过 `Update` 方法将该用户的 `role` 字段更新为 `admin`，实现自动赋予管理员权限的功能。

**技术细节**：
- **参数差异**：旧版 GORM 使用 `*gorm.Scope` 作为钩子参数，新版（如 v2.0+）改用 `tx *gorm.DB`，需注意版本兼容性。
- **更新方式**：在钩子内直接更新当前记录可能导致递归调用（例如触发 `BeforeUpdate`/`AfterUpdate` 钩子），建议通过事务上下文 `tx` 操作。

---

### 三、钩子与 GORM 工作流的整合
1. **数据验证流程**：  
   `BeforeSave` 用于全局校验，确保非法数据（如空字段、格式错误）不会进入数据库。例如，可在 `IsValid` 中验证邮箱格式或密码强度。

2. **自动化逻辑**：  
   `AfterCreate` 适用于后置操作，如发送欢迎邮件、初始化用户权限或记录日志。此处通过 ID 判断首用户并升级角色，是一种常见的初始化管理员策略。

---

### 四、潜在改进建议
1. **修正校验逻辑**：  
   修改条件为 `if !u.IsValid()`，确保仅当数据无效时阻止保存。

2. **适配新版 GORM**：  
   若使用新版 GORM，建议将 `AfterCreate` 参数改为 `tx *gorm.DB`：
   ```go
   func (u *User) AfterCreate(tx *gorm.DB) (err error) {
       if u.ID == 1 {
           tx.Model(u).Update("role", "admin")
       }
       return
   }
   ```

3. **避免递归更新**：  
   在 `AfterCreate` 中更新记录时，可通过跳过钩子避免循环触发：
   ```go
   tx.Session(&gorm.Session{SkipHooks: true}).Update("role", "admin")
   ```

---

### 五、关联技术点（来自搜索结果）
- **跳过钩子**：通过 `SkipHooks` 会话模式可绕过钩子执行。
- **并发问题**：高并发下需注意主键冲突风险（如旧版 GORM v1.24.6 的 `Save` 方法缺陷）。
- **零值更新**：`Update` 会强制更新字段（包括零值），而 `Updates` 默认忽略零值。

通过合理运用钩子，开发者可以实现高度定制化的数据库操作流程，同时需关注版本差异和潜在边界条件。


---

注意，在 GORM 中的保存/删除 操作会默认进行事务处理，所以在事物中，所有的改变都是无效的，直到它被提交为止:

```go
func (u *User) AfterCreate(tx *gorm.DB) (err error) {
    tx.Model(u).Update("role", "admin")
    return
}
```

# 更新一个对象

可用于更新的钩子

```go
// 开启事务
BeforeSave
BeforeUpdate
// 链表前的保存
// 更新时间戳 `UpdatedAt`
// 保存自身
// 链表后的保存
AfterUpdate
AfterSave
// 提交或回滚的事务
```

代码示例:

```go
func (u *User) BeforeUpdate() (err error) {
    if u.readonly() {
        err = errors.New("read only user")
    }
    return
}

// 在事务结束后，进行更新数据
func (u *User) AfterUpdate(tx *gorm.DB) (err error) {
  if u.Confirmed {
    tx.Model(&Address{}).Where("user_id = ?", u.ID).Update("verfied", true)
  }
    return
}
```

# 删除一个对象

可用于删除的钩子

```go
// 开启事务
BeforeDelete
// 删除自身
AfterDelete
// 提交或回滚事务
```

代码示例:

```go
// 在事务结束后进行更新数据
func (u *User) AfterDelete(tx *gorm.DB) (err error) {
  if u.Confirmed {
    tx.Model(&Address{}).Where("user_id = ?", u.ID).Update("invalid", false)
  }
    return
}
```

# 查询一个对象

可用于查询的钩子

```go
// 从数据库中读取数据
// 加载之前 (急于加载)
AfterFind
```


代码示例：

```go
func (u *User) AfterFind() (err error) {
  if u.MemberShip == "" {
    u.MemberShip = "user"
  }
    return
}
```
