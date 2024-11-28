package stacks

import (
	"alexchomiak/go-docker-api/db"
	"alexchomiak/go-docker-api/types"

	"github.com/gofiber/fiber/v2"
)

func GetStacks(c *fiber.Ctx) error {
	result := []types.ComposeStack{}
	db.Gorm.Raw("SELECT * FROM \"compose_stacks\"").Scan(&result)
	return c.JSON(result)
}
