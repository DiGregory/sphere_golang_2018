package main

import (
	"strconv"
	"math"
	"sort"
)

/*
	сюда вам надо писать функции, которых не хватает, чтобы проходили тесты в gotchas_test.go

	IntSliceToString
	MergeSlices
	GetMapValuesSortedByKey
*/
func IntSliceToString(slice []int) string {
	NewString := ""
	for _, v := range slice {
		NewString = NewString + strconv.Itoa(v)
	}
	return NewString
}

func MergeSlices(floatslice []float32, intslice []int32) []int {
	newslice := []int{}
	for _, v := range floatslice {
		newslice = append(newslice, int(math.Round(float64(v))))
	}
	for _,v:=range intslice{
		newslice=append(newslice,int(v))
	}
	return newslice
}
func GetMapValuesSortedByKey(InputMap map[int]string)[]string {
	var sortedmap=[]string{}


	keys := make([]int, 0, len(InputMap))
	for k := range InputMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		sortedmap=append(sortedmap,InputMap[k])

	}


	return sortedmap
}

