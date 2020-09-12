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

//Struct del super boot
type superBoot struct {
	HardDriveName               [16]byte
	VirtualTreeCount            int64
	DirectoryDetailCount        int64
	InodesCount                 int64
	BlocksCount                 int64
	VirtualTreeFree             int64
	DirectoryDetailFree         int64
	InodesFree                  int64
	BlocksFree                  int64
	CreationDate                [19]byte
	LastAssemblyDate            [19]byte
	MontageCount                int64
	PrDirectoryTreeBitmap       int64
	PrDirectoryTree             int64
	PrDirectoryDetailBitmap     int64
	PrDirectoryDetail           int64
	PrInodeTableBitmap          int64
	PrInodeTable                int64
	PrBlocksBitmap              int64
	PrBlocks                    int64
	PrLog                       int64
	SizeDirectoryTree           int64
	SizeDirectoryDetail         int64
	SizeInode                   int64
	SizeBlock                   int64
	FirstFreeBitDirectoryTree   int64
	FirstFreeBitDirectoryDetail int64
	FirstFreeBitInodeTable      int64
	FirstFreeBitBlocks          int64
	MagicNum                    [9]byte
}

//Struct del arbol virtual de directorio
type virtualDirectoryTree struct {
	CreatedAt              [19]byte
	DirectoryName          [20]byte
	Subdirectories         [6]int64
	PrDirectoryDetail      int64
	PrVirtualDirectoryTree int64
	Owner                  [16]byte
}

//Struct del detalle de directorio
type directoryDetail struct {
	Files             [5]ddFile
	PrDirectoryDetail int64
}

//Struct de archivos
type ddFile struct {
	Name             [20]byte
	PrInode          int64
	CreationDate     [19]byte
	ModificationDate [19]byte
}

//Struct del i-nodo
type iNode struct {
	Count          int64
	SizeFile       int64
	AllocatedBlock int64
	Blocks         [4]int64
	PrIndirect     int64
	Owner          [16]byte
}

//Struct del bloque de dato
type dataBlock struct {
	Data [25]byte
}

//Struct del LOG [Bitacora]
type bitacora struct {
	Operation       [6]byte
	Type            byte
	Name            [100]byte
	Content         [100]byte
	TransactionDate [19]byte
	Size            int64
}

//Comando MKFS para formatear una particion
func Mkfs(idPart string, Type string) {
	//Obtenemos la particion a partir del id
	path, mountedPart, err := searchPartition(idPart)
	if err != nil {
		return
	}
	//Obtenemos el file del disco
	file, _, err := readFile(path)
	if err != nil {
		return
	}
	//Definimos el tipo de particion que es
	partitionType := typeOf(mountedPart.partition)
	var primaryPartition partition
	var logicalPartition extendedBootRecord
	switch partitionType {
	case 0:
		primaryPartition = mountedPart.partition.(partition)
	case 1:
		logicalPartition = mountedPart.partition.(extendedBootRecord)
	}
	//Tamaños de las estruturas
	var partitionSize int64
	//Inicio de la particon
	var partitionStart int64
	//Nombre de la particion
	var partitionName string
	//Trabajamos con la particion primaria
	if primaryPartition.Status != 0 {
		partitionSize = primaryPartition.Size
		partitionStart = primaryPartition.Start
		partitionName = strings.Replace(string(primaryPartition.Name[:]), "\x00", "", -1)
	} else { //Trabajos con la particion logica
		partitionSize = logicalPartition.Size - int64(binary.Size(extendedBootRecord{}))
		partitionStart = logicalPartition.Start
		partitionName = strings.Replace(string(logicalPartition.Name[:]), "\x00", "", -1)
	}
	//Aplicamos la formateada full de la particion
	writeFormat(file, partitionStart, partitionSize)
	//Creamos el superbloque
	sb := createSB(partitionName, partitionSize, partitionStart)
	//Procedemos a escribir en el disco el superboot asignado a esa particion
	writeSB(file, partitionStart, &sb)
	//[a] Creamos la carpeta '/'
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	folder := virtualDirectoryTree{
		Subdirectories:         [6]int64{-1, -1, -1, -1, -1, -1},
		PrDirectoryDetail:      0,
		PrVirtualDirectoryTree: -1,
	}
	copy(folder.CreatedAt[:], timestamp)
	copy(folder.DirectoryName[:], "/")
	//Escribimos el arbol virtual de directorio de '/'
	writeVDT(file, sb.PrDirectoryTree, &folder)
	//Reescribimos el bitmap de arbol virtual de directorios
	bitMapVDT := []byte{'1'}
	writeBitmap(file, sb.PrDirectoryTreeBitmap, bitMapVDT)
	//[b] Creamos el detalle directorio de la carpeta '/'
	dd := structDD()
	//Escribimos el arbol virtual de directorio de '/'
	writeDD(file, sb.PrDirectoryDetail, &dd)
	//Reescribimos el bitmap de detellae de directorio
	bitMapDD := []byte{'1'}
	writeBitmap(file, sb.PrDirectoryDetailBitmap, bitMapDD)
	file.Close()
	fmt.Println("[-] Formateo exitoso.")
}

