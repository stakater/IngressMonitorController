package util

import (
	"sort"
	"strconv"
	"strings"
)

func SliceAtoi(stringSlice []string) ([]int, error) {
	var intSlice = []int{}

	for _, stringValue := range stringSlice {
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return intSlice, err
		}
		intSlice = append(intSlice, intValue)
	}

	return intSlice, nil
}

func SliceAtoi64(stringSlice []string) ([]int64, error) {
	var intSlice = []int64{}

	for _, stringValue := range stringSlice {
		intValue, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return intSlice, err
		}
		intSlice = append(intSlice, intValue)
	}

	return intSlice, nil
}

func SliceItoa(intSlice []int) []string {
	var stringSlice = []string{}

	for _, intValue := range intSlice {
		stringValue := strconv.Itoa(intValue)
		stringSlice = append(stringSlice, stringValue)
	}

	return stringSlice
}

func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func SplitAndSort(s string, sep string) []string {

	slice := strings.Split(s, sep)
	sort.Strings(slice)
	return slice
}
