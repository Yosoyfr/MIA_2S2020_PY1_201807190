package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
	interpreterF(readMIAFile("/home/yosoyfr/MIA/test_discos/mkfs.mia"))
	//commands.Mkfs("vda1", "fast")
	fmt.Println("-----------------------")
	//commands.Mkdir("vda1", "/home/yosoyfr/Descargas", true)
	//commands.Mkdir("vda1", "/home/yosoyfr/Escritorio", true)
	//commands.Mkdir("vda1", "/bin", true)
	//commands.Mkdir("vda1", "/etc/usr", true)
	commands.Mkdir("vda1", "/snap", false)
	commands.Reports("vda1", "sb", "/home/yosoyfr/MIA/test_discos/report.pdf")
}

//Funcionalidad del interprete
func interpreterF(input string) {
	interpreter.CommandChecker(interpreter.ScanInput(input))
}

func commandsTest() {
	commands.FKDisk("disc_2.dsk", 200, 'B', 'L', 'W', "LOGICA6")
	datos, err := ioutil.ReadFile("disc_2.dsk")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(datos)
}

//Funcion para leer los archivos con extension ".mia"
func readMIAFile(route string) string {
	var output string
	file, err := os.Open(route)
	if err != nil {
		fmt.Println("Error: El sistema no puede encontrar el archivo especificado.")
		return output
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}
	return output
}
