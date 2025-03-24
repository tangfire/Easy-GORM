# 自动迁移

使用 migrate 来维持你的表结构一直处于最新状态。

警告：migrate 仅支持创建表、增加表中没有的字段和索引。为了保护你的数据，它并不支持改变已有的字段类型或删除未被使用的字段

```go
db.AutoMigrate(&User{})

db.AutoMigrate(&User{}, &Product{}, &Order{})

// 创建表的时候，添加表后缀
db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&User{})
```

# 其他数据库迁移工具

GORM 的数据库迁移工具能够支持主要的数据库，但是如果你要寻找更多的迁移工具， GORM 会提供的数据库接口，这可能可以给到你帮助。

```go
// 返回 `*sql.DB`
db.DB()
```

# 表结构的方法

## Has Table

```go
// 检查模型中 User 表是否存在
db.HasTable(&User{})

// 检查 users 表是否存在
db.HasTable("users")
```

## Create Table

```go
// 通过模型 User 创建表
db.CreateTable(&User{})

// 在创建 users 表的时候，会在 SQL 语句中拼接上 `"ENGINE=InnoDB"`
db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&User{})
```

## Drop table

```go
// 删除模型 User 表
db.DropTable(&User{})

// 删除 users 表
db.DropTable("users")

// 删除模型 User 表和 products 表
db.DropTableIfExists(&User{}, "products")
```

# ModifyColumn

以给定的值来定义字段类型

```go
// User 模型，改变 description 字段的数据类型为 `text`
db.Model(&User{}).ModifyColumn("description", "text")
```



# DropColumn

```go
// User 模型，删除  description 字段
db.Model(&User{}).DropColumn("description")
```


# Add Indexes


```go
// 为 `name` 字段建立一个名叫 `idx_user_name` 的索引
db.Model(&User{}).AddIndex("idx_user_name", "name")

// 为 `name`, `age` 字段建立一个名叫 `idx_user_name_age` 的索引
db.Model(&User{}).AddIndex("idx_user_name_age", "name", "age")

// 添加一条唯一索引
db.Model(&User{}).AddUniqueIndex("idx_user_name", "name")

// 为多个字段添加唯一索引
db.Model(&User{}).AddUniqueIndex("idx_user_name_age", "name", "age")
```


# Remove Index

```go
// 移除索引
db.Model(&User{}).RemoveIndex("idx_user_name")
```

# Add Foreign Key

```go
// 添加主键
// 第一个参数 : 主键的字段
// 第二个参数 : 目标表的 ID
// 第三个参数 : ONDELETE
// 第四个参数 : ONUPDATE
db.Model(&User{}).AddForeignKey("city_id", "cities(id)", "RESTRICT", "RESTRICT")
```


这段代码用于在 GORM 中为数据库表添加外键约束，具体作用是为 `User` 模型（对应数据库表 `users`）的 `city_id` 字段添加一个指向 `cities` 表主键 `id` 的外键约束。以下是逐部分解析：

---

### 一、代码结构解析
```go
db.Model(&User{}).AddForeignKey("city_id", "cities(id)", "RESTRICT", "RESTRICT")
```
- **`AddForeignKey` 方法**：GORM 提供的用于显式声明外键约束的方法。
- **参数说明**：
    1. **`"city_id"`**：当前表（`users`）的外键字段名，表示该字段将关联到其他表。
    2. **`"cities(id)"`**：目标表及其主键字段，此处为 `cities` 表的 `id` 列。
    3. **`"RESTRICT"`**：删除父表记录时的约束行为（例如，若 `cities` 表的某条记录被 `users` 表引用，则删除该记录会被拒绝）。
    4. **`"RESTRICT"`**：更新父表记录时的约束行为（例如，若修改 `cities.id` 的值，而该值被 `users.city_id` 引用，则更新会被拒绝）。

---

### 二、外键约束的作用
1. **数据完整性**  
   确保 `users.city_id` 的值必须存在于 `cities.id` 中，防止无效的关联数据插入。
