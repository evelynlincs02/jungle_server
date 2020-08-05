package utils

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

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

func RandomInt(n int) int { // [0,n)
	return rand.Intn(n)
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length uint) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// https://gist.github.com/j33ty/79e8b736141be19687f565ea4c6f4226
//  Alloc is bytes of allocated heap objects.
//  TotalAlloc is cumulative bytes allocated for heap objects.
//  Sys is the total bytes of memory obtained from the OS.
//  NumGC is the number of completed GC cycles.
func MemUsageString() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return fmt.Sprintf("Alloc=%v KiB\tTotalAlloc = %v KiB\tSys = %v KiB\tNumGC = %v\n",
		bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func bToKb(b uint64) uint64 {
	return b / 1024
}
