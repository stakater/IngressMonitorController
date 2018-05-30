package main

import "strconv"

func sliceAtoi(stringSlice []string) ([]int, error) {
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
