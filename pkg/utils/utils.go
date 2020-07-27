package utils

// type T interface {
// 	IndexOf(T)
// }

func FindInt(a []int, target int) int {
	for i, v := range a {
		if v == target {
			return i
		}
	}
	return len(a)
}

func FindString(a []string, target string) int {
	for i, v := range a {
		if v == target {
			return i
		}
	}
	return len(a)
}

func RemoveIdx(a []int, i int) []int {
	// NOTE: 此方法較省，但不保有原來順序
	a[i] = a[len(a)-1]
	a = a[:len(a)-1]
	return a
}

func Unique(s []string) []string {
	keys := make(map[string]struct{})
	for i := range s {
		keys[s[i]] = struct{}{}
	}

	uni := make([]string, 0, len(keys))
	for k := range keys {
		uni = append(uni, k)
	}

	return uni
}

// func GetMapKey(m map[string]interface{}) []string {
// 	keys := make([]string, 0, len(m))
// 	for k, _ := range m {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }
