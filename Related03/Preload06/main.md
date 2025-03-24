# 预加载

```go
db.Preload("Orders").Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4);

db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4) AND state NOT IN ('cancelled');

db.Where("state = ?", "active").Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
//// SELECT * FROM users WHERE state = 'active';
//// SELECT * FROM orders WHERE user_id IN (1,2) AND state NOT IN ('cancelled');

db.Preload("Orders").Preload("Profile").Preload("Role").Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4); // has many
//// SELECT * FROM profiles WHERE user_id IN (1,2,3,4); // has one
//// SELECT * FROM roles WHERE id IN (4,5,6); // belongs to
```

这段代码演示了如何使用GORM库中的`Preload`方法进行关联数据的预加载，以优化查询并避免N+1问题。以下是每个示例的详细解释：

---

### 1. 基本预加载
```go
db.Preload("Orders").Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4);
```
- **作用**：查询所有用户及其关联的所有订单。
- **流程**：
    1. 先执行`SELECT * FROM users`获取所有用户数据。
    2. 提取用户ID（如1,2,3,4），执行`SELECT * FROM orders WHERE user_id IN (1,2,3,4)`加载这些用户的订单。
- **关联类型**：假设`User`模型通过`has many`关联`Orders`，GORM自动通过`user_id`外键关联。

---

### 2. 带条件的预加载
```go
db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4) AND state NOT IN ('cancelled');
```
- **作用**：查询所有用户，但仅预加载状态**未取消**的订单。
- **条件参数**：
    - 第一个参数是关联名`"Orders"`。
    - 第二个参数是SQL条件语句`"state NOT IN (?)"`，`?`会被替换为后续参数（`"cancelled"`）。
- **结果**：生成的SQL在关联表中添加了`AND state NOT IN ('cancelled')`条件，过滤掉已取消的订单。

---

### 3. 主查询条件与预加载条件结合
```go
db.Where("state = ?", "active").Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
//// SELECT * FROM users WHERE state = 'active';
//// SELECT * FROM orders WHERE user_id IN (1,2) AND state NOT IN ('cancelled');
```
- **作用**：查询所有状态为`active`的用户，并预加载他们未取消的订单。
- **主查询条件**：`WHERE state = 'active'`过滤用户，假设结果中用户ID为1,2。
- **预加载条件**：仅加载这些用户的订单，且订单状态不是`cancelled`。
- **关键点**：主查询的条件影响关联查询的`IN`列表（用户ID范围），预加载条件过滤关联表数据。

---

### 4. 预加载多个关联
```go
db.Preload("Orders").Preload("Profile").Preload("Role").Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4);     // has many
//// SELECT * FROM profiles WHERE user_id IN (1,2,3,4);   // has one
//// SELECT * FROM roles WHERE id IN (4,5,6);             // belongs to
```
- **作用**：一次查询用户，并预加载多个关联（订单、个人资料、角色）。
- **不同关联类型的处理**：
    - **Has Many（Orders）**：通过`user_id`外键批量查询订单。
    - **Has One（Profile）**：同样通过`user_id`查询，每个用户只有一个个人资料。
    - **Belongs To（Role）**：用户属于角色，外键`role_id`存储在用户表中，因此查询`roles WHERE id IN (用户表的role_id集合)`。
- **效率**：通过批量查询（`IN`语句）减少数据库访问次数，避免N+1问题。

---

### 总结
- **Preload核心作用**：通过批量查询关联数据，减少数据库查询次数。
- **条件传递**：可通过附加参数为关联表添加过滤条件。
- **关联类型差异**：`has many`/`has one`通过外键查询关联表，`belongs to`通过主键反向查询。
- **适用场景**：需要同时访问主模型及其关联数据时（如用户详情页展示订单、资料、角色）。

通过合理使用`Preload`，可以显著优化复杂关联数据的查询性能。

# 自动预加载


始终自动预加载的关联

```go
type User struct {
  gorm.Model
  Name       string
  CompanyID  uint
  Company    Company `gorm:"PRELOAD:false"` //没有预加载
  Role       Role                           // 已经预加载
}

db.Set("gorm:auto_preload", true).Find(&users)
```

