package config

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gruntwork-io/terragrunt/shell"
	"github.com/gruntwork-io/terragrunt/util"
	logging "github.com/op/go-logging"
)

// ExtraCommand is a definition of user extra command that should be executed in place of terraform
type ExtraCommand struct {
	TerragruntExtensionBase `hcl:",squash"`

	Commands   []string `hcl:"commands"`
	Aliases    []string `hcl:"aliases"`
	Arguments  []string `hcl:"arguments"`
	ExpandArgs *bool    `hcl:"expand_args"`
	UseState   *bool    `hcl:"use_state"`
	ActAs      string   `hcl:"act_as"`
	VersionArg string   `hcl:"version"`
}

func (item ExtraCommand) extraInfo() string {
	return fmt.Sprintf("[%s]", strings.Join(util.RemoveElementFromList(item.list(), item.Name), ", "))
}

func (item ExtraCommand) help() (result string) {
	if item.Description != "" {
		result += fmt.Sprintf("\n%s\n", item.Description)
	}

	if item.OS != nil {
		result += fmt.Sprintf("\nApplied only on the following OS: %s\n", strings.Join(item.OS, ", "))
	}

	if item.Arguments != nil {
		result += fmt.Sprintf("\nAutomatically added argument(s): %s\n", strings.Join(item.Arguments, ", "))
	}

	return result
}

func (item *ExtraCommand) normalize() {
	if item.Commands == nil {
		// There is no list of commands, so we consider the name to be the extra command
		item.Commands = []string{item.Name}
	} else if validName.MatchString(item.Name) && !util.ListContainsElement(item.list(), item.Name) {
		// The name is considered as an alias if it match name criteria
		item.Aliases = append(item.Aliases, item.Name)
	}

	def := func(value bool) *bool { return &value }
	if item.UseState == nil {
		item.UseState = def(false)
	}
	if item.ExpandArgs == nil {
		item.ExpandArgs = def(true)
	}
}

var validName = regexp.MustCompile(`^[\w\.-]+$`)

func (item *ExtraCommand) list() []string {
	result := make([]string, 0, len(item.Commands)+len(item.Aliases))
	result = append(result, item.Commands...)
	result = append(result, item.Aliases...)

	for i := range result {
		result[i] = strings.TrimSpace(strings.Split(result[i], "=")[0])
	}
	result = util.RemoveDuplicatesFromListKeepFirst(result)
	sort.Strings(result)
	return result
}

func (item *ExtraCommand) resolve(cmd string) *ActualCommand {
	cmd = item.resolveAlias(cmd)
	if !util.ListContainsElement(item.Commands, cmd) {
		return nil
	}

	var behaveAs string

	if item.ActAs != "" {
		// The command must act as another command for extra argument validation
		item.options().TerraformCliArgs[0] = item.ActAs
	} else {
		item.ActAs = cmd
		if item.UseState == nil || *item.UseState {
			// We simulate that the extra command acts as the plan command to init the state file
			// and get the modules
			behaveAs = "plan"
		}
	}

	return &ActualCommand{cmd, behaveAs, item}
}

func (item *ExtraCommand) resolveAlias(cmd string) string {
	options := item.options()

	for _, alias := range item.Aliases {
		split := strings.SplitN(alias, "=", 2)
		if cmd != split[0] {
			continue
		}

		if len(split) == 1 {
			return item.Commands[0]
		}

		cmd = split[1]
		if strings.ContainsAny(split[1], " |,&$") {
			cmd = "bash"

			var args string
			for _, arg := range append(item.Arguments, options.TerraformCliArgs[1:]...) {
				if !strings.Contains(arg, " \t") {
					args += " " + arg
				} else {
					args += fmt.Sprintf(` "%s"`, arg)
				}
			}

			script := split[1]
			if strings.Contains(script, " $*") {
				script = strings.Replace(script, " $*", args, -1)
			} else if !strings.Contains(script, "|") {
				script += args
			}

			item.Arguments = []string{"-c", script}
			options.TerraformCliArgs = options.TerraformCliArgs[:1]
		}
	}
	return cmd
}

// ----------------------- ExtraCommandList -----------------------

//go:generate genny -in=extension_base_list.go -out=generated_extra_command.go gen "GenericItem=ExtraCommand"
func (list *ExtraCommandList) argName() string { return "extra_command" }

func (list ExtraCommandList) sort() ExtraCommandList {
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Merge elements from an imported list to the current list
func (list *ExtraCommandList) Merge(imported ExtraCommandList) {
	list.merge(imported, mergeModeAppend, list.argName())
}

// GetVersions returns the the list of versions for extra commands that have a version available
func (list ExtraCommandList) GetVersions() string {
	var result string
	for _, item := range list.Enabled() {
		if item.VersionArg == "" {
			continue
		}

		if strings.Contains(item.Name, " ") {
			result += fmt.Sprintf("\n%s\n", item.Name)
		}
		for _, cmd := range item.Commands {
			logLevel := logging.GetLevel("")
			if logLevel == logging.NOTICE {
				logging.SetLevel(logging.WARNING, "")
			}
			os.Setenv("TERRAGRUNT_COMMAND", cmd)
			args := []string{item.VersionArg}
			if strings.ContainsAny(item.VersionArg, " |,&$") {
				cmd = "bash"
				args = util.ExpandArguments([]string{"-c", item.VersionArg}, item.options().WorkingDir)
			}
			out, err := shell.RunShellCommandAndCaptureOutput(item.options(), false, cmd, args...)
			logging.SetLevel(logLevel, "")
			if err != nil {
				item.logger().Infof("Got %s %s while getting version for %s", color.RedString(err.Error()), out, cmd)
			} else {
				result += fmt.Sprintln(strings.TrimSpace(out))
			}
		}
	}
	return result
}

// ActualCommand returns
func (list ExtraCommandList) ActualCommand(cmd string) ActualCommand {
	for _, item := range list.Enabled() {
		if match := item.resolve(cmd); match != nil {
			return *match
		}
	}
	return ActualCommand{cmd, "", nil}
}

// ActualCommand represents the command that should be executed
type ActualCommand struct {
	Command  string
	BehaveAs string
	Extra    *ExtraCommand
}