2. **约束行为**
    - **`RESTRICT`**：阻止对父表的删除或更新操作，如果存在关联的子表记录。
    - 其他可选行为（未在此代码中使用）：
        - **`CASCADE`**：父表记录删除/更新时，自动删除/更新子表关联记录。
        - **`SET NULL`**：父表记录删除/更新时，将子表外键字段设为 `NULL`。
        - **`NO ACTION`**：与 `RESTRICT` 类似，但某些数据库的实现可能不同。

---

### 三、适用场景与注意事项
1. **显式声明外键**  
   当模型结构未通过 GORM 标签隐式定义外键时（例如未使用 `gorm:"foreignKey:CityID"`），可以通过此方法手动添加。
2. **数据库兼容性**  
   不同数据库（如 MySQL、SQLite）对外键的支持程度可能不同，需确保数据库引擎支持外键（例如 SQLite 默认启用外键需添加 `?_foreign_keys=1` 参数）。
3. **迁移操作**  
   通常在 `AutoMigrate` 后调用此方法，以确保表结构正确更新。

---

### 四、对比隐式外键定义
若通过模型标签隐式定义外键，代码可能如下：
```go
type User struct {
    gorm.Model
    CityID uint `gorm:"foreignKey:CityID;references:ID"`
}

type City struct {
    gorm.Model
    Users []User
}
```
此时 GORM 会自动生成外键约束，无需手动调用 `AddForeignKey`。

---

### 总结
此代码通过显式声明外键约束，强制保证 `users` 表与 `cities` 表之间的关联完整性，并通过 `RESTRICT` 策略防止误删/改父表数据。若需更灵活的行为（如级联删除），可调整第三、第四个参数。


# Remove ForeignKey

```go
db.Model(&User{}).RemoveForeignKey("city_id", "cities(id)")
```

这段代码的作用是**移除 `User` 表（模型）中 `city_id` 字段与 `cities` 表 `id` 字段之间的外键约束**。以下是详细解析：

---

### 一、代码结构
```go
db.Model(&User{}).RemoveForeignKey("city_id", "cities(id)")
```
- **`RemoveForeignKey` 方法**：GORM 提供的用于删除外键约束的方法。
- **参数说明**：
    1. **`"city_id"`**：当前表（`users`）中定义的外键字段名。
    2. **`"cities(id)"`**：目标表及其主键字段（即外键关联的目标）。

---

### 二、适用场景
1. **修改表结构**  
   当需要调整表之间的关联关系时（例如从外键关联改为逻辑关联），需先删除原有的物理外键约束。
2. **规避约束限制**  
   若外键约束导致某些操作受限（如删除父表记录），可能需要临时或永久移除外键以绕过限制。
3. **迁移兼容性**  
   某些数据库（如 SQLite）对复杂外键支持有限，移除后可能更易于跨数据库兼容。

---

### 三、注意事项
1. **外键名称的依赖性**  
   不同数据库生成外键名称的规则可能不同。例如，MySQL 会生成唯一的外键名（如 `fk_users_city_id`），而 PostgreSQL 可能依赖声明顺序。若未显式命名外键，直接通过字段和目标表删除可能在某些场景下失效。
2. **数据库支持**  
   需确保数据库引擎支持外键操作（例如 SQLite 默认需启用 `_foreign_keys=1` 参数）。
3. **数据完整性风险**  
   删除外键后，`users.city_id` 的值可能不再受 `cities.id` 的约束，需通过应用层逻辑保证数据有效性。

---

### 四、对比添加外键的代码
在添加外键时，需指定约束行为（如 `RESTRICT`）：
```go
// 添加外键（参数包含 ON DELETE/UPDATE 规则）
db.Model(&User{}).AddForeignKey("city_id", "cities(id)", "RESTRICT", "RESTRICT")
```
而删除外键仅需标识外键字段与目标表，无需指定约束规则。

---

### 五、操作建议
1. **结合迁移使用**  
   建议在数据库迁移脚本中成对使用 `AddForeignKey` 和 `RemoveForeignKey`，确保版本可控。
2. **检查外键存在性**  
   删除前可通过 `db.Dialect().HasForeignKey()` 检查外键是否存在（需结合具体数据库实现）。

---

### 总结
此代码通过 `RemoveForeignKey` 移除了 `users` 表与 `cities` 表之间的外键关联，适用于需要解除物理约束的场景。操作时需注意数据库兼容性及潜在的数据完整性风险。





