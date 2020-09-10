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
	//interpreterF(interpreter.ReadMIAFile("/home/yosoyfr/MIA/test_discos/input.mia"))
	//interpreterF(interpreter.ReadMIAFile("/home/yosoyfr/MIA/discos/input.mia"))
	//Prueba de leer consola
	//readConsole()
	//inp := "Francisco Luis Suarez Lopez, Mara Isabel Lopez Garcia, Heather Gabriela Paz Lopez, Amanda Garcia de Lopez, Luis Antonio Suarez Roldan, Francisco Luis Lopez Smith, Luis Rolando Lopez Garcia, Amanda Argentina Lopez Garcia, Mario Rene Lopez Garcia, Carlos Luis Mendez Lopez, Mario Samuel Lopez Aldana, Peggy Lily Lopez Aldana"
	commands.Mount("/home/yosoyfr/MIA/discos/Disco1.dsk", "Particion1")
	//commands.Mkfile("vda1", "/home/user/docs/caca.txt", false, 0, "puta")
	commands.Reports("vda1", "DIRECTORIO", "/home/yosoyfr/MIA/discos/directorio.pdf", "")
	commands.ReportDirectoryTree("vda1", "/bin")
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
