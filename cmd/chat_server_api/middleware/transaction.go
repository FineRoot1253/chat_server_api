package middleware

import (
	"log"
	"net/http"

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

func GetTransactionMiddleWare(db gorm.DB) fiber.Handler{
	return func(c *fiber.Ctx) error {
		tx:= db.Begin()
		defer func (){
			if r := recover(); r!= nil {
				tx.Rollback()
			}
		}()

		c.Locals("TX",tx)
		
		c.Next()

		if StatusInList(c.Context().Response.StatusCode(), []int{http.StatusOK, http.StatusCreated}) {
			log.Print("committing transactions")
			if err := tx.Commit().Error; err != nil {
				log.Print("trx commit error: ", err)
			}
		} else {
			log.Print("rolling back transaction due to status code: ", c.Context().Response.StatusCode())
			tx.Rollback()
		}

		return nil
	}
}