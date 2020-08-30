package commands

import "fmt"

//Funcion encargada de desmontar discos de la lista
func Unmount(id string) {
	findPart := false
	//Recorremos la lista de discos montados
	for i, disk := range mountedDisks {
		//Recorremos la lista de particiones hasta encontrar la del id
		for j, part := range disk.parts {
			if part.id == id {
				mountedDisks[i].parts = append(disk.parts[:j], disk.parts[j+1:]...)
				findPart = true
				break
			}
		}
	}
	//Si la particion no fue encontrada
	if !findPart {
		fmt.Println("[ALERTA]: La particion no fue encontrada")
		return
	}
	fmt.Println("La particion a sido desmontada con exito.")
}
