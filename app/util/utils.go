package util

import "sort"

// Проверить наличие строки в массиве
func ContainsString(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}

	return false
}

// Удалить дубликаты строк в массиве
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

// Конвертировать массив строк в массив итерфейсов
func ConvertStringSliceToInterface(slice []string) []interface{} {
	newSlice := make([]interface{}, len(slice))
	for i := range slice {
		newSlice[i] = slice[i]
	}

	return newSlice
}

// Сортировать массив строк по длине строки, от большего к меньшему
func SortStringSliceByLength(slice []string) {
	sort.Slice(slice, func(i, j int) bool {
		return len(slice[i]) > len(slice[j])
	})
}
