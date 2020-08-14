package main

import (
	"fmt"

	commands "./commands"
)

//Funcion Main
func main() {
	fmt.Println("Prueba de creacion de un disco ----------")
	commands.MKDisk("disc_2.dsk", 18, 0)
	commands.ReadFile("disc_2.dsk")
}
