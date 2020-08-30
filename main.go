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
	//interpreterF()
	//commands.Reports("Hoja1_carnet.dsk", "disc", "png", "")
	commands.Mount("Hoja1_carnet.dsk", "Logic1")
	commands.Mount("Hoja2_carnet.dsk", "Part4")
	commands.Mount("Hoja2_carnet.dsk", "Logic1")
	commands.Mount("Hoja1_carnet.dsk", "Logic1")
	commands.Mount("Hoja1_carnet.dsk", "Part1")
	commands.Mount("Hoja2_carnet.dsk", "Part1")
	commands.Mount("Hoja2_carnet.dsk", "Part3")
	commands.Mount("Hoja2_carnet.dsk", "Part")
	commands.ShowMountedDisks()
	commands.Unmount("vda1")
	commands.Unmount("vda2")
	commands.Unmount("vdb3")
	commands.ShowMountedDisks()
	commands.Mount("Hoja1_carnet.dsk", "Logic1")
	commands.Mount("Hoja1_carnet.dsk", "Part1")
	commands.Mount("Hoja2_carnet.dsk", "Part1")
	commands.ShowMountedDisks()
}

//Funcionalidad del interprete
func interpreterF() {
	fmt.Println("Prueba del interpreter ----------")
	input := readMIAFile("ht.mia")
	interpreter.CommandChecker(interpreter.ScanInput(input))
}

func commandsTest() {
	//commands.ReadFile("disc_3.dsk")
	//commands.MKDisk("disc_2.dsk", 5, 'K')
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
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}
	return output
}
