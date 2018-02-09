package config

import (
	"fmt"
	"strings"

	"github.com/cheekybits/genny/generic"
)

// GenericItem is a generic implementation
type GenericItem generic.Type

// GenericItemList represents an array of GenericItem
type GenericItemList []GenericItem

// IGenericItem returns TerragruntExtensioner from the supplied type
func IGenericItem(item interface{}) TerragruntExtensioner {
	return item.(TerragruntExtensioner)
}

func (list GenericItemList) init(config *TerragruntConfigFile) {
	for i := range list {
		IGenericItem(&list[i]).init(config)
	}
}

// Merge elements from an imported list to the current list priorising those already existing
func (list *GenericItemList) merge(imported GenericItemList, mode mergeMode, argName string) {
	if len(imported) == 0 {
		return
	} else if len(*list) == 0 {
		*list = imported
		return
	}

	log := IGenericItem(&(*list)[0]).logger().Debugf

	// Create a map with existing elements
	index := make(map[string]int, len(*list))
	for i, item := range *list {
		index[IGenericItem(&item).id()] = i
	}

	// Create a list of the hooks that should be added to the list
	new := make(GenericItemList, 0, len(imported))
	for _, item := range imported {
		name := IGenericItem(&item).id()
		if pos, exist := index[name]; exist {
			// It already exist in the list, so is is an override, we remove it from its current position
			// and add it to the list of newly addd elements to keep its original declaration ordering.
			new = append(new, (*list)[pos])
			delete(index, name)
			log("Skipping %s %v as it is overridden in the current config", argName, name)
		} else {
			new = append(new, item)
		}
	}

	if len(index) != len(*list) {
		// Some elements must bre removed from the original list, we must
		newList := make(GenericItemList, 0, len(index))
		for _, item := range *list {
			name := IGenericItem(&item).id()
			if _, found := index[name]; found {
				newList = append(newList, item)
			}
		}
		*list = newList
	}

	if mode == mergeModeAppend {
		*list = append(*list, new...)
	} else {
		*list = append(new, *list...)
	}
}

// Help returns the information relative to the elements within the list
func (list GenericItemList) Help(listOnly bool, lookups ...string) (result string) {
	list.sort()
	add := func(item TerragruntExtensioner, name string) {
		extra := item.extraInfo()
		if extra != "" {
			extra = " " + extra
		}
		result += fmt.Sprintf("\n%s%s%s\n%s", TitleID(item.id()), name, extra, item.help())
	}

	var table [][]string
	width := []int{30, 0, 0}

	if listOnly {
		addLine := func(values ...string) {
			table = append(table, values)
			for i, value := range values {
				if len(value) > width[i] {
					width[i] = len(value)
				}
			}
		}
		add = func(item TerragruntExtensioner, name string) {
			addLine(TitleID(item.id()), name, item.extraInfo())
		}
	}

	for _, item := range list.Enabled() {
		item := IGenericItem(&item)
		match := len(lookups) == 0
		for i := 0; !match && i < len(lookups); i++ {
			match = strings.Contains(item.name(), lookups[i]) || strings.Contains(item.id(), lookups[i]) || strings.Contains(item.extraInfo(), lookups[i])
		}
		if !match {
			continue
		}
		var name string
		if item.id() != item.name() {
			name = " " + item.name()
		}
		add(item, name)
	}

	if listOnly {
		for i := range table {
			result += fmt.Sprintln()
			for j := range table[i] {
				result += fmt.Sprintf("%-*s", width[j]+1, table[i][j])
			}
		}
	}

	return
}

// Enabled returns only the enabled items on the list
func (list GenericItemList) Enabled() GenericItemList {
	result := make(GenericItemList, 0, len(list))
	for _, item := range list {
		iItem := IGenericItem(&item)
		if iItem.enabled() {
			iItem.normalize()
			result = append(result, item)
		}
	}
	return result
}

// Run execute the content of the list
func (list GenericItemList) Run(args ...interface{}) (result []interface{}, err error) {
	if len(list) == 0 {
		return
	}

	for _, item := range list {
		iItem := IGenericItem(&item)
		var temp interface{}
		iItem.logger().Infof("Running %s (%s): %s", iItem.itemType(), iItem.id(), iItem.name())
		iItem.normalize()
		if temp, err = iItem.run(args...); err != nil {
			err = fmt.Errorf("Error while executing %s(%s): %v", iItem.itemType(), iItem.id(), err)
			return
		}
		result = append(result, temp)
	}
	return
}
