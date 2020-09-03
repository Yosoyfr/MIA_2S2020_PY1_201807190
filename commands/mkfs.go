package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	DirectoryName          [16]byte
	Subdirectories         [6]int64
	PrDirectoryDetail      int64
	PrVirtualDirectoryTree int64
	Owner                  int64
}

//Struct del detalle de directorio
type directoryDetail struct {
	Files             [5]ddFile
	PrDirectoryDetail int64
}

//Struct de archivos
type ddFile struct {
	Name             [16]byte
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
	Owner          int64
}

//Struct del bloque de dato
type dataBlock struct {
	Data [25]byte
}

//Struct del LOG [Bitacora]
type bitacora struct {
	Operation       [6]byte
	Type            int8
	Name            [16]byte
	Content         int8
	TransactionDate [19]byte
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
	defer file.Close()
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
	//Variable que representa el numero de estructuras
	var numberOfStructures int64
	//Tamaños de las estruturas
	var partitionSize int64
	//Inicio de la particon
	var partitionStart int64
	//Nombre de la particion
	var partitionName string
	superBootSize := int64(binary.Size(superBoot{}))
	virtualTreeSize := int64(binary.Size(virtualDirectoryTree{}))
	directoryDetailSize := int64(binary.Size(directoryDetail{}))
	iNodeSize := int64(binary.Size(iNode{}))
	blockSize := int64(binary.Size(dataBlock{}))
	logSize := int64(binary.Size(bitacora{}))
	//Trabajamos con la particion primaria
	if primaryPartition.Status != 0 {
		partitionSize = primaryPartition.Size
		partitionStart = primaryPartition.Start
		partitionName = strings.Replace(string(primaryPartition.Name[:]), "\x00", "", -1)
	} else { //Trabajos con la particion logica
		partitionSize = logicalPartition.Size
		partitionStart = logicalPartition.Start
		partitionName = strings.Replace(string(logicalPartition.Name[:]), "\x00", "", -1)
	}
	//Calculamos el numero de estructuras
	numberOfStructures = (partitionSize - 2*superBootSize) / (27 + virtualTreeSize + directoryDetailSize + 5*iNodeSize + 20*blockSize + logSize)
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
	sb.VirtualTreeFree = numberOfStructures
	sb.DirectoryDetailFree = numberOfStructures
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
	//Los first free inician en 0
	sb.FirstFreeBitDirectoryTree = 0
	sb.FirstFreeBitDirectoryDetail = 0
	sb.FirstFreeBitInodeTable = 0
	sb.FirstFreeBitBlocks = 0
	//Numero magico : Carnet
	copy(sb.MagicNum[:], "201807190")
	//Procedemos a escribir en el disco el superboot asignado a esa particion
	file.Seek(partitionStart, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct MBR
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, &sb)
	writeNextBytes(file, binaryDisc.Bytes())
	fmt.Println("[-] Formateo exitoso.")
}
