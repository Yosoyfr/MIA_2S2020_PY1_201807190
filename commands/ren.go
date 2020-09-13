package commands

import (
	"fmt"
	"os"
	"strings"
)

//Permite cambiar el nombre de un archivo o carpeta en el sistema de archivios
func Ren(id string, route string, name string) {
	//Revismos que la ruta a insertar sea correcta
	if route[0] != '/' {
		fmt.Println("[ERROR] El path no es valido.")
		return
	}
	//Obtenemos las carpetas
	folders := strings.Split(route, "/")
	folders = folders[1:]
	//Verificamos si es un archivo o carpeta lo que vamos a editar
	isFile := false
	if strings.HasSuffix(strings.ToLower(folders[len(folders)-1]), ".txt") {
		isFile = true
	}
	//Obtenemos la particion a partir del id
	path, mountedPart, err := searchPartition(id)
	if err != nil {
		return
	}
	//Obtenemos el file del disco
	file, _, err := readFile(path)
	if err != nil {
		return
	}
	//Definimos el tipo de particion que es
	indexSB, _ := getPartitionType(mountedPart)
	//Recuperamos el superbloque de la particion
	sb := getSB(file, indexSB)
	//Recuperamos el arbol de directorio de '/'
	root := getVirtualDirectotyTree(file, sb.PrDirectoryTree, 0)
	if isFile {
		//Obtenemos el nombre del archivo
		var aux [20]byte
		copy(aux[:], folders[len(folders)-1])
		folders = folders[:len(folders)-1]
		//Procedemos a obtener el puntero del DD del directorio
		var index int64
		if len(folders) > 0 {
			index = existDetailDirectory(file, &sb, root, folders, 0)
		} else {
			index = root.PrDirectoryDetail
		}
		if index == -1 {
			fmt.Println("[ERROR]: El directorio", route, "no existe.")
			file.Close()
			return
		}
		//Buscamos el archivos para que sea editado
		changed := changeFilename(file, sb, aux, index, name)
		fmt.Println(changed)
	} else {
		changeFoldername(file, &sb, root, folders, name)
	}
	file.Close()
}

//Funcion que recorre todo el detalle de directorio para encontrar un archivo
func changeFilename(file *os.File, sb superBoot, filename [20]byte, index int64, newname string) bool {
	//Obtenemos el detalle de directorio
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, index)
	//Recorremos para encontrar su
	for i := 0; i < len(dd.Files); i++ {
		if dd.Files[i].Name == filename {
			copy(dd.Files[i].Name[:], newname)
			return true
		}
	}
	if dd.PrDirectoryDetail != -1 {
		return changeFilename(file, sb, filename, dd.PrDirectoryDetail, newname)
	}
	return false
}

func changeFoldername(file *os.File, sb *superBoot, vdt virtualDirectoryTree, folders []string, newname string) bool {
	//Casteamos el nombre del VDT
	var auxVDT [20]byte
	copy(auxVDT[:], folders[0])
	//Lo quitamos de la lista de carpetas
	folders = folders[1:]
	//Identificamos el puntero de la carpeta a buscar
	index := existPath(file, sb, vdt, auxVDT)
	if index != -1 {
		//Obtenemos el vdt de ese puntero
		aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, index)
		//Iteramos una vez mas el metodo si el arreglo de carpetas aun contiene datos
		if len(folders) > 1 {
			return changeFoldername(file, sb, aux, folders, newname)
		}
	}
	fmt.Println("La carpeta a comparar con la nueva es:", string(auxVDT[:]))
	return false
}
