package stacks

import (
	"alexchomiak/go-docker-api/db"
	"alexchomiak/go-docker-api/env"
	"alexchomiak/go-docker-api/types"
	"alexchomiak/go-docker-api/utility"
	"errors"
	"os"
	"os/exec"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type StartComposeStackRequestBody struct {
	Stack string `json:"stackId"`
}

// CreateComposeCluster deploys a Docker Compose stack and stores the reference
func StartComposeStack(c *fiber.Ctx) error {
	// Parse the YAML file from the request body
	body := new(StartComposeStackRequestBody)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}

	// * Find Stack from Stack DB
	stack := types.ComposeStack{StackID: body.Stack}
	result := db.Gorm.First(&stack)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"errors": "Stack not found",
		})
	}

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "Error querying for stack",
		})
	}

	file, err := utility.CreateTempFile(env.StacksDir, "docker-compose-*.yaml", []byte(stack.ComposeFileContents))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "Error creating temporary file for Stack Provisioning",
		})
	}
	defer os.Remove(file.Name())

	cmd := exec.Command("docker-compose", "-f", file.Name(), "-p", stack.StackID, "up", "-d")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to tear down stack",
			"details": string(output),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stack Started successfully",
	})
}
