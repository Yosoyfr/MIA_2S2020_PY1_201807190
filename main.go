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
	//inp := "Carta de suicidio por si no sale sale archivos"
	commands.Mount("/home/yosoyfr/MIA/discos/Disco1.dsk", "Particion1")
	//commands.Mkfile("vda1", "/cartasuicidio.txt", false, 0, inp)
	//commands.Reports("vda1", "DIRECTORIO", "/home/yosoyfr/MIA/discos/directorio.pdf", "")
	//commands.Reports("vda1", "SB", "/home/yosoyfr/MIA/discos/sb.pdf", "")
	commands.Reports("vda1", "TREE_FILE", "/home/yosoyfr/MIA/discos/tree_file.pdf", "/cartasuicidio.txt")
	commands.Reports("vda1", "TREE_DIRECTORIO", "/home/yosoyfr/MIA/discos/tree_dir.pdf", "/")
	//commands.Reports("vda1", "DISK", "/home/yosoyfr/MIA/discos/disk.pdf", "")
	//commands.Reports("vda1", "MBR", "/home/yosoyfr/MIA/discos/mbr.pdf", "")
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
