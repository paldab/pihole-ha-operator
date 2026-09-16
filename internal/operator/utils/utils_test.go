package utils_test

import (
	"maps"
	"testing"

	"github.com/paldab/pihole-ha-operator/internal/operator/utils"
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
		t.Fatal("MergeMap function should not have changed anything in the newMap but changes happened")
	}

	var overrideMap = make(map[string]string)
	newMap := utils.MergeMap(baseMap, overrideMap)
	if !maps.Equal(baseMap, newMap) {
		t.Fatal("MergeMap function should not have changed anything in the newMap but changes happened")
	}
}

const emptyStringChecksum = "12ae32cb1ec02d01eda3581b127c1fee3b0dc53572ed6baf239721a03d82e126"

func TestChecksum_EmptyString(t *testing.T) {
	val, err := utils.CalculateChecksum("")

	if err != nil {
		t.Fatalf("CalculateChecksum shoould not give an error but received one: %v", err)
	}

	if val != emptyStringChecksum {
		t.Fatalf("CalculateChecksum value should be empty but instead received: %s", val)
	}
}

func TestChecksum_StringContent(t *testing.T) {
	const testValue = "hellothisis a test $1! 123123 $ 2 @@@@"

	val, err := utils.CalculateChecksum(testValue)

	if err != nil {
		t.Fatalf("CalculateChecksum shoould not give an error but received one: %v", err)
	}

	if val == emptyStringChecksum {
		t.Fatalf("CalculateChecksum value should be something else than the default empty string value but instead received: %s", val)
	}
}

func TestChecksum_NumberContent(t *testing.T) {
	const testValue = 912312

	val, err := utils.CalculateChecksum(testValue)

	if err != nil {
		t.Fatalf("CalculateChecksum shoould not give an error but received one: %v", err)
	}

	if val == emptyStringChecksum {
		t.Fatalf("CalculateChecksum value should be something else than the default empty string value but instead received: %s", val)
	}
}
