# 自动创建/更新


GORM 将在创建或保存一条记录的时候自动保存关联和它的引用，如果关联有一个主键， GORM 将调用 `Update` 来更新它， 不然，它将会被创建。

```go
user := User{
    Name:            "jinzhu",
    BillingAddress:  Address{Address1: "Billing Address - Address 1"},
    ShippingAddress: Address{Address1: "Shipping Address - Address 1"},
    Emails:          []Email{
        {Email: "jinzhu@example.com"},
        {Email: "jinzhu-2@example@example.com"},
    },
    Languages:       []Language{
        {Name: "ZH"},
        {Name: "EN"},
    },
}

db.Create(&user)
//// BEGIN TRANSACTION;
//// INSERT INTO "addresses" (address1) VALUES ("Billing Address - Address 1");
//// INSERT INTO "addresses" (address1) VALUES ("Shipping Address - Address 1");
//// INSERT INTO "users" (name,billing_address_id,shipping_address_id) VALUES ("jinzhu", 1, 2);
//// INSERT INTO "emails" (user_id,email) VALUES (111, "jinzhu@example.com");
//// INSERT INTO "emails" (user_id,email) VALUES (111, "jinzhu-2@example.com");
//// INSERT INTO "languages" ("name") VALUES ('ZH');
//// INSERT INTO user_languages ("user_id","language_id") VALUES (111, 1);
//// INSERT INTO "languages" ("name") VALUES ('EN');
//// INSERT INTO user_languages ("user_id","language_id") VALUES (111, 2);
//// COMMIT;

db.Save(&user)
```

这段代码展示了使用 GORM（Go 的 ORM 库）进行复杂关联关系数据操作的完整流程。我们可以分两个阶段来解析：

---

### 一、`db.Create(&user)` 阶段（事务性创建）
代码通过 `db.Create(&user)` 创建用户及其所有关联数据，GORM 会自动开启事务并执行以下操作：

1. **地址表插入**
    - `BillingAddress` 和 `ShippingAddress` 作为用户的一对一关联
   ```sql
   INSERT INTO addresses (address1) VALUES ("Billing Address - Address 1");
   INSERT INTO addresses (address1) VALUES ("Shipping Address - Address 1");
   ```
   这两个地址会被分别插入并生成 ID（假设为 1 和 2）

2. **用户表插入**
   ```sql
   INSERT INTO users (name,billing_address_id,shipping_address_id) VALUES ("jinzhu", 1, 2);
   ```
   用户记录会携带两个地址的外键（`billing_address_id` 和 `shipping_address_id`）

3. **邮箱表插入（一对多）**
   ```sql
   INSERT INTO emails (user_id,email) VALUES (111, "jinzhu@example.com");
   INSERT INTO emails (user_id,email) VALUES (111, "jinzhu-2@example.com");
   ```
   通过 `Emails` 切片字段自动创建多条关联记录（`user_id` 外键指向用户表）

4. **语言表插入（多对多）**
   ```sql
   INSERT INTO languages ("name") VALUES ('ZH');
   INSERT INTO user_languages ("user_id","language_id") VALUES (111, 1);
   INSERT INTO languages ("name") VALUES ('EN');
   INSERT INTO user_languages ("user_id","language_id") VALUES (111, 2);
   ```
   这里涉及两个表的操作：
    - `languages` 表插入语言记录
    - `user_languages` 中间表插入关联记录（自动维护多对多关系）

---

### 二、`db.Save(&user)` 阶段（更新操作）
`Save` 方法在 GORM 中用于保存所有字段（包含零值），其行为特点：

1. **主键存在时的更新逻辑**
    - 如果 `user` 对象已存在主键（如 ID=111），则会执行 `UPDATE` 操作
    - 如果关联数据（如地址、邮箱）发生变化，GORM 会根据关联配置自动更新或创建新记录

2. **关联数据的级联处理**
    - 若修改了 `BillingAddress` 或 `ShippingAddress` 的字段，会触发地址表的更新
    - 若 `Emails` 或 `Languages` 切片内容变化（如增删元素），会对应更新关联表

3. **事务完整性保证**
    - 与 `Create` 类似，所有更新操作会被包裹在事务中以保证原子性

---

### 关键特性说明
1. **关联关系配置**
    - 用户与地址：通过 `has one` 或 `belongs to` 配置
    - 用户与邮箱：通过 `has many` 配置
    - 用户与语言：通过 `many2many` 配置（自动维护中间表）

2. **事务自动管理**
   GORM 默认开启事务，确保所有关联操作的原子性。如果任何一步操作失败，整个事务会回滚

3. **外键自动处理**
    - 地址的 `ID` 会自动写入用户表的外键字段
    - 邮箱的 `user_id` 会自动关联用户主键
    - 中间表 `user_languages` 会同时记录用户和语言的关联

---

### 典型应用场景
这种模式特别适合需要保持数据一致性的复杂业务对象创建，例如：
- 用户注册时需要同时创建资料、地址、偏好设置
- 订单创建时需要同时处理商品、支付、物流信息
- 医疗系统中患者档案与检查报告、用药记录的关联存储

通过 GORM 的关联关系配置，开发者可以避免手动处理复杂的 SQL 拼接和外键管理，显著提升开发效率和数据安全性。

# 关闭自动更新


如果你的关联记录已经存在在数据库中， 你可能会不想去更新它。

你可以设置 `gorm:association_autoupdate` 为 `false`

```go
// 不更新有主键的关联，但会更新引用
db.Set("gorm:association_autoupdate", false).Create(&user)
db.Set("gorm:association_autoupdate", false).Save(&user)
```

或者使用 GORM 的标签， `gorm:"association_autoupdate:false"`

```go
type User struct {
  gorm.Model
  Name       string
  CompanyID  uint
  // 不更新有主键的关联，但会更新引用
  Company    Company `gorm:"association_autoupdate:false"`
}
```

# 关闭自动创建




