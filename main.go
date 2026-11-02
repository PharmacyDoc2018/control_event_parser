package main

import (
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	p := initProcess()
	defer p.db.Close()

	p.cliConfig = cli.InitConfig()
	p.setupCommands()

	for {
		line, err := p.cliConfig.Rl.Readline()
		if err != nil {
			break
		}

		err = p.cliConfig.CommandExe(line)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println()
	}
}
