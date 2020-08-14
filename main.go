package main

import (
	"fmt"

	commands "./commands"
)

//Funcion Main
func main() {
	fmt.Println("Prueba de creacion de un disco ----------")
	commands.MKDisk("disc_3.dsk", 25, 1)
	commands.ReadFile("disc_3.dsk")
	commands.RMDisk("disc_2.dsk")
}
