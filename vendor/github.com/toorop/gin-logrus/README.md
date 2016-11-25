# gin-logrus

[Logrus](https://github.com/Sirupsen/logrus) logger middleware for [Gin](https://gin-gonic.github.io/gin/)

![golang gin middleware logrus logger](http://i.imgur.com/P140Vi0.png)

## Usage
```go
import (
  "github.com/Sirupsen/logrus"
  "github.com/toorop/gin-logrus"
  "github.com/gin-gonic/gin"

log := logrus.New()
// hooks, config,...

r := gin.New()
r.Use(ginlogrus.Logger(log), gin.Recovery())

// pingpong
r.GET("/ping", func(c *gin.Context) {
	c.Data(200, "text/plain", []byte("pong"))
})

r.Run("127.0.0.1:8080")
```