//Funcion para crear un superbloque
func createSB(partitionName string, partitionSize int64, partitionStart int64) superBoot {
	//Tamaños de las estructuras
	superBootSize := int64(binary.Size(superBoot{}))
	virtualTreeSize := int64(binary.Size(virtualDirectoryTree{}))
	directoryDetailSize := int64(binary.Size(directoryDetail{}))
	iNodeSize := int64(binary.Size(iNode{}))
	blockSize := int64(binary.Size(dataBlock{}))
	logSize := int64(binary.Size(bitacora{}))
	//Calculamos el numero de estructuras
	numberOfStructures := (partitionSize - 2*superBootSize) / (27 + virtualTreeSize + directoryDetailSize + 5*iNodeSize + 20*blockSize + logSize)
	//Creamos el superbloque para esta particion
	sb := superBoot{}
	//Nombre del disco duro virtual
	copy(sb.HardDriveName[:], partitionName)
	//Asignaos la cantidad de cada una de las estructuras
	sb.VirtualTreeCount = numberOfStructures
	sb.DirectoryDetailCount = numberOfStructures
	sb.InodesCount = 5 * numberOfStructures
	sb.BlocksCount = 20 * numberOfStructures
	//Cantidad de estructuras libres
	sb.VirtualTreeFree = numberOfStructures - 1
	sb.DirectoryDetailFree = numberOfStructures - 1
	sb.InodesFree = 5 * numberOfStructures
	sb.BlocksFree = 20 * numberOfStructures
	//Fechas
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	copy(sb.CreationDate[:], timestamp)
	copy(sb.LastAssemblyDate[:], timestamp)
	//Cantidad de montajes
	sb.MontageCount = 1
	//Apuntadores de cada una de las estructuras
	sb.PrDirectoryTreeBitmap = partitionStart + superBootSize
	sb.PrDirectoryTree = sb.PrDirectoryTreeBitmap + sb.VirtualTreeCount
	sb.PrDirectoryDetailBitmap = sb.PrDirectoryTree + virtualTreeSize*sb.VirtualTreeCount
	sb.PrDirectoryDetail = sb.PrDirectoryDetailBitmap + sb.DirectoryDetailCount
	sb.PrInodeTableBitmap = sb.PrDirectoryDetail + directoryDetailSize*sb.DirectoryDetailCount
	sb.PrInodeTable = sb.PrInodeTableBitmap + sb.InodesCount
	sb.PrBlocksBitmap = sb.PrInodeTable + iNodeSize*sb.InodesCount
	sb.PrBlocks = sb.PrBlocksBitmap + sb.BlocksCount
	sb.PrLog = sb.PrBlocks + blockSize*sb.BlocksCount
	//Tamaño de las estructuras del superboot
	sb.SizeDirectoryTree = virtualTreeSize
	sb.SizeDirectoryDetail = directoryDetailSize
	sb.SizeInode = iNodeSize
	sb.SizeBlock = blockSize
	//Los first free
	//[a] Se crea la carpeta '/' en la pasocion 0
	sb.FirstFreeBitDirectoryTree = 1
	sb.FirstFreeBitDirectoryDetail = 1
	sb.FirstFreeBitInodeTable = 0
	sb.FirstFreeBitBlocks = 0
	//Numero magico : Carnet
	copy(sb.MagicNum[:], "201807190")
	return sb
}

