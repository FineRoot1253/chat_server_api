package presenter

import "github.com/gofiber/fiber/v2"

func Success(data interface{}, description string) *fiber.Map {
	return &fiber.Map{
		"code": 1,
		"data": data,
		"msg":  description,
	}
}

func Failure(description string) *fiber.Map {
	return &fiber.Map{
		"code": -1,
		"data": nil,
		"msg":  description,
	}
}
