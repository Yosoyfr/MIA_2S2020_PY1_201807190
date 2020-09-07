package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	/*
		Imports para los comandos de consola
	*/
	commands "./commands"
	/*
		imports para el interprete
	*/
	"./interpreter"
)

//Funcion Main
func main() {
	//interpreterF(readMIAFile("/home/yosoyfr/MIA/test_discos/input.mia"))
	//interpreterF(readMIAFile("/home/yosoyfr/MIA/test_discos/mkfs.mia"))
	//interpreterF("mkfs -id->vda1 -type->fast" + "\n")
	//interpreterF(interpreter.ReadMIAFile("input.mia"))
	/*
		commands.Mkfs("vda1", "fast")
		fmt.Println("-----------------------")
		commands.Mkdir("vda1", "/home/yosoyfr/Descargas", true)
		//commands.Mkdir("vda1", "/home/yosoyfr/Escritorio", true)
		commands.Mkdir("vda1", "/media", true)
		commands.Mkdir("vda1", "/log", true)
		commands.Mkdir("vda1", "/bin", true)
		commands.Mkdir("vda1", "/opt", true)
		commands.Mkdir("vda1", "/proc", true)
		commands.Mkdir("vda1", "/etc/usr", true)
		commands.Mkdir("vda1", "/dev", false)
		commands.Reports("vda1", "directorio", "/home/yosoyfr/MIA/test_discos/directorio.pdf")
		commands.Reports("vda1", "sb", "/home/yosoyfr/MIA/test_discos/report.pdf")
	*/

	//Prueba de leer consola
	//readConsole()
	commands.Reports("vda1", "BM_ARBDIR", "/home/yosoyfr/MIA/test_discos/report.pdf", "")
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

//Funcion para el test de comandos
func commandsTest() {
	commands.FKDisk("disc_2.dsk", 200, 'B', 'L', 'W', "LOGICA6")
	datos, err := ioutil.ReadFile("disc_2.dsk")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(datos)
}
