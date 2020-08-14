package commands

import (
	"fmt"
	"log"
	"os"
)

//Funcion para eliminar un archivo que represente un disco duro
func RMDisk(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Desea eliminar este disco duro?")
	fmt.Println("0 - Concelar \n 1 - Confirmar")
	var input int
	fmt.Scanln(&input)
	if input == 1 {
		fmt.Println("Disco Eliminado con exito!!")
	} else {
		fmt.Println("Cancelado")
	}

}
