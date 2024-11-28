package env

import "os"

var StacksDir = "./stacks" // Directory to store stack YAML files

func init() {
	if value, exists := os.LookupEnv("STACK_DIR"); exists {
		StacksDir = value
	}
}
