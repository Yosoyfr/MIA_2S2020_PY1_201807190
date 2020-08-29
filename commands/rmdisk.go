package commands

import (
	"fmt"
	"os"
)

//Funcion para eliminar un archivo que represente un disco duro
func RMDisk(path string) {
	fmt.Println("Desea eliminar este disco duro?")
	fmt.Println("0 - Cancelar\n1 - Confirmar")
	var input int
	fmt.Scanln(&input)
	if input == 1 {
		err := os.Remove(path)
		if err != nil {
			fmt.Println("Error: El sistema no puede encontrar el archivo especificado.")
			return
		}
		fmt.Println("Disco Eliminado con exito!!")
	} else {
		fmt.Println("Cancelado")
	}

}
