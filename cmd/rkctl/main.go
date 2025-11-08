package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "0.0.1"

func main() {
	showVersion := flag.Bool("version", false, "показать версию и выйти")
	flag.Parse()

	if *showVersion {
		fmt.Println("rkctl", version)
		return
	}
	fmt.Fprintln(os.Stdout, "rkctl: скелет CLI. Команды появятся позже.")
}
