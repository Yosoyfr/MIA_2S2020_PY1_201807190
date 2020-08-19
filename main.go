package main

import (
	"bufio"
	"fmt"
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
	/*
		fmt.Println("Prueba de creacion de un disco ----------")
		commands.MKDisk("disc_3.dsk", 25, 1)
		commands.ReadFile("disc_3.dsk")
		commands.RMDisk("disc_2.dsk")
	*/
	commands.ReadFile("disc_3.dsk")
	fmt.Println("Prueba del interpreter ----------")
	input := readMIAFile("input.mia")
	interpreter.CommandChecker(interpreter.ScanInput(input))
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
