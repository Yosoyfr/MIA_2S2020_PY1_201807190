package commands

import (
	"fmt"
	"os"
	"strings"
	"time"
)

//Funcion MKFILE para la creacion de archivos txt en los detalles de directorio
func Mkfile(id string, route string, p bool, size int64, txt string) {
	//Revismos que la ruta a insertar sea correcta
	if route[0] != '/' {
		fmt.Println("[ERROR] El path no es valido.")
		return
	}
	//Obtenemos las carpetas
	folders := strings.Split(route, "/")
	folders = folders[1:]
	//Obtenemos el nombre del archivo
	filename := folders[len(folders)-1]
	folders = folders[:len(folders)-1]
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
	//[-] Proceso de escribir inodos y data en bloques que representan un archivo
	//Recuperamos el arbol de directorio de '/'
	root := getVirtualDirectotyTree(file, sb.PrDirectoryTree, 0)
	//Si el comando requiere que creemos todo el directorio donde estara el archivo
	if p {
		createAllPath(file, &sb, indexSB, root, folders, 0)
	}
	//De lo contrario tendra que existir ya el directorio completo
	//Procedemos a obtener el puntero del DD del directorio
	var index int64
	if len(folders) > 0 {
		index = existDetailDirectory(file, &sb, root, folders, 0)
	} else {
		index = root.PrDirectoryDetail
	}
	if index == -1 {
		fmt.Println("[ERROR]: El directorio donde se desea crear el archivo no existe.")
		return
	}
	//Obtenemos el detalle de directorio
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, index)
	//Obtenemos el detalle correspondiente
	dd, index = firstFitDD(file, &sb, dd, index)
	//Asignamos el archivo creado al detalle de directorio
	assignmentFile(file, &sb, indexSB, &dd, filename)
	indexDD := index*sb.SizeDirectoryDetail + sb.PrDirectoryDetail
	//Procedemos a reescribir el detalle de directorio
	writeDD(file, indexDD, &dd)
	//Obtenemos el arreglo de bloques que se generaron a partir de la cadena de entrada
	arrDB := structDataBlocks(txt)
	//Creamos el inodo que hace referencia a ese archivo y los escribimos
	inode := structInode(&sb, int64(len(txt)), int64(len(arrDB)))
	indexI := sb.PrInodeTable + inode.Count*sb.SizeInode
	writeInode(file, indexI, &inode)
	//Procedemos a escribir la estructura de todos los inodos que sean necesarios
	if int64(len(arrDB)) != 0 {
		buildInodes(file, &sb, inode, arrDB)
	}
	//Reescribimos el superboot
	writeSB(file, indexSB, &sb)
	file.Close()
}

func buildInodes(file *os.File, sb *superBoot, inode iNode, arrDB []dataBlock) {
	//Obtenemos la data del primer bloque de la lista de bloques generados
	data := arrDB[0]
	//Obtenemos el indice del bitmap a donde se va a escribir
	bmDB := sb.FirstFreeBitBlocks
	//Escribimos el bloque en el disco
	indexDB := sb.PrBlocks + bmDB*sb.SizeBlock
	writeBlock(file, indexDB, &data)
	//Reescribimos nuestro bitmap de bloques
	writeBitmap(file, sb.PrBlocksBitmap+bmDB, []byte{'1'})
	//Eliminamos ese arreglo de datablock
	arrDB = arrDB[1:]
	//Procedemos a editar el inodo que lo contiene
	inode = createInodes(sb, inode)
	//Procedemos a escribir o reescribir el inodo segun sea el caso
	indexI := sb.PrInodeTable + inode.Count*sb.SizeInode
	writeInode(file, indexI, &inode)
	//Reescribimos nuestro bitmap de inodos
	writeBitmap(file, sb.PrInodeTableBitmap+inode.Count, []byte{'1'})
	if len(arrDB) > 0 {
		buildInodes(file, sb, inode, arrDB)
	}
}

