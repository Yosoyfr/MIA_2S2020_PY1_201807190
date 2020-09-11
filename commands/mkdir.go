package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

//Funcion para crear carpetas del sistema de archivos
func Mkdir(id string, route string, p bool) {
	//Revismos que la ruta a insertar sea correcta
	if route[0] != '/' {
		fmt.Println("[ERROR] El path a crear no es valido.")
		return
	}
	//Obtenemos las carpetas
	folders := strings.Split(route, "/")
	folders = folders[1:]
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
	superboot := getSB(file, indexSB)
	//[-] Proceso de escritura de una nueva carpeta
	//Recuperamos el arbol de directorio de '/'
	root := getVirtualDirectotyTree(file, superboot.PrDirectoryTree, 0)
	//Funcion para crear todo el path que se asigne como parametro
	if p {
		createAllPath(file, &superboot, indexSB, root, folders, 0)
	} else {
		createPath(file, &superboot, indexSB, root, folders, 0)
	}
	file.Close()
}

//Funcion para ir verificando si existe una ruta
func existPath(file *os.File, sb *superBoot, vdt virtualDirectoryTree, folder [20]byte) int64 {
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] != -1 {
			aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, vdt.Subdirectories[i])
			if folder == aux.DirectoryName {
				return vdt.Subdirectories[i]
			}
		}
	}
	if vdt.PrVirtualDirectoryTree != -1 {
		aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, vdt.PrVirtualDirectoryTree)
		return existPath(file, sb, aux, folder)
	}
	return -1
}

//Funcion para crear todo el path completo que se le asigne
func createAllPath(file *os.File, sb *superBoot, indexSB int64, vdt virtualDirectoryTree, folders []string, bm int64) {
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
		if len(folders) > 0 {
			createAllPath(file, sb, indexSB, aux, folders, index)
		}
	} else {
		//Si no existe creamos ese arbol de directorio
		//Recuperamos el bitmap donde sera insertado
		temp := sb.FirstFreeBitDirectoryTree
		vdt, bm = firstFitVDT(file, sb, vdt, bm)
		//Procedemos a construir la estructura padre e hijo de los vdt trabajados
		buildVDT(file, sb, indexSB, vdt, bm, string(auxVDT[:]))
		fmt.Println("[-] El directorio \"", string(auxVDT[:]), "\" ha sido creado con exito.")
		//Recuperamos el ultimo hijo insertado
		aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, temp)
		//Iteramos una vez mas el metodo si el arreglo de carpetas aun contiene datos
		if len(folders) > 0 {
			createAllPath(file, sb, indexSB, aux, folders, temp)
		}
	}
}

//Funcion para crear la ultima carpeta de un path
func createPath(file *os.File, sb *superBoot, indexSB int64, vdt virtualDirectoryTree, folders []string, bm int64) {
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
		if len(folders) > 0 {
			createPath(file, sb, indexSB, aux, folders, index)
		}
	} else {
		//Si no existe nos salta error de que el directorio no existe
		if len(folders) > 0 {
			fmt.Println("[ERROR]: El directorio donde se desea crear la carpeta no existe.")
		} else {
			vdt, bm = firstFitVDT(file, sb, vdt, bm)
			//Procedemos a construir la estructura padre e hijo de los vdt trabajados
			buildVDT(file, sb, indexSB, vdt, bm, string(auxVDT[:]))
			fmt.Println("[-] El directorio", string(auxVDT[:]), "ha sido creado con exito.")
		}
	}
}

//Obtenemos el directorio donde se va a insertar
func firstFitVDT(file *os.File, sb *superBoot, vdt virtualDirectoryTree, bm int64) (virtualDirectoryTree, int64) {
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] == -1 {
			return vdt, bm
		}
	}
	if vdt.PrVirtualDirectoryTree != -1 {
		aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, vdt.PrVirtualDirectoryTree)
		return firstFitVDT(file, sb, aux, vdt.PrVirtualDirectoryTree)
	}
	return vdt, bm
}

func buildVDT(file *os.File, sb *superBoot, indexSB int64, vdt virtualDirectoryTree, bm int64, folder string) {
	//Creamos el nuevo VDT
	newVDT := createVDT(file, sb, indexSB, &vdt, folder, sb.FirstFreeBitDirectoryTree)
	//Obtenemos la posicion en donde sera reescrito el vdt padre
	index := bm*sb.SizeDirectoryTree + sb.PrDirectoryTree
	//Reescribimos el vdt padre
	writeVDT(file, index, &vdt)
	//Reescribimos el bitmap de arbol virtual de directorios
	bitMapVDT := []byte{'1'}
	writeBitmap(file, sb.PrDirectoryTreeBitmap+sb.FirstFreeBitDirectoryTree, bitMapVDT)
	//Obtenemos la posicion en donde sera insertado el vdt hijo
	indexNVDT := sb.FirstFreeBitDirectoryTree*sb.SizeDirectoryTree + sb.PrDirectoryTree
	//Insertamos el vdt hijo
	writeVDT(file, indexNVDT, &newVDT)
	//Creamos el detalle directorio del nuevo directorio
	dd := structDD()
	//Escribimos el arbol virtual de directorio de '/'
	indexDD := sb.FirstFreeBitDirectoryDetail*sb.SizeDirectoryDetail + sb.PrDirectoryDetail
	writeDD(file, indexDD, &dd)
	//Reescribimos el bitmap de detellae de directorio
	bitMapDD := []byte{'1'}
	writeBitmap(file, sb.PrDirectoryDetailBitmap+sb.FirstFreeBitDirectoryDetail, bitMapDD)
	//Reescribimos valores del superboot
	sb.VirtualTreeFree--
	sb.FirstFreeBitDirectoryTree++
	sb.DirectoryDetailFree--
	sb.FirstFreeBitDirectoryDetail++
	//Reescribimos el superboot
	writeSB(file, indexSB, sb)
}