/*
	Funciones para la escritura de estructuras en el disco
*/

//Funcion para formatear la particion
func writeFormat(file *os.File, index int64, size int64) {
	format := make([]int8, size)
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, &format)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para escribir en el archivo la estructura de un super bloque de directorio
func writeSB(file *os.File, index int64, sb *superBoot) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, sb)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para escribir en el archivo la estructura de un arbol virtual de directorio
func writeVDT(file *os.File, index int64, vdt *virtualDirectoryTree) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, vdt)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para escribir en el archivo la estructura de un detalle de directorio
func writeDD(file *os.File, index int64, dd *directoryDetail) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, dd)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para reescribir algun bitmap en el disco
func writeBitmap(file *os.File, index int64, bitMap []byte) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, &bitMap)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para escribir en el archivo la estructura de un i-nodo
func writeInode(file *os.File, index int64, inode *iNode) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, inode)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para escribir en el archivo la estructura de un i-nodo
func writeBlock(file *os.File, index int64, data *dataBlock) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, data)
	writeNextBytes(file, binaryDisc.Bytes())
}

//Funcion para escribir en el archivo la estructura de un i-nodo
func writeBitacora(file *os.File, index int64, log *bitacora) {
	file.Seek(index, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, log)
	writeNextBytes(file, binaryDisc.Bytes())
}

/*
	Funciones para obtener estructuras en el disco
*/

//Funcion para recuperar el bitmap en el disco de alguna estructura
func getBitmap(file *os.File, index int64, size int64) []byte {
	bitMap := make([]byte, size)
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &bitMap)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return bitMap
}

//Funcion para recuperar un arbol virtual de directorios
func getVirtualDirectotyTree(file *os.File, pr int64, bm int64) virtualDirectoryTree {
	vdt := virtualDirectoryTree{}
	size := int64(binary.Size(vdt))
	index := pr + bm*size
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &vdt)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return vdt
}

//Funcion para recuperar un detalle de directorio
func getDirectotyDetail(file *os.File, pr int64, bm int64) directoryDetail {
	dd := directoryDetail{}
	size := int64(binary.Size(dd))
	index := pr + bm*size
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &dd)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return dd
}

//Funcion para recuperar un inodo
func getInode(file *os.File, pr int64, bm int64) iNode {
	inode := iNode{}
	size := int64(binary.Size(inode))
	index := pr + bm*size
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &inode)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return inode
}

//Funcion para recuperar un un bloque de datos
func getBlock(file *os.File, pr int64, bm int64) dataBlock {
	block := dataBlock{}
	size := int64(binary.Size(block))
	index := pr + bm*size
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &block)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return block
}

//Funcion para recuperar un un bloque de datos
func getBitacora(file *os.File, pr int64, bm int64) (bitacora, int64) {
	bita := bitacora{}
	size := int64(binary.Size(bita))
	index := pr + bm*size
	file.Seek(index, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err := binary.Read(buffer, binary.BigEndian, &bita)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return bita, index
}

//Funcion para recorrer la bitacora y obtener el ultimo punto libre
func getFreeLog(file *os.File, index int64, limit int64) int64 {
	//Obtenemos la bitacora inicial
	bita, pr := getBitacora(file, index, 0)
	for i := 1; bita.TransactionDate != [19]byte{}; i++ {
		//Obtenemos la siguiente bitacora
		bita, pr = getBitacora(file, index, int64(i))
	}
	limit = limit*int64(binary.Size(bita)) + index - int64(binary.Size(bita))
	if pr > limit {
		return -1 
	}
	return pr
}


//Funcion que te devuelve un struct virtual directory tree
func structLog(operation string, content string, name string, size int64, Type byte) bitacora {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	bita := bitacora{
		Type: Type,
		Size: size,
	}
	copy(bita.Operation[:], operation)
	copy(bita.Content[:], content)
	copy(bita.Name[:], name)
	copy(bita.TransactionDate[:], timestamp)
	return bita
}
