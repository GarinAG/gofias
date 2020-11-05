package util

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

func ConvertFloat32SliceToInterface(slice []float32) []interface{} {
	newSlice := make([]interface{}, len(slice))
	for i := range slice {
		newSlice[i] = slice[i]
	}

	return newSlice
}
