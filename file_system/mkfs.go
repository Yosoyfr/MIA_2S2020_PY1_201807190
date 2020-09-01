package file_system

import (
	"encoding/binary"
	"fmt"
)

//Struct del super boot
type superBoot struct {
	hardDriveName               [16]byte
	virtualTreeCount            int64
	directoryDetailCount        int64
	inodesCount                 int64
	blockCount                  int64
	virtualTreeFree             int64
	directoryDetailFree         int64
	inodesFree                  int64
	blocksFree                  int64
	creationDate                [19]byte
	lastAssemblyDate            [19]byte
	montageCount                int64
	prDirectoryTreeBitmap       int64
	prDirectoryTree             int64
	prDirectoryDetailBitmap     int64
	prDirectoryDetail           int64
	prInodeTableBitmap          int64
	prInodeTable                int64
	prBlocksBitmap              int64
	prBlocks                    int64
	prLog                       int64
	sizeDirectoryTree           int64
	sizeDirectoryDetail         int64
	sizeInode                   int64
	sizeBlock                   int64
	firstFreeBitDirectoryTree   int64
	firstFreeBitDirectoryDetail int64
	firstFreeBitInodeTable      int64
	firstFreeBitBlocks          int64
	magicNum                    int32
}

//Struct del arbol virtual de directorio
type virtualDirectoryTree struct {
	createdAt              [19]byte
	directoryName          [16]byte
	subdirectories         [6]int64
	prDirectoryDetail      int64
	prVirtualDirectoryTree int64
	owner                  int64
}

//Struct del detalle de directorio
type directoryDetail struct {
	files             [5]ddFile
	prDirectoryDetail int64
}

//Struct de archivos
type ddFile struct {
	name             [16]byte
	prInode          int64
	creationDate     [19]byte
	modificationDate [19]byte
}

//Struct del i-nodo
type iNode struct {
	count          int64
	sizeFile       int64
	allocatedBlock int64
	blocks         [4]int64
	prIndirect     int64
	owner          int64
}

//Struct del bloque de dato
type dataBlock struct {
	data [25]byte
}

//Struct del LOG [Bitacora]
type log struct {
	operation       [6]byte
	Type            int8
	name            [16]byte
	content         int8
	transactionDate [19]byte
}

func PruebaFile() {
	alv := log{}
	fmt.Println(binary.Size(alv))
}
