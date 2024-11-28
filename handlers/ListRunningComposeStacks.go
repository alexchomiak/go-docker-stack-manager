package handlers

import (
	"alexchomiak/go-docker-api/db"
	"context"
	"errors"
	"log"

	servertypes "alexchomiak/go-docker-api/types"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"gorm.io/gorm"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
)

// ListRunningComposeStacks lists all running Docker Compose stacks with detailed container info
func ListRunningComposeStacks(c *fiber.Ctx) error {
	// Create Docker client
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error creating Docker client: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize Docker client",
		})
	}

	// List all running containers
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list running containers",
		})
	}

	// Organize containers by stack
	stacks := make(map[string]fiber.Map)

	for _, container := range containers {
		// Extract the stack name from the container labels
		stackName := container.Labels["com.docker.compose.project"]
		if stackName == "" {
			stackName = "unknown"
		}

		// Inspect container for detailed info
		inspectData, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Printf("Error inspecting container %s: %v", container.ID, err)
			continue
		}

		// Add full container data from inspect
		containerDetails := fiber.Map{
			"id":   container.ID[:12],
			"name": container.Names[0],
			"info": inspectData, // Full inspect data
		}

		// Add container to the stack
		if _, exists := stacks[stackName]; !exists {
			stacks[stackName] = fiber.Map{
				"stack":      stackName,
				"containers": []fiber.Map{},
				"networks":   []string{}, // Populate later from network settings
			}
		}

		stack := stacks[stackName]
		stack["containers"] = append(stack["containers"].([]fiber.Map), containerDetails)
		stacks[stackName] = stack
	}

	// Fetch network and metadata for each stack
	for stackName, stack := range stacks {
		networks := getStackNetworks(cli, stackName)
		stack["networks"] = networks
		stacks[stackName] = stack
	}

	// Convert to a slice for JSON response
	result := make([]fiber.Map, 0, len(stacks))
	for _, stack := range stacks {
		stackId := stack["stack"].(string)
		record := servertypes.ComposeStack{StackID: stackId}
		query := db.Gorm.First(&record)
		if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
			continue
		}
		if query.Error != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error Filtering Stacks from Database",
			})
		}
		stack["record"] = record
		result = append(result, stack)
	}

	return c.JSON(result)
}

// Fetch stack networks using Docker client
func getStackNetworks(cli *client.Client, stackName string) []string {
	networkFilters := filters.NewArgs()
	networkFilters.Add("label", "com.docker.compose.project="+stackName)

	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: networkFilters,
	})
	if err != nil {
		log.Printf("Error fetching networks for stack '%s': %v", stackName, err)
		return []string{}
	}

	networkNames := make([]string, len(networks))
	for i, network := range networks {
		networkNames[i] = network.Name
	}
	return networkNames
}
