package main

import (
	"flag"
	"fmt"
	// use when drawing progress bars!
	//"gopkg.in/cheggaaa/pb.v1"
	"encoding/json"
	"io/ioutil"
)

type SFConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LoginURL     string `json:"loginUrl"`
	InstanceName string `json:"instanceId"`
}

var actionMap = map[string]struct{}{
	"insert": {},
	"update": {},
	"delete": {},
	"upsert": {},
}

type Command struct {
	Action string
	Files  []string
}

type Process struct {
	Commands []Command

	// if false, serial
	IsParallel bool
}

func (p Process) OptimizeCommands() []Command {
	m := p.CommandMap()
	cmds := make([]Command, 0)

	for action, files := range m {
		cmds = append(cmds, Command{action, files})
	}

	return cmds
}

// when not parallel, this returns an optimized mapping of actions to files. for instance, if the user enters the
// command:
// data insert file1 file2 update file3 insert file4
// this command will create a mapping where file1, file2, and file4 are all under "insert"
func (p Process) CommandMap() map[string][]string {
	m := make(map[string][]string)

	for _, cmd := range p.Commands {
		for _, file := range cmd.Files {
			m[cmd.Action] = append(m[cmd.Action], file)
		}
	}

	return m
}

type Job struct {

}

// example usage
// data --config config.json insert test1.csv test2.json update test3.xml

func main() {
	syncFlag := flag.Bool("s", false, "If true, does data load in sync mode")

	flag.Parse()

	cmds := parseArgs(flag.Args())

	process := Process{cmds, *syncFlag}

	fmt.Println(process)
	fmt.Println(process.CommandMap())
}

func parseArgs(args []string) []Command {
	// TODO handle invalid syntax
	cmds := make([]Command, 0)

	for len(args) > 0 {
		cmd := Command{}

		cmd.Action = args[0]

		i := 1

		for i < len(args) && !isCommand(args[i]) {
			cmd.Files = append(cmd.Files, args[i])
			i++
		}

		args = args[i:]
		cmds = append(cmds, cmd)
	}

	return cmds
}

func loadConfigFile(path string) (config SFConfig, err error) {
	file, err := ioutil.ReadFile(path)

	if err != nil {
		return SFConfig{}, err
	}

	err = json.Unmarshal(file, &config)

	return
}

func isCommand(str string) bool {
	_, ok := actionMap[str]
	return ok
}
