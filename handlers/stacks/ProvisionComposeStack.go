package stacks

import (
	"alexchomiak/go-docker-api/db"
	"alexchomiak/go-docker-api/env"
	"alexchomiak/go-docker-api/types"
	"alexchomiak/go-docker-api/utility"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ProvisionComposeStack(c *fiber.Ctx) error {
	yamlData := c.Body()
	yamlKey, err := utility.NormalizeAndHashYaml(yamlData)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing & processing YAML file",
		})
	}

	if len(yamlData) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No YAML file provided",
		})
	}

	// * Query for existing stack with the same hash
	indexQuery := types.ComposeStack{ComposeFileHash: yamlKey}
	result := db.Gorm.First(&indexQuery)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error Querying Compose Index",
		})
	}

	record_exists := true
	if err == nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
		record_exists = false
	}

	if record_exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Stack already exists against defined YAML Hash. POST to this API to Start Stack",
		})
	}

	tempFile, err := utility.CreateTempFile(env.StacksDir, "docker-compose-*.yaml", yamlData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer os.Remove(tempFile.Name())

	// * Create New Stack Metadata Obj
	composeStack := types.ComposeStack{
		StackID:             uuid.New().String(),
		ComposeFileContents: string(yamlData),
		ComposeFileHash:     yamlKey,
	}

	// Deploy the stack using docker-compose
	cmd := exec.Command("docker-compose", "-f", tempFile.Name(), "-p", composeStack.StackID, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to deploy stack",
			"details": string(output),
		})
	}

	result = db.Gorm.Create(&composeStack)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failure storing stack in Database. YAML Stack May already exist.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stack provisioned successfully",
		"stackId": composeStack.StackID,
	})
}

// Helper function to extract stack name from the YAML content
func extractStackName(yamlData []byte) string {
	lines := strings.Split(string(yamlData), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
