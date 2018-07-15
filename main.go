package main

import (
	"flag"
	"fmt"
	// use when drawing progress bars!
	//"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"encoding/json"
)

type SFConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	LoginURL string `json:"loginUrl"`
	InstanceName string `json:"instanceId"`
}

var actionMap = map[string]struct{}{
	"insert":{},
	"update":{},
	"delete":{},
	"upsert":{},
}

type Command struct {
	Action string
	Files []string
}

type Process struct {
	Commands []Command

	// if false, serial
	IsParallel bool
}

// when not parallel, this returns an optimized mapping of actions to files. for instance, if the user enters the
// command:
// data insert file1 file2 update file3 insert file4
// this command will create a mapping where file1, file2, and file4 are all under "insert"
func (p Process) CommandMap() map[string][]string {
	m := make(map[string][]string)

	for _, cmd := range p.Commands{
		for _, file := range cmd.Files {
			m[cmd.Action] = append(m[cmd.Action], file)
		}
	}

	return m
}

// example usage
// data --config config.json insert test1.csv test2.json update test3.xml

func main() {
	//configFlag := flag.String("c", "", "Path to config file")
	syncFlag := flag.Bool("s", false, "If true, does data load in sync mode")

	flag.Parse()

	//config, err := loadConfigFile(*configFlag)

	//if err != nil {
	//	log.Fatalln(err)
	//}

	cmds := parseArgs(flag.Args())

	process := Process{cmds, *syncFlag}

	//fmt.Println(config)
	fmt.Println(process)
}

func parseArgs(args []string) []Command {
	cmds := make([]Command, 0)

	for len(args) > 0 {
		cmd := Command{}


		if isCommand(args[0]) {
			cmd.Action = args[0]

			i := 1

			for i < len(args) && !isCommand(args[i]) {
				cmd.Files = append(cmd.Files, args[i])
				i++
			}

			args = args[i:]
		} else {
			cmd.Files = append(cmd.Files, args[0])
			args = args[1:]
		}

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