func createVDT(file *os.File, sb *superBoot, indexSB int64, vdt *virtualDirectoryTree, folder string, freeBit int64) virtualDirectoryTree {
	var folderName [20]byte
	copy(folderName[:], folder)
	created := false
	//Asignamos el puntero
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] == -1 {
			vdt.Subdirectories[i] = freeBit
			created = true
			break
		}
	}
	//Si en dado caso no hay espacio para nuevo puntero se crea el puntero indirecto de la carpeta
	if !created {
		//Asignamos el nuevo apuntador que sera el indirecto
		vdt.PrVirtualDirectoryTree = freeBit
		//Creamos el directorio indirecto
		indirect := structVDT(vdt.DirectoryName, vdt.PrDirectoryDetail)
		indirect.CreatedAt = vdt.CreatedAt
		//Escribimos el apuntador indirecto en el archivo
		freeBit++
		auxVDT := createVDT(file, sb, indexSB, &indirect, folder, freeBit)
		//Obtenemos la posicion en donde sera reescrito el vdt padre
		indexPRVDT := vdt.PrVirtualDirectoryTree*sb.SizeDirectoryTree + sb.PrDirectoryTree
		//Reescribimos el vdt padre
		writeVDT(file, indexPRVDT, &indirect)
		//Reescribimos el bitmap de arbol virtual de directorios
		bitMapVDT := []byte{'1'}
		writeBitmap(file, sb.PrDirectoryTreeBitmap+vdt.PrVirtualDirectoryTree, bitMapVDT)
		//Reescribimos valores del superboot
		sb.VirtualTreeFree--
		sb.FirstFreeBitDirectoryTree++
		//Reescribimos el superboot
		writeSB(file, indexSB, sb)
		return auxVDT
	}
	//Creamos la nueva carpeta hijo
	newFolder := structVDT(folderName, sb.FirstFreeBitDirectoryDetail)
	return newFolder
}

//Funcion que te devuelve un struct virtual directory tree
func structVDT(name [20]byte, prDD int64) virtualDirectoryTree {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newFolder := virtualDirectoryTree{
		Subdirectories:         [6]int64{-1, -1, -1, -1, -1, -1},
		PrDirectoryDetail:      prDD,
		PrVirtualDirectoryTree: -1,
		DirectoryName:          name,
	}
	copy(newFolder.CreatedAt[:], timestamp)
	return newFolder
}

//Funcion que te devuelve un struct directory detail
func structDD() directoryDetail {
	dd := directoryDetail{PrDirectoryDetail: -1}
	//Creamos un File que vendra con el puntero del inodo en -1 por defecto
	f := ddFile{PrInode: -1}
	for i := 0; i < len(dd.Files); i++ {
		dd.Files[i] = f
	}
	return dd
}

//Funcion para obtener el superbloque de una particion
func getSB(file *os.File, index int64) superBoot {
	superboot := superBoot{}
	//Nos posicionamos en esa parte del archivo
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, int64(binary.Size(superboot)))
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &superboot)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return superboot
}

//Funcion para definir el tipo de particion que estamos trabajando y nos devuelve la posicion inicial de la particion y el nombre
func getPartitionType(mountedPart mountedParts) (int64, string) {
	//Definomos el tipo de particion
	partitionType := typeOf(mountedPart.partition)
	var primaryPartition partition
	var logicalPartition extendedBootRecord
	switch partitionType {
	case 0:
		primaryPartition = mountedPart.partition.(partition)
	case 1:
		logicalPartition = mountedPart.partition.(extendedBootRecord)
	}
	//Posicion del bit donde comienza el superboot de esa particon
	var indexSB int64
	//Nombre de la particon
	var name string
	//Trabajamos con la particion primaria
	if primaryPartition.Status != 0 {
		indexSB = primaryPartition.Start
		name = strings.Replace(string(primaryPartition.Name[:]), "\x00", "", -1)
	} else { //Trabajos con la particion logica
		indexSB = logicalPartition.Start
		name = strings.Replace(string(logicalPartition.Name[:]), "\x00", "", -1)
	}
	return indexSB, name
}
