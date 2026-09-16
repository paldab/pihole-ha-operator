// Package utils contain utility functions
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"slices"
)

func ValueOrDefault[T any](baseValue, defaultValue *T) *T {
	if baseValue == nil {
		return defaultValue
	}

	return baseValue
}

func MergeMap(baseMap, overrideMap map[string]string) map[string]string {
	if baseMap == nil {
		baseMap = make(map[string]string)
	}

	if overrideMap == nil {
		return baseMap
	}

	if len(overrideMap) == 0 {
		return baseMap
	}

	maps.Copy(baseMap, overrideMap)
	return baseMap
}

// CalculateChecksum returns the checksum of values given.
// Returns an empty string as default
func CalculateChecksum[V any](values V) (string, error) {
	data, err := json.Marshal(values)

	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func GetSortedKeysFromMap[K ~string, T any](records map[K]T) []string {
	keys := make([]string, 0, len(records))

	for key := range records {
		keys = append(keys, string(key))
	}

	slices.Sort(keys)

	return keys
}
