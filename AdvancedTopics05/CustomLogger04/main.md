# Logger

Gorm 建立了对 Logger 的支持，默认模式只会在错误发生的时候打印日志。

```go
// 开启 Logger, 以展示详细的日志
db.LogMode(true)

// 关闭 Logger, 不再展示任何日志，即使是错误日志
db.LogMode(false)

// 对某个操作展示详细的日志，用来排查该操作的问题
db.Debug().Where("name = ?", "jinzhu").First(&User{})
```

# 自定义 Logger

参考 GORM 的默认 logger 是怎么自定义的 https://github.com/jinzhu/gorm/blob/master/logger.go

例如，使用 [Revel](https://revel.github.io/)  的 Logger 作为 GORM 的输出

```go
db.SetLogger(gorm.Logger{revel.TRACE})
```


使用`os.Stdout`作为输出

```go
db.SetLogger(log.New(os.Stdout, "\r\n", 0))
```

