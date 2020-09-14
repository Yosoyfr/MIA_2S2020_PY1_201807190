package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	/*
		Imports para los comandos de consola
	*/

	/*
		imports para el interprete
	*/
	"./interpreter"
)

//Funcion Main
func main() {
	//interpreterF(interpreter.ReadMIAFile("/home/yosoyfr/MIA/discos/input.mia"))
	readConsole()
	//commands.Mount("/home/yosoyfr/MIA/discos/Disco1.dsk", "Particion1")
	//commands.Mv("vda1", "/boot", "/f")
}

func readConsole() {
	reader := bufio.NewReader(os.Stdin)
	text := ""
	for !strings.EqualFold(text, "EXIT\n") {
		fmt.Print("MIA_201807190 -> ")
		text, _ = reader.ReadString('\n')
		text = strings.Replace(text, "\r", "", -1)
		interpreterF(text)
	}
	fmt.Println("[EXIT]")
}

//Funcionalidad del interprete
func interpreterF(input string) {
	interpreter.CommandChecker(interpreter.ScanInput(input))
}

