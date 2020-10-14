package util

import "sort"

func ContainsString(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}

	return false
}

func UniqueStringSlice(slice []string) []string {
	keys := make(map[string]bool)
	var list []string

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

func ConvertStringSliceToInterface(slice []string) []interface{} {
	newSlice := make([]interface{}, len(slice))
	for i := range slice {
		newSlice[i] = slice[i]
	}

	return newSlice
}

func SortStringSliceByLength(slice []string) {
	sort.Slice(slice, func(i, j int) bool {
		return len(slice[i]) > len(slice[j])
	})
}

func RemoveStringsDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	var result []string

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}
