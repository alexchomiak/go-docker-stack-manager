package utility

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"gopkg.in/yaml.v3"
)

func NormalizeAndHashYaml(yamlData []byte) (string, error) {
	// Parse the YAML into a generic map
	var parsedData map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &parsedData); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Normalize the map (sort keys recursively)
	normalizedData := normalizeMap(parsedData)

	// Convert the normalized data to JSON (to ensure consistent serialization)
	jsonData, err := json.Marshal(normalizedData)
	if err != nil {
		return "", fmt.Errorf("failed to convert normalized data to JSON: %w", err)
	}

	// Generate a SHA256 hash of the JSON
	hash := sha256.Sum256(jsonData)

	// Return the hash as a hex string
	return fmt.Sprintf("%x", hash), nil
}

// normalizeMap recursively sorts map keys to ensure consistent ordering
func normalizeMap(input map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	keys := make([]string, 0, len(input))

	// Collect and sort the keys
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Re-insert keys in sorted order
	for _, k := range keys {
		val := input[k]
		// Recursively normalize nested maps
		if nestedMap, ok := val.(map[string]interface{}); ok {
			normalized[k] = normalizeMap(nestedMap)
		} else if nestedList, ok := val.([]interface{}); ok {
			normalized[k] = normalizeList(nestedList)
		} else {
			normalized[k] = val
		}
	}

	return normalized
}

// normalizeList recursively normalizes elements of a list
func normalizeList(input []interface{}) []interface{} {
	// Sort primitive lists to ensure consistent order
	if len(input) > 0 {
		// Check if all elements are of the same type (for sorting)
		if reflect.TypeOf(input[0]).Kind() == reflect.String {
			sort.SliceStable(input, func(i, j int) bool {
				return input[i].(string) < input[j].(string)
			})
		} else if reflect.TypeOf(input[0]).Kind() == reflect.Int {
			sort.SliceStable(input, func(i, j int) bool {
				return input[i].(int) < input[j].(int)
			})
		} else if reflect.TypeOf(input[0]).Kind() == reflect.Float64 {
			sort.SliceStable(input, func(i, j int) bool {
				return input[i].(float64) < input[j].(float64)
			})
		}
	}

	// Recursively normalize maps in the list
	for i, val := range input {
		if nestedMap, ok := val.(map[string]interface{}); ok {
			input[i] = normalizeMap(nestedMap)
		} else if nestedList, ok := val.([]interface{}); ok {
			input[i] = normalizeList(nestedList)
		}
	}

	return input
}
