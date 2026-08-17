// Package utils contain utility functions
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"sort"
)

func ValueOrDefault[T any](baseValue, defaultValue *T) *T {
	if baseValue == nil {
		return defaultValue
	}

	return baseValue
}

func MergeMap(baseMap, overrideMap map[string]string) map[string]string {
	if baseMap == nil {
		return make(map[string]string)
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

	sort.Strings(keys)

	return keys
}