这段代码演示了GORM中模型关联与自动预加载的混合使用方式，结合结构体标签和全局配置来控制关联数据的加载行为。以下是对代码的逐层解析：

---

### 一、模型结构体定义分析
```go
type User struct {
    gorm.Model
    Name      string
    CompanyID uint
    Company   Company `gorm:"PRELOAD:false"` // 禁用自动预加载
    Role      Role                           // 启用自动预加载
}
```

1. **基础字段**
    - `gorm.Model` 是GORM内置的基础模型，包含`ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`四个字段
    - `Name`和`CompanyID`是普通字段，对应数据库表的列

2. **关联字段的特殊配置**
    - `Company Company `gorm:"PRELOAD:false"``  
      通过结构体标签显式禁用预加载，即使全局开启自动预加载，该关联也不会被加载
    - `Role Role`  
      未加特殊标签的关联字段，会遵循全局自动预加载配置

---

### 二、查询代码解析
```go
db.Set("gorm:auto_preload", true).Find(&users)
```

1. **全局自动预加载配置**
    - `Set("gorm:auto_preload", true)` 启用全局自动预加载，会尝试预加载所有未标记`PRELOAD:false`的关联
    - 该设置对同一DB实例的后续查询持续生效

2. **关联加载的实际表现**
    - **Role关联**  
      自动执行预加载SQL：
      ```sql
      SELECT * FROM roles WHERE id IN (用户的role_id列表);
      ```
    - **Company关联**  
      由于结构体标签的覆盖，不会生成预加载SQL，需要手动加载：
      ```go
      db.Preload("Company").Find(&users)
      ```

---

### 三、关联类型推测（基于字段定义）
1. **Company关联类型**
    - 存在`CompanyID`外键字段，推测为`belongs to`关系  
      关联查询逻辑：通过`users.company_id`查找`companies`表

2. **Role关联类型**
    - 无显式外键字段，可能通过`user_roles`中间表实现多对多关系  
      或通过默认外键`role_id`实现一对多关系

---

### 四、配置优先级说明
当同时存在多种预加载配置时，优先级顺序为：
```
结构体标签 > 查询时Preload() > 全局auto_preload配置
```
本例中：
- `Company`的结构体标签`PRELOAD:false`优先级最高
- `Role`继承全局`auto_preload:true`配置

---

### 五、典型使用场景
1. **性能优化场景**  
   对不常用的关联（如用户所属公司信息）禁用自动预加载，避免不必要的JOIN查询

2. **高频访问场景**  
   对常用关联（如用户角色信息）保持自动预加载，减少N+1查询问题

3. **敏感数据控制**  
   防止自动加载包含敏感信息的关联（如用户隐私档案）

---

### 六、扩展建议
1. **关联加载验证**  
   建议通过GORM的Debug模式查看实际生成的SQL：
   ```go
   db.Debug().Set("gorm:auto_preload", true).Find(&users)
   ```

2. **替代配置方式**  
   若需更精细控制，可使用链式预加载：
   ```go
   db.Preload("Role").Find(&users) // 显式指定加载Role
   ```

3. **关联类型明确定义**  
   推荐通过`gorm:"foreignKey:CompanyID"`等标签明确关联类型，避免隐式推断带来的歧义

该代码模式适用于需要平衡开发便利性与查询性能的中大型项目，通过结构体标签与全局配置的组合，实现关联加载策略的集中管理。


# 嵌套预加载

```go
db.Preload("Orders.OrderItems").Find(&users)
db.Preload("Orders", "state = ?", "paid").Preload("Orders.OrderItems").Find(&users)

```

这段代码演示了在 GORM 中如何通过 `Preload` 方法实现 **多层嵌套关联数据的预加载** 和 **带条件的关联过滤**。以下是逐行解析：

---

### 1. 基础嵌套预加载
```go
db.Preload("Orders.OrderItems").Find(&users)
```
**作用**：查询所有用户，并预加载每个用户的 **所有订单** 及每个订单的 **所有订单项**。

