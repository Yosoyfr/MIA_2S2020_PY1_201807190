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
	}else {
		createPath(file, &superboot, indexSB, root, folders, 0)
	}
	file.Close()
}

//Funcion para ir verificando si existe una ruta
func existPath(file *os.File, sb *superBoot, vdt virtualDirectoryTree, folder [16]byte) int64 {
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] != -1 {
			aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, vdt.Subdirectories[i])
			if folder == aux.DirectoryName {
				return vdt.Subdirectories[i]
			}
		}
	}
	return -1
}

//Funcion para crear todo el path completo que se le asigne
func createAllPath(file *os.File, sb *superBoot, indexSB int64, vdt virtualDirectoryTree, folders []string, bm int64) {
	//Casteamos el nombre del VDT
	var auxVDT [16]byte
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
		fmt.Println("En el directorio ", string(vdt.DirectoryName[:]))
		fmt.Println("Crear subdirectorio:", string(auxVDT[:]))
		//Recuperamos el bitmap donde sera insertado
		temp := sb.FirstFreeBitDirectoryTree
		//Procedemos a construir la estructura padre e hijo de los vdt trabajados
		buildVDT(file, sb, indexSB, vdt, bm, string(auxVDT[:]))
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
	var auxVDT [16]byte
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
		//Si no existe creamos ese arbol de directorio
		fmt.Println("En el directorio ", string(vdt.DirectoryName[:]))
		fmt.Println("Crear subdirectorio:", string(auxVDT[:]))
		if len(folders) > 0 {
			fmt.Println("[ERROR]: El directorio donde se desea crear la carpeta no existe")	
		} else {
			//Procedemos a construir la estructura padre e hijo de los vdt trabajados
			buildVDT(file, sb, indexSB, vdt, bm, string(auxVDT[:]))
		}
	}
}

func buildVDT(file *os.File, sb *superBoot, indexSB int64, vdt virtualDirectoryTree, bm int64, folder string) {
	//Creamos el nuevo VDT
	newVDT := createVDT(&vdt, folder, sb.FirstFreeBitDirectoryTree)
	fmt.Println(vdt)
	fmt.Println(newVDT)
	/*
	//Obtenemos la posicion en donde sera reescrito el vdt padre
	index := bm*sb.SizeDirectoryTree + sb.PrDirectoryTree
	//Reescribimos el vdt padre
	writeVDT(file, index, &vdt)
	//Obtenemos la posicion en donde sera insertado el vdt hijo
	indexNVDT := sb.FirstFreeBitDirectoryTree*sb.SizeDirectoryTree + sb.PrDirectoryTree
	//Insertamos el vdt hijo
	writeVDT(file, indexNVDT, &newVDT)
	//Reescribimos valores del superboot
	sb.VirtualTreeFree--
	sb.FirstFreeBitDirectoryTree++
	//Reescribimos el superboot
	writeSB(file, indexSB, sb)
	*/
}

func createVDT(vdt *virtualDirectoryTree, folder string, freeBit int64) virtualDirectoryTree {
	var folderName [16]byte
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
		vdt.PrVirtualDirectoryTree = freeBit
		fmt.Println("Crear apuntador indirecto")
		indirect := structVDT(vdt.DirectoryName)
		indirect.CreatedAt = vdt.CreatedAt
		fmt.Println("Indirect:", indirect)
		//Escribimos el apuntador indirecto en el archivo
		freeBit++
		return createVDT(&indirect, folder, freeBit)
	}
	fmt.Println("*******************")
	//Creamos la nueva carpeta hijo
	newFolder := structVDT(folderName)
	fmt.Println(vdt)
	fmt.Println(newFolder)
	return newFolder
}

//Funcion que te devuelve un struct virtual directory tree
func structVDT(name [16]byte) virtualDirectoryTree {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newFolder := virtualDirectoryTree{
		Subdirectories:         [6]int64{-1, -1, -1, -1, -1, -1},
		PrDirectoryDetail:      0,
		PrVirtualDirectoryTree: -1,
		DirectoryName:          name,
	}
	copy(newFolder.CreatedAt[:], timestamp)
	return newFolder
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