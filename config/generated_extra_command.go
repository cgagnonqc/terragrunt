// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package config

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/shell"
)

// ExtraCommandList represents an array of ExtraCommand
type ExtraCommandList []ExtraCommand

// IExtraCommand returns TerragruntExtensioner from the supplied type
func IExtraCommand(item interface{}) TerragruntExtensioner {
	return item.(TerragruntExtensioner)
}

func (list ExtraCommandList) init(config *TerragruntConfigFile) {
	for i := range list {
		IExtraCommand(&list[i]).init(config)
	}
}

// Merge elements from an imported list to the current list priorising those already existing
func (list *ExtraCommandList) merge(imported ExtraCommandList, mode mergeMode, argName string) {
	if len(imported) == 0 {
		return
	} else if len(*list) == 0 {
		*list = imported
		return
	}

	log := IExtraCommand(&(*list)[0]).logger().Debugf

	// Create a map with existing elements
	index := make(map[string]int, len(*list))
	for i, item := range *list {
		index[IExtraCommand(&item).id()] = i
	}

	// Create a list of the hooks that should be added to the list
	newList := make(ExtraCommandList, 0, len(imported))
	for _, item := range imported {
		name := IExtraCommand(&item).id()
		if pos, exist := index[name]; exist {
			// It already exist in the list, so is is an override
			// We remove it from its current position and add it to the list of newly added elements to keep its original declaration ordering.
			newList = append(newList, (*list)[pos])
			delete(index, name)
			log("Skipping %s %v as it is overridden in the current config", argName, name)
			continue
		}
		newList = append(newList, item)
	}

	if len(index) != len(*list) {
		// Some elements must be removed from the original list, we simply regenerate the list
		// including only elements that are still in the index.
		newList := make(ExtraCommandList, 0, len(index))
		for _, item := range *list {
			name := IExtraCommand(&item).id()
			if _, found := index[name]; found {
				newList = append(newList, item)
			}
		}
		*list = newList
	}

	if mode == mergeModeAppend {
		*list = append(*list, newList...)
	} else {
		*list = append(newList, *list...)
	}
}

// Help returns the information relative to the elements within the list
func (list ExtraCommandList) Help(listOnly bool, lookups ...string) (result string) {
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
		item := IExtraCommand(&item)
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
func (list ExtraCommandList) Enabled() ExtraCommandList {
	result := make(ExtraCommandList, 0, len(list))
	for _, item := range list {
		iItem := IExtraCommand(&item)
		if iItem.enabled() {
			iItem.normalize()
			result = append(result, item)
		}
	}
	return result
}

// Run execute the content of the list
func (list ExtraCommandList) Run(status error, args ...interface{}) (result []interface{}, err error) {
	if len(list) == 0 {
		return
	}

	list.sort()

	var (
		errs       errorArray
		errOccured bool
	)
	for _, item := range list {
		iItem := IExtraCommand(&item)
		if (status != nil || errOccured) && !iItem.ignoreError() {
			continue
		}
		iItem.logger().Infof("Running %s (%s): %s", iItem.itemType(), iItem.id(), iItem.name())
		iItem.normalize()
		temp, currentErr := iItem.run(args...)
		currentErr = shell.FilterPlanError(currentErr, iItem.options().TerraformCliArgs[0])
		if currentErr != nil {
			if _, ok := currentErr.(errors.PlanWithChanges); ok {
				errs = append(errs, currentErr)
			} else {
				errOccured = true
				errs = append(errs, fmt.Errorf("Error while executing %s(%s): %v", iItem.itemType(), iItem.id(), currentErr))
			}
		}
		iItem.setState(currentErr)
		result = append(result, temp)
	}
	switch len(errs) {
	case 0:
		break
	case 1:
		err = errs[0]
	default:
		err = errs
	}
	return
}
