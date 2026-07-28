package utils_test

import (
	"maps"
	"testing"

	"github.com/paldab/pihole-ha-operator/internal/utils"
)

func TestMergeMap_OverrideValue(t *testing.T) {
	baseMap := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	}

	overrideMap := map[string]string{
		"c": "z",
	}

	newMap := utils.MergeMap(baseMap, overrideMap)

	if newMap["c"] != "z" {
		t.Fatalf("expected 'z' got %s", newMap["c"])
	}

	if newMap["a"] != "a" {
		t.Fatal("unoverride values have been overridden")
	}

}

func TestMergeMap_EmptyOverride(t *testing.T) {
	baseMap := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	}

	var overrideMapNil map[string]string = nil

	newMapNil := utils.MergeMap(baseMap, overrideMapNil)

	if !maps.Equal(baseMap, newMapNil) {
		t.Fatal("MergeMap function should not have changed anything in the newMap but changes happend")
	}

	var overrideMap = make(map[string]string)
	newMap := utils.MergeMap(baseMap, overrideMap)
	if !maps.Equal(baseMap, newMap) {
		t.Fatal("MergeMap function should not have changed anything in the newMap but changes happend")
	}
}
