package tool

import (
	"math/rand"
	"slices"
)

// ##########  数组  ############
func InArray[T comparable](Ele T, Array []T) bool {
	return slices.Contains(Array, Ele)
	// for _, v := range Array {
	// 	if v == Ele {
	// 		return true
	// 	}
	// }
	// return false
}

// 数组去重
func UniqueArray[T comparable](arr []T) []T {
	seen := make(map[T]bool, len(arr))
	uniqueArr := make([]T, 0, len(arr))
	for _, v := range arr {
		if !seen[v] {
			seen[v] = true
			uniqueArr = append(uniqueArr, v)
		}
	}
	return uniqueArr
}

// 合并数组
func MergeArray[T comparable](dest []T, src []T) (result []T) {
	result = make([]T, len(dest)+len(src))
	copy(result, dest)            //将第一个数组传入result
	copy(result[len(dest):], src) //将第二个数组接在尾部，也就是 len(dest):
	return
}

/**
 * 过滤数组
 * @param arr 数组
 * @param shouldKeep 过滤条件
 * @returns 过滤后的数组
 * @example
 * filteredInt := FilterArray([]int{1, 0, 2}, func(v int) bool { return v != 0 })
 * filteredStr := FilterArray([]string{"a", "", "b"}, func(v string) bool { return v != "" })
 */
func FilterArray[T any](arr []T, shouldKeep func(T) bool) []T {
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if shouldKeep(v) {
			result = append(result, v)
		}
	}
	return result
}

// 打乱数组

func ShuffleArray[T any](arr []T) []T {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
	// for i := len(arr) - 1; i > 0; i-- {
	// 	j := rand.Intn(i + 1)
	// 	arr[i], arr[j] = arr[j], arr[i]
	// }
	// return arr
}

// 反转数组
func ReverseArray[T any](arr []T) {
	slices.Reverse(arr)
	// for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
	// 	arr[i], arr[j] = arr[j], arr[i] // 交换
	// }
}
