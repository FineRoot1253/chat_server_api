package middleware

import (


	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func StatusInList(status int, statusList []int) bool {
	for _, i := range statusList {
		if i == status {
			return true
		}
	}
	return false
}

func GetTransactionMiddleWare(db *gorm.DB) fiber.Handler{
	return func(c *fiber.Ctx) error {
		tx:= db.Begin()
		defer func (){
			if r := recover(); r!= nil {
				tx.Rollback()
			}
		}()

		c.Locals("TX",tx)
		
		c.Next()

		return nil
	}
}