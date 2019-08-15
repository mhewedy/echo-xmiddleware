package main

import (
	"echo-middleware/xmiddleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4/middleware"
	"github.com/qor/audited"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo/v4"
)

type Product struct {
	gorm.Model
	audited.AuditedModel
	code string
}

func main() {
	e := echo.New()
	db := initDatabase()
	defer db.Close()

	// JWT middleware
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Skipper: func(c echo.Context) bool {
			if c.Request().URL.String() == "/get-token" {
				return true
			}
			return false
		},
		SigningKey: []byte("secret"),
	}))
	// GormAudit middleware, depends on registration of middleware.JWT
	e.Use(xmiddleware.GormAudit(db))

	// APIs
	e.GET("/get-token", createToken)
	// HasRole middleware, depends on registration of middleware.JWT
	e.POST("/create-product", createProduct, xmiddleware.HasRole("admin"))

	// start
	e.Logger.Fatal(e.Start(":1323"))
}

func initDatabase() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	audited.RegisterCallbacks(db)
	db.AutoMigrate(&Product{})
	return db
}

func createToken(c echo.Context) error {
	newJwt := jwt.New(jwt.SigningMethodHS256)
	claims := newJwt.Claims.(jwt.MapClaims)
	claims["id"] = 101
	claims["exp"] = time.Now().Add(time.Hour * 24 * 10).Unix()
	claims["roles"] = [...]string{"admin"}
	token, _ := newJwt.SignedString([]byte("secret"))
	return c.JSON(http.StatusOK, map[string]string{"id_token": token})
}

var createProduct = func(c echo.Context) error {
	db := xmiddleware.GetGormDB(c)
	product := Product{code: "my-prod-code"}
	db.Create(&product)
	return c.JSON(http.StatusOK, &product)
}
