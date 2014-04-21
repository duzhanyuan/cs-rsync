package mapsort

import (
	"sort"
)

type StringInterfaceSorter []StringInterface

type StringInterface struct {
	Key string
	Val interface{}
}

func newStringInterfaceSorter(m map[string]interface{}) StringInterfaceSorter {
	sIS := make(StringInterfaceSorter, 0, len(m))

	for k, v := range m {
		sIS = append(sIS, StringInterface{k, v})
	}

	return sIS
}

func (sIS StringInterfaceSorter) Len() int {
	return len(sIS)
}

func (sIS StringInterfaceSorter) Less(i, j int) bool {
	// return sIS[i].Val < sIS[j].Val // 按值排序
	return sIS[i].Key < sIS[j].Key // 按键排序
}

func (sIS StringInterfaceSorter) Swap(i, j int) {
	sIS[i], sIS[j] = sIS[j], sIS[i]
}

func Sort(m map[string]interface{}) StringInterfaceSorter {
	stringInterfaceSort := newStringInterfaceSorter(m)
	sort.Sort(stringInterfaceSort)
	return stringInterfaceSort
}
