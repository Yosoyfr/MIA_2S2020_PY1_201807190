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
	createAllPath(file, &superboot, root, folders, 1)
	//Si el arreglo de carpetas es de dos, quiere decir que se va a escribir la carpeta en la raiz '/'
	/*
	if len(folders) == 2 {
		newFolder := createFolder(&root, folders[1], superboot.FirstFreeBitDirectoryTree)
		fmt.Println(root)
		fmt.Println(newFolder)
		writeVDT(file, superboot.PrDirectoryTree, &root)
		indexNF := superboot.FirstFreeBitDirectoryTree*superboot.SizeDirectoryTree + superboot.PrDirectoryTree
		writeVDT(file, indexNF, &newFolder)
		superboot.VirtualTreeFree--
		superboot.FirstFreeBitDirectoryTree++
		writeSB(file, indexSB, &superboot)
	} else {
		rootAux := root
		for i := 1; i < len(folders); i++ {
			prBM := existFolder(file, rootAux, folders, int64(i))
			fmt.Println(prBM)
			if prBM == -1 {
				break
			}
			rootAux = getVirtualDirectotyTree(file, superboot.PrDirectoryTree, prBM)
		}
	}

	if p {
	} else {

	}
	*/
	file.Close()
}

//Funcion para ir verificando si existe una ruta completa
func existPath(file *os.File, sb superBoot, vdt virtualDirectoryTree, folders []string, level int)  {
	var folderName [16]byte
	copy(folderName[:], folders[level])
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] != -1 {
			aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, vdt.Subdirectories[i])
			if folderName == aux .DirectoryName{
				if level + 1  < len(folders)  {
					level++
					existPath(file, sb, aux, folders, level)
				}	
				fmt.Println(string(aux.DirectoryName[:]))
			}
		}
	}
}

//Funcion para crear todo el path 
func createAllPath(file *os.File, sb *superBoot, vdt virtualDirectoryTree, folders []string, level int) {
	var folderName [16]byte
	copy(folderName[:], folders[level])
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] != -1 {
			aux := getVirtualDirectotyTree(file, sb.PrDirectoryTree, vdt.Subdirectories[i])
			if folderName == aux.DirectoryName{
				if level + 1  < len(folders)  {
					level++
					createAllPath(file, sb, aux, folders, level)
				}	
				fmt.Println(string(aux.DirectoryName[:]))
			}
			if true {
				//Creamos la carpeta que no se encuentra
				fmt.Println("Creamos: ", string(folderName[:]))
			}
		}
	}
}


func createFolder(vdt *virtualDirectoryTree, folder string, freeBit int64) virtualDirectoryTree {
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
	}
	//Creamos la nueva carpeta hijo
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newFolder := virtualDirectoryTree{
		Subdirectories:         [6]int64{-1, -1, -1, -1, -1, -1},
		PrDirectoryDetail:      0,
		PrVirtualDirectoryTree: -1,
		DirectoryName:          folderName,
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
