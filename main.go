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
	interpreterF(interpreter.ReadMIAFile("/home/yosoyfr/MIA/discos/input.mia"))
	//Prueba de leer consola
	//readConsole()
	//commands.FDisk("/home/yosoyfr/MIA/discos/Disco1.dsk", 1, 'M', 'L', 'W', "Logica4")
	commands.FDiskDelete("/home/yosoyfr/MIA/discos/Disco1.dsk", true, "Extendida")
	//commands.Mount("/home/yosoyfr/MIA/discos/Disco1.dsk", "Particion1")
	commands.Reports("vda1", "MBR", "/home/yosoyfr/MIA/discos/mbr2.pdf", "")
	commands.Reports("vda1", "DISK", "/home/yosoyfr/MIA/discos/disk2.pdf", "")
	commands.Reports("vda1", "BITACORA", "/home/yosoyfr/MIA/discos/log.pdf", "")
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
	//commands.Mount("/home/yosoyfr/MIA/discos/Disco1.dsk", "Particion1")
	//commands.Mkfile("vda1", "/bin/t6.txt", true, 0, "test")
	//commands.Mkfile("vda1", "/home/user/docs/Hola.txt", false, 0, "Hola")
	//commands.Reports("vda1", "DIRECTORIO", "/home/yosoyfr/MIA/discos/directorio.pdf", "")
	//commands.Reports("vda1", "SB", "/home/yosoyfr/MIA/discos/sb.pdf", "")
	//commands.Reports("vda1", "TREE_FILE", "/home/yosoyfr/MIA/discos/tree_file.pdf", "/cartasuicidio.txt")
	//commands.Reports("vda1", "TREE_DIRECTORIO", "/home/yosoyfr/MIA/discos/tree_dir.pdf", "/")
	//commands.Reports("vda1", "TREE_COMPLETE", "/home/yosoyfr/MIA/discos/tree_complete.pdf", "")
	//commands.Reports("vda1", "DISK", "/home/yosoyfr/MIA/discos/disk.pdf", "")
	//commands.Reports("vda1", "MBR", "/home/yosoyfr/MIA/discos/mbr.pdf", "")
	commands.FDisk("disc_2.dsk", 200, 'B', 'L', 'W', "LOGICA6")
	datos, err := ioutil.ReadFile("disc_2.dsk")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(datos)
}