**生成的 SQL**：
```sql
-- 第一步：查询所有用户
SELECT * FROM users;

-- 第二步：查询这些用户的所有订单
SELECT * FROM orders WHERE user_id IN (1,2,3,4);

-- 第三步：查询这些订单的所有订单项
SELECT * FROM order_items WHERE order_id IN (100,101,102,103);
```

**关联层级**：
- 假设模型关系如下：
  ```go
  type User struct {
    Orders []Order // 用户有多个订单（has many）
  }

  type Order struct {
    OrderItems []OrderItem // 订单有多个订单项（has many）
  }
  ```

**关键点**：
- 使用 `.` 语法实现多级嵌套预加载（`Orders.OrderItems`）
- 一次性批量加载所有关联数据，避免逐层查询的 N+1 性能问题

---

### 2. 带条件的嵌套预加载
```go
db.Preload("Orders", "state = ?", "paid").
  Preload("Orders.OrderItems").
  Find(&users)
```
**作用**：查询所有用户，但仅预加载 **状态为 paid 的订单**，且这些订单的 **所有订单项**。

**生成的 SQL**：
```sql
-- 第一步：查询所有用户
SELECT * FROM users;

-- 第二步：查询这些用户的 paid 状态订单
SELECT * FROM orders 
WHERE user_id IN (1,2,3,4) 
  AND state = 'paid'; -- 关键过滤条件

-- 第三步：查询这些订单的订单项
SELECT * FROM order_items 
WHERE order_id IN (200,201); -- 仅过滤后的订单ID
```

**代码结构解析**：
| 代码片段                          | 作用                                                                 |
|-----------------------------------|--------------------------------------------------------------------|
| `Preload("Orders", "state = ?", "paid")` | 1️⃣ 加载用户的订单，但添加 `state = 'paid'` 过滤条件 |
| `Preload("Orders.OrderItems")`           | 2️⃣ 加载这些过滤后订单的订单项 |

**关键点**：
- **条件作用域**：第一个 `Preload` 的条件仅作用于 `Orders` 层级
- **链式预加载**：第二个 `Preload` 基于已过滤的 `Orders` 继续加载其子关联
- **执行顺序**：
    1. 先过滤出符合条件的订单（paid 状态）
    2. 再加载这些订单的订单项

---

### 对比两种查询结果
| 用户 | 原始查询 (`Orders.OrderItems`) | 带条件查询 (`Orders(state=paid).OrderItems`) |
|------|--------------------------------|---------------------------------------------|
| 用户A | 所有订单 + 所有订单项          | 仅 paid 订单 + 这些订单的订单项              |
| 用户B | 所有订单 + 所有订单项          | 仅 paid 订单 + 这些订单的订单项              |

---

### 扩展场景：给子关联添加条件
如果需要对订单项也添加过滤（例如只加载数量大于 5 的订单项）：
```go
db.Preload("Orders", "state = ?", "paid").
  Preload("Orders.OrderItems", "quantity > ?", 5).
  Find(&users)
```
**生成的 SQL**：
```sql
-- 订单项查询会变成：
SELECT * FROM order_items 
WHERE order_id IN (200,201) 
  AND quantity > 5; -- 新增子级条件
```

---

### 最佳实践建议
1. **性能优化**：
    - 通过 `Preload` 批量加载关联数据，避免循环查询
    - 预估数据量，过度预加载可能一次性加载过多无用数据

2. **条件优先级**：
   ```go
   // 结构体标签 > 链式Preload条件 > 全局auto_preload
   type Order struct {
     Items []OrderItem `gorm:"PRELOAD:false"` // 优先级最高
   }
   ```

3. **调试技巧**：
   ```go
   db.Debug().Preload(...).Find(...) // 查看实际执行的SQL
   ```

这种多级预加载模式非常适合电商类系统（如用户→订单→商品）、社交系统（用户→帖子→评论）等需要深度关联查询的场景。


# 自定义预加载 SQL


您可以通过传入`func（db * gorm.DB）* gorm.DB`来自定义预加载SQL，例如：

```go
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("orders.amount DESC")
}).Find(&users)
//// SELECT * FROM users;
//// SELECT * FROM orders WHERE user_id IN (1,2,3,4) order by orders.amount DESC;
```