//Funcion que crea la estructura de inodos con su bloques
//bm es el apuntador del bitmap de inodos donde se va a crear
func createInodes(sb *superBoot, inode iNode) iNode {
	inserted := false
	for i := 0; i < len(inode.Blocks); i++ {
		if inode.Blocks[i] == -1 {
			inode.Blocks[i] = sb.FirstFreeBitBlocks
			sb.FirstFreeBitBlocks++
			sb.BlocksFree--
			inserted = true
			if i == len(inode.Blocks)-1 {
				inode.PrIndirect = sb.FirstFreeBitInodeTable
			}
			break
		}
	}
	//Si no se ha insertado es porque todas las casillas estan ocupadas y se tienen que crear un nuevo inodo que sera indirecto del que se esta trabajando
	if !inserted {
		aux := structInode(sb, inode.SizeFile, inode.AllocatedBlock)
		return createInodes(sb, aux)
	}
	return inode
}

//Funcion que devuelve un struct Inode
func structInode(sb *superBoot, size int64, dbs int64) iNode {
	//Struct de un inodo
	inode := iNode{
		Count:          sb.FirstFreeBitInodeTable,
		SizeFile:       size,
		Blocks:         [4]int64{-1, -1, -1, -1},
		PrIndirect:     -1,
		AllocatedBlock: dbs,
	}
	sb.InodesFree--
	sb.FirstFreeBitInodeTable++
	return inode
}

func structDataBlocks(txt string) []dataBlock {
	//Creamos los bloques que contendran 25 caracteres cada uno
	var arr []dataBlock
	var aux dataBlock
	for i, j := 0, 0; i < len(txt); i++ {
		aux.Data[j] = txt[i]
		j++
		if j == 25 {
			j = 0
			arr = append(arr, aux)
			aux = dataBlock{}
		}
	}
	if aux.Data[0] != 0 {
		arr = append(arr, aux)
	}
	return arr
}

//Obtenemos el detalle de directorio donde se va a insertar
func firstFitDD(file *os.File, sb *superBoot, dd directoryDetail, bm int64) (directoryDetail, int64) {
	for i := 0; i < len(dd.Files); i++ {
		if dd.Files[i].PrInode == -1 {
			return dd, bm
		}
	}
	if dd.PrDirectoryDetail != -1 {
		aux := getDirectotyDetail(file, sb.PrDirectoryDetail, dd.PrDirectoryDetail)
		return firstFitDD(file, sb, aux, dd.PrDirectoryDetail)
	}
	return dd, bm
}

//Funcion que nos retorna el puntero del detalle de directorio de un directorio
func existDetailDirectory(file *os.File, sb *superBoot, vdt virtualDirectoryTree, folders []string, bm int64) int64 {
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
			return existDetailDirectory(file, sb, aux, folders, index)
		}
		return aux.PrDirectoryDetail
	}
	return -1
}

//Funcion para construir una estructrura de un inode y su DDfile
func assignmentFile(file *os.File, sb *superBoot, indexSB int64, dd *directoryDetail, name string) {
	//Casteamos el nombre del archivo
	var filename [16]byte
	copy(filename[:], name)
	//Creamos el DDFILE
	f := structDDF(filename, sb.FirstFreeBitInodeTable)
	created := false
	//Asignamos el puntero
	for i := 0; i < len(dd.Files); i++ {
		if dd.Files[i].PrInode == -1 {
			dd.Files[i] = f
			created = true
			break
		}
	}
	//Si en dado caso no se ha creado, quiere decir que ese detalle de directorio ya esta lleno, por lo que es custion de crear un indirecto
	if !created {
		//Asignamos el nuevo apuntador que sera indirecto del actual
		dd.PrDirectoryDetail = sb.FirstFreeBitDirectoryDetail
		//Creamos el detalle de directorio indirecto
		indirect := structDD()
		assignmentFile(file, sb, indexSB, &indirect, name)
		//Obtenemos la posicion en donde sera escrito el detalle de directorio indirecto
		indexPRDD := dd.PrDirectoryDetail*sb.SizeDirectoryDetail + sb.PrDirectoryDetail
		//Procedemos a escribir en el disco el detalle de directorio indirecto
		writeDD(file, indexPRDD, &indirect)
		//Reescribimos el bitmap de detalle de directorio
		writeBitmap(file, sb.PrDirectoryTreeBitmap+dd.PrDirectoryDetail, []byte{'1'})
		sb.FirstFreeBitDirectoryDetail++
		sb.DirectoryDetailFree--
	}
}

//Funcion que devuelve un struct Directory Detail File
func structDDF(filename [16]byte, prInode int64) ddFile {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f := ddFile{
		Name:    filename,
		PrInode: prInode,
	}
	copy(f.CreationDate[:], timestamp)
	copy(f.ModificationDate[:], timestamp)
	return f
}
