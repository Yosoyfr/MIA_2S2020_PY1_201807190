package commands

import (
	"fmt"
	"os"
	"strings"
)

//Permite mover una carpeta o archivo hacia otra
func Mv(id string, route string, dest string)  {
	//Revismos que las rutas sean correctas
	if route[0] != '/' || dest[0] != '/' {
		fmt.Println("[ERROR] Las rutas no son validad.")
		return
	}
	//Obtenemos las carpetas actuales
	folders := strings.Split(route, "/")
	folders = folders[1:]
	//Obtenemos las carpetas destino
	newfolders := strings.Split(dest, "/")
	newfolders = newfolders[1:]
	//Verificamos si es un archivo o carpeta lo que vamos a mover
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
	//moved := false
	if isFile {
		//Obtenemos el nombre del archivo
		var aux [20]byte
		copy(aux[:], folders[len(folders)-1])
		folders = folders[:len(folders)-1]
		//Obtener el puntero del DD del directorio de la carpeta actual en donde se encentra el archivo
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
		//Obtener el puntero del DD del directorio de la carpeta destino
		var destiny int64
		if len(newfolders) > 0 && newfolders[0] != ""{
			destiny = existDetailDirectory(file, &sb, root, newfolders, 0)
		} else {
			destiny = root.PrDirectoryDetail
		}
		if destiny == -1 {
			fmt.Println("[ERROR]: El directorio destino", dest, "no existe.")
			file.Close()
			return
		}
		//Recuperamos el archivo y editamos el detalle de directorio donde se encontraba a un valor default
		f := editFilePointer(file, sb, index, aux, -1)
		if f.PrInode != -1 {
			//Obtenemos el detalle de directorio destino
			dd := getDirectotyDetail(file, sb.PrDirectoryDetail, destiny)
			//Obtenemos el detalle correspondiente
			dd, destiny = firstFitDD(file, &sb, dd, destiny)
			//Asignamos el archivo al directorio destino
			assignmentFile(file, &sb, &dd, f)
			//Reescribimos en el disco el directorio destino
			pr := destiny*sb.SizeDirectoryDetail + sb.PrDirectoryDetail
			writeDD(file, pr, &dd)
			fmt.Println("[-] El archivo", route, "a sido movido hacia", dest,"con exito.")
		} else {
			fmt.Println("[ERROR] El archivo a mover no ha sido encontrado.")
		}
	}else {
		if route == "/" {
			fmt.Println("[ERROR]: La carpeta raiz del sistema no puede moverse.")
			file.Close()
			return
		}
		//Quitamos el puntero de la carpeta a mover y el directorio padre
		//index :=  editDirectoryPointer(file, sb, root, folders)
		//fmt.Println(index)
	}
	//Reescribimos el superboot
	writeSB(file, indexSB, &sb)
	file.Close()
}

//Funcion para editar un puntero del detalle de directirio
func editFilePointer(file *os.File, sb superBoot, index int64, filename [20]byte, bm int64) (ddFile) {
	//Obtenemos el detalle de directorio
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, index)
	for i := 0; i < len(dd.Files); i++ {
		if dd.Files[i].Name == filename {
			f := dd.Files[i]
			dd.Files[i] = ddFile{PrInode: bm}
			//Reescribimos el detalle de directorio
			pr := sb.PrDirectoryDetail + index*sb.SizeDirectoryDetail
			writeDD(file, pr, &dd)
			return f
		}
	}
	if dd.PrDirectoryDetail != -1 {
		return editFilePointer(file, sb, dd.PrDirectoryDetail, filename, bm)
	}
	return ddFile{PrInode: -1}
}


func editDirectoryPointer(file *os.File, sb superBoot, vdt virtualDirectoryTree, folders []string) (int64) {
	//Casteamos el nombre del VDT
	var auxVDT [20]byte
	copy(auxVDT[:], folders[0])
	//Lo quitamos de la lista de carpetas
	folders = folders[1:]
	//Identificamos el puntero de la carpeta a buscar
	index := existPath(file, &sb, vdt, auxVDT)
	if index != -1 {
		//Obtenemos el vdt de ese puntero
		aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, index)
		//Iteramos una vez mas el metodo si el arreglo de carpetas aun contiene datos
		fmt.Println(vdt)
		if len(folders) > 0 {
			return editDirectoryPointer(file, sb, aux, folders)
		}
		//Comparamos y retornamos el indice del directorio
		if aux.DirectoryName == auxVDT {
			//pr := sb.PrDirectoryTree + index*sb.SizeDirectoryTree
			//writeVDT(file, pr, &aux)
			return index
		}
	}
	return -1
}