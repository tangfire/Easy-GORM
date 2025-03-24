# 链式操作

Gorm 继承了链式操作接口， 所以你可以写像下面一样的代码：

```go
db, err := gorm.Open("postgres", "user=gorm dbname=gorm sslmode=disable")

// 创建一个新的关系
tx := db.Where("name = ?", "jinzhu")

// 新增更多的筛选条件
if someCondition {
    tx = tx.Where("age = ?", 20)
} else {
    tx = tx.Where("age = ?", 30)
}

if yetAnotherCondition {
    tx = tx.Where("active = ?", 1)
}
```

直到调用立即方法之前都不会产生查询，在某些场景中会很有用。

就像你可以封装一个包来处理一些常见的逻辑


# 创建方法

创建方法就是那些会产生 SQL 查询并且发送到数据库，通常它就是一些 CRUD 方法， 就像:

`Create`, `First`, `Find`, `Take`, `Save`, `UpdateXXX`, `Delete`, `Scan`, `Row`, `Rows`…

下面是一个创建方法的例子:

```go
tx.Find(&user)
```

生成

```go
SELECT * FROM users where name = 'jinzhu' AND age = 30 AND active = 1;
```

# Scopes方法

Scope 方法基于链式操作理论创建的。

使用它，你可以提取一些通用逻辑，写一些更可用的库。

```go
func AmountGreaterThan1000(db *gorm.DB) *gorm.DB {
    return db.Where("amount > ?", 1000)
}

func PaidWithCreditCard(db *gorm.DB) *gorm.DB {
    return db.Where("pay_mode_sign = ?", "C")
}

func PaidWithCod(db *gorm.DB) *gorm.DB {
    return db.Where("pay_mode_sign = ?", "C")
}

func OrderStatus(status []string) func (db *gorm.DB) *gorm.DB {
    return func (db *gorm.DB) *gorm.DB {
        return db.Scopes(AmountGreaterThan1000).Where("status in (?)", status)
    }
}

db.Scopes(AmountGreaterThan1000, PaidWithCreditCard).Find(&orders)
// 查找所有大于1000的信用卡订单和金额

db.Scopes(AmountGreaterThan1000, PaidWithCod).Find(&orders)
// 查找所有大于1000的 COD 订单和金额

db.Scopes(AmountGreaterThan1000, OrderStatus([]string{"paid", "shipped"})).Find(&orders)
// 查找大于1000的所有付费和运单
```

# 多个创建方法


当使用 GORM 的创建方法，后面的创建方法将复用前面的创建方法的搜索条件（不包含内联条件）

```go
db.Where("name LIKE ?", "jinzhu%").Find(&users, "id IN (?)", []int{1, 2, 3}).Count(&count)
```

生成

```go
SELECT * FROM users WHERE name LIKE 'jinzhu%' AND id IN (1, 2, 3)

SELECT count(*) FROM users WHERE name LIKE 'jinzhu%'
```

# 线程安全

所有的链式操作都将会克隆并创建一个新的数据库对象（共享一个连接池），GORM 对于多个 goroutines 的并发使用是安全的。





