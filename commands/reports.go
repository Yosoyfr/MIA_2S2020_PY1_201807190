package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

//Funcion manager del tipo de reporte a crear
func Reports(id string, rep string, path string, route string) {
	//Carpeta en donde se crear el reporte
	dir := filepath.Dir(path)
	createDirectory(dir)
	//Reportes de bitmaps
	if strings.HasPrefix(rep, "BM_") {
		reportBM(id, rep, path, route)
		return
	}
	//Superboot a trabajar
	sb := superBoot{}
	//Obtenemos el file y la particion a trabajar
	diskPath, mountedPart, err := searchPartition(id)
	if err != nil {
		return
	}
	file, mbr, err := readFile(diskPath)
	if err != nil {
		return
	}
	//Definimos el tipo de particion que es
	indexSB, name := getPartitionType(mountedPart)
	//Recuperamos el superbloque de la particion
	sb = getSB(file, indexSB)
	//Reportes con GRAPHVIZ
	var report string
	switch rep {
	case "MBR":
		fmt.Println("[REPORT] Creando reporte de MBR.")
		report = reportMBR(file, mbr)
	case "DISK":
		fmt.Println("[REPORT] Creando reporte de DISK.")
		report = reportDisk(file, mbr)
	case "SB":
		fmt.Println("[REPORT] Creando reporte de SB.")
		report = reportSuperBoot(file, mountedPart, sb, name)
	case "DIRECTORIO":
		fmt.Println("[REPORT] Creando reporte de DIRECTORIO.")
		report = reportVirtualDirectoryTree(file, sb)
	case "TREE_COMPLETE":
		fmt.Println("[REPORT] Creando reporte de TREE_COMPLETE.")
		report = reportTreeComplete(file, sb)
	case "TREE_DIRECTORIO":
		if len(route) == 0 {
			fmt.Println("[ERROR]: Necesita especificar una ruta para crear el reporte.")
			err = fmt.Errorf("ERROR")
		} else {
			fmt.Println("[REPORT] Creando reporte de TREE_DIRECTORIO.")
			report, err = reportDirectoryTree(file, sb, route)
		}
	case "TREE_FILE":
		if len(route) == 0 {
			fmt.Println("[ERROR]: Necesita especificar una ruta para crear el reporte.")
			err = fmt.Errorf("ERROR")
		} else {
			fmt.Println("[REPORT] Creando reporte de TREE_FILE.")
			report, err = reportTreeFile(file, sb, route)
		}
	case "BITACORA":
		fmt.Println("[REPORT] Creando reporte de BITACORA.")
		report = reportLog(file, sb)
	default:
		fmt.Println("[ERROR]: El tipo de reporte a crear no existe |", rep, "|.")
		err = fmt.Errorf("ERROR")
	}
	file.Close()
	//Verificamos que no se hubieran generado errores en la creacion de los dots
	if err != nil {
		return
	}
	err = ioutil.WriteFile("report.dot", []byte(report), 0644)
	if err != nil {
		log.Fatal(err)
	}
	extension := path[(len(path) - 3):(len(path))]
	exec.Command("dot", "-T"+extension, "report.dot", "-o", path).Output()
	fmt.Println("[REPORT] El reporte fue generado con exito.")
}

func reportBM(id string, rep string, path string, route string) {
	//Superboot a trabajar
	sb := superBoot{}
	//Obtenemos el file y la particion a trabajar
	diskPath, mountedPart, err := searchPartition(id)
	if err != nil {
		return
	}
	file, _, err := readFile(diskPath)
	if err != nil {
		return
	}
	//Definimos el tipo de particion que es
	indexSB, _ := getPartitionType(mountedPart)
	//Recuperamos el superbloque de la particion
	sb = getSB(file, indexSB)
	//Recuperamos el bitmap que se indique en REP
	var bitmap []byte
	switch rep {
	case "BM_ARBDIR":
		fmt.Println("[BITMAP] Creando reporte de Bitmap de arbol de directorios.")
		sizeBM := sb.PrDirectoryTree - sb.PrDirectoryTreeBitmap
		bitmap = getBitmap(file, sb.PrDirectoryTreeBitmap, sizeBM)
	case "BM_DETDIR":
		fmt.Println("[BITMAP] Creando reporte de Bitmap de detalle de directorios.")
		sizeBM := sb.PrDirectoryDetail - sb.PrDirectoryDetailBitmap
		bitmap = getBitmap(file, sb.PrDirectoryDetailBitmap, sizeBM)
	case "BM_INODE":
		fmt.Println("[BITMAP] Creando reporte de Bitmap de inodos.")
		sizeBM := sb.PrInodeTable - sb.PrInodeTableBitmap
		bitmap = getBitmap(file, sb.PrInodeTableBitmap, sizeBM)
	case "BM_BLOCK":
		fmt.Println("[BITMAP] Creando reporte de Bitmap de bloques.")
		sizeBM := sb.PrBlocks - sb.PrBlocksBitmap
		bitmap = getBitmap(file, sb.PrBlocksBitmap, sizeBM)
	default:
		fmt.Println("[ERROR]: El tipo de reporte a crear no existe |", rep, "|.")
	}
	//Creamos el archivo que representa el reporte de bitmap
	createReportBM(path, bitmap)
	fmt.Println("[REPORT] El reporte ha sido generado con exito.")
}

func reportMBR(file *os.File, mbr masterBootRecord) string {
	//Variable que concatenara todas las sentencias en lenguaje DOT para crear el reporte con GRAPHVIZ
	var dot string = "digraph REP_MBR{\n"
	//Variable que almacenara temporalmente la posicion de la particion extendida
	indexExtended := -1
	//Informacion del MBR
	dot += "MBR[\n"
	dot += "shape=none;label=<<TABLE CELLSPACING='-1' CELLBORDER='1'>\n"
	dot += " \t<tr><td colspan=\"2\"><b>MBR "
	dot += file.Name()
	dot += "</b></td></tr>\n"
	dot += "<tr><td WIDTH='200'>NOMBRE</td><td WIDTH='300'>VALOR</td></tr>\n"
	dot += "<tr>  <td><b>mbr_tamaño</b></td><td>"
	dot += strconv.FormatInt(mbr.Size, 10)
	dot += " bytes</td>  </tr>\n"
	dot += "<tr>  <td><b>mbr_fecha_creacion</b></td> <td>"
	dot += string(mbr.CreatedAt[:])
	dot += "</td>  </tr>\n"
	dot += "<tr>  <td><b>mbr_disk_signature</b></td> <td>"
	dot += strconv.FormatInt(mbr.DiskSignature, 10)
	dot += "</td>  </tr>\n"
	//Informacion de las particiones
	for i, part := range mbr.Partitions {
		if part.Status != 0 {
			if part.Type == 'E' {
				indexExtended = i
			}
			dot += "<tr>  <td><b>part_status_"
			dot += strconv.Itoa(i + 1)
			dot += "</b></td> <td>"
			dot += "1"
			dot += "</td>  </tr>\n"
			dot += "<tr>  <td><b>part_type_"
			dot += strconv.Itoa(i + 1)
			dot += "</b></td> <td>"
			dot += string(part.Type)
			dot += "</td>  </tr>\n"
			dot += "<tr>  <td><b>part_fit_"
			dot += strconv.Itoa(i + 1)
			dot += "</b></td> <td>"
			dot += string(part.Fit)
			dot += "</td>  </tr>\n"
			dot += "<tr>  <td><b>part_start_"
			dot += strconv.Itoa(i + 1)
			dot += "</b></td> <td>"
			dot += strconv.FormatInt(part.Start, 10)
			dot += "</td>  </tr>\n"
			dot += "<tr>  <td><b>part_size_"
			dot += strconv.Itoa(i + 1)
			dot += "</b></td> <td>"
			dot += strconv.FormatInt(part.Size, 10)
			dot += "</td>  </tr>\n"
			dot += "<tr>  <td><b>part_name_"
			dot += strconv.Itoa(i + 1)
			dot += "</b></td> <td>"
			dot += strings.Replace(string(part.Name[:]), "\x00", "", -1)
			dot += "</td>  </tr>\n"
		}
	}
	dot += "</TABLE>\n>];\n"
	if indexExtended != -1 {
		//EBR aux
		ebr := extendedBootRecord{}
		indexEBR := mbr.Partitions[indexExtended].Start
		if indexExtended != -1 {
			for i := 1; true; i++ {
				file.Seek(indexEBR, 0)
				//Se obtiene la data del archivo binario
				data := readNextBytes(file, int64(binary.Size(ebr)))
				buffer := bytes.NewBuffer(data)
				err := binary.Read(buffer, binary.BigEndian, &ebr)
				if err != nil {
					log.Fatal("binary.Read failed", err)
				}
				if ebr.Status != 0 {
					//Informacion de los EBR's
					dot += "subgraph cluster_"
					dot += strconv.Itoa(i)
					dot += "{\n label=\"EBR_"
					dot += strconv.Itoa(i)
					dot += "\"\n"
					dot += "\ntbl_"
					dot += strconv.Itoa(i)
					dot += "[shape=box, label=<\n "
					dot += "<TABLE border='0' cellborder='1' cellspacing='0'  width='300' height='160' >\n "
					dot += "<tr>  <td width='150'><b>Nombre</b></td> <td width='150'><b>Valor</b></td>  </tr>\n"
					dot += "<tr>  <td><b>part_status</b></td> <td>"
					dot += "1"
					dot += "</td>  </tr>\n"
					dot += "<tr>  <td><b>part_fit</b></td> <td>"
					dot += string(ebr.Fit)
					dot += "</td>  </tr>\n"
					dot += "<tr>  <td><b>part_start</b></td> <td>"
					dot += strconv.FormatInt(ebr.Start, 10)
					dot += "</td>  </tr>\n"
					dot += "<tr>  <td><b>part_size</b></td> <td>"
					dot += strconv.FormatInt(ebr.Size, 10)
					dot += "</td>  </tr>\n"
					dot += "<tr>  <td><b>part_next</b></td> <td>"
					dot += strconv.FormatInt(ebr.Next, 10)
					dot += "</td>  </tr>\n"
					dot += "<tr>  <td><b>part_name</b></td> <td>"
					dot += strings.Replace(string(ebr.Name[:]), "\x00", "", -1)
					dot += "</td>  </tr>\n"
					dot += "</TABLE>\n>];\n}\n"
				}
				if ebr.Next == -1 {
					break
				}
				indexEBR = ebr.Next
			}
		}
	}
	dot += "}\n"
	return dot
}

func reportDisk(file *os.File, mbr masterBootRecord) string {
	//Variable que concatenara todas las sentencias en lenguaje DOT para crear el reporte con GRAPHVIZ
	var dot string = "digraph REP_DISK{\n"
	dot += "DISC[\nshape=box\nlabel=<\n"
	dot += "<table border='0' cellborder='2' width='500' height=\"180\">\n"
	dot += " \t<tr><td colspan=\"6\"><b>DISK "
	dot += file.Name()
	dot += "</b></td></tr>\n"
	dot += "<tr>\n"
	dot += "<td height='200' width='100'> MBR </td>\n"
	//Informacion de las particiones
	for k, part := range mbr.Partitions {
		if part.Status != 0 {
			//Estructura de una particion extendida
			if part.Type == 'E' {
				nLogics := 0
				dotAux := ""
				//EBR aux
				ebr := extendedBootRecord{}
				indexEBR := part.Start
				//Recorremos cada una de los ebr para agregar las particiones logicas
				for i := 1; true; i++ {
					nLogics++
					ebr = getEBR(file, indexEBR)
					//Porcentaje que ocupa esta particion logica
					percentage := float64(ebr.Size) * 100 / float64(mbr.Size)
					//Informacion de los EBR's
					if percentage != 0 {
						if ebr.Status != 0 {
							dotAux += "<td height='200' width='75'>EBR</td>\n"
							dotAux += "     <td height='200' width='150'>LOGICA<br/>"
							dotAux += strings.Replace(string(ebr.Name[:]), "\x00", "", -1)
							dotAux += "<br/> Porcentaje: "
							dotAux += fmt.Sprintf("%f", percentage)
							dotAux += "%</td>\n"
						} else {
							dotAux += "      <td height='200' width='150'>LIBRE <br/> Porcentaje: "
							dotAux += fmt.Sprintf("%f", percentage)
							dotAux += "%</td>\n"
						}
					}
					//Cuando encontramos a la ultima particion logica
					if ebr.Next == -1 {
						//Porcentaje libre luego de encontrar la ultima logica
						freeExtended := float64(part.Start + part.Size - (ebr.Start - int64(binary.Size(ebr))) - ebr.Size)
						percentage := freeExtended * 100 / float64(mbr.Size)
						if percentage != 0 {
							dotAux += " <td height='200' width='150'>LIBRE <br/> Porcentaje: "
							dotAux += fmt.Sprintf("%f", percentage)
							dotAux += "% </td>\n"
						}
						if i == 1 && ebr.Status == 0 {
							nLogics = 0
						}
						break
					} else {
						//Si existe algun espacio muerto entre dos particiones logicas
						pr := ebr.Start + ebr.Size
						nEBR := getEBR(file, ebr.Next)
						if pr != nEBR.Start {
							//Porcentaje libre entre ese espacio encontrado
							nLogics++
							freeExtended := nEBR.Start - pr
							percentage := float64(freeExtended) * 100 / float64(mbr.Size)
							if percentage != 0 {
								dotAux += " <td height='200' width='150'>LIBRE <br/> Porcentaje: "
								dotAux += fmt.Sprintf("%f", percentage)
								dotAux += "% </td>\n"
							}
						}
					}
					indexEBR = ebr.Next
				}
				dotAux += "     </tr>\n     </table>\n     </td>\n"
				//Encabezado de la extendida
				dot += "<td  height='200' width='100'>\n     <table border='0'  height='200' WIDTH='100' cellborder='1'>\n"
				dot += "<tr>  <td height='60' colspan='"
				dot += strconv.Itoa(nLogics*2 + 1)
				dot += "'>EXTENDIDA: "
				dot += strings.Replace(string(part.Name[:]), "\x00", "", -1)
				dot += "</td>  </tr>\n     <tr>\n"
				dot += dotAux
			} else { //Particiones Primarias
				dot += "<td height='200' width='200'>PRIMARIA <br/> "
				dot += strings.Replace(string(part.Name[:]), "\x00", "", -1)
				dot += "<br/> Utilizado: "
				//Porcentaje que ocupa esta particion primaria
				percentage := float64(part.Size) * 100 / float64(mbr.Size)
				dot += fmt.Sprintf("%f", percentage)
				dot += "%</td>\n"
				//Verificamos si existe fragmentacion
				currentPartition := mbr.Partitions[k].Start + mbr.Partitions[k].Size
				if k != 3 {
					nextPartition := mbr.Partitions[k+1].Start
					if mbr.Partitions[k+1].Status != 0 {
						fragment := nextPartition - currentPartition
						if fragment != 0 {
							percentage := float64(fragment) * 100 / float64(mbr.Size)
							dot += "<td height='200' width='"
							dot += strconv.FormatInt(int64(percentage)*5, 10)
							dot += "'>LIBRE <br/>"
							dot += fmt.Sprintf("%f", percentage)
							dot += "%</td>\n"
						}
					}
				} else {
					fragment := mbr.Size - currentPartition
					if fragment != 0 {
						percentage := float64(fragment) * 100 / float64(mbr.Size)
						dot += "<td height='200' width='"
						dot += strconv.FormatInt(int64(percentage)*5, 10)
						dot += "'>LIBRE <br/>"
						dot += fmt.Sprintf("%f", percentage)
						dot += "%</td>\n"
					}
				}
			}
		} else { //Si la particion esta libre
			dot += "<td height='200' width='200'>LIBRE <br/>"
			//Porcentaje que ocupa esta particion
			percentage := float64(part.Size) * 100 / float64(mbr.Size)
			dot += fmt.Sprintf("%f", percentage)
			dot += "%</td>\n"
		}

	}
	dot += "</tr> \n     </table>        \n>];\n\n}"
	return dot
}

func searchPartition(id string) (string, mountedParts, error) {
	//Se instancia un struct de particion montada
	mountedPartition := mountedParts{}
	//Se instancia un struct de un disco
	mountedDisk := mounted{}
	//Buscamos en la lista de discos montados
	for _, disk := range mountedDisks {
		for _, part := range disk.parts {
			if part.id == id {
				mountedDisk = disk
				mountedPartition = part
				break
			}
		}
	}
	if mountedDisk.path == "" {
		fmt.Println("[ERROR] El id no coincide con ninguna particon en la lista de particiones montadas.")
		return "", mountedPartition, fmt.Errorf("ERROR")
	}
	return mountedDisk.path, mountedPartition, nil
}

func reportSuperBoot(file *os.File, mountedPart mountedParts, superboot superBoot, name string) string {
	var dot string = "digraph REP_SB{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n\n"
	//Empezamos a escribir el reporte
	dot += "Node0 [label=<\n"
	dot += "<table border=\"0\" cellborder=\"1\" cellpadding=\"8\">\n"
	dot += "\t<tr><td colspan=\"2\">Superbloque "
	dot += mountedPart.id
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_nombre_hd</td><td>"
	dot += name
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_arbol_virtual_count</td><td>"
	dot += strconv.FormatInt(superboot.VirtualTreeCount, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_detalle_directorio_count</td><td>"
	dot += strconv.FormatInt(superboot.DirectoryDetailCount, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_inodos_count</td><td>"
	dot += strconv.FormatInt(superboot.InodesCount, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_bloques_count</td><td>"
	dot += strconv.FormatInt(superboot.BlocksCount, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_arbol_virtual_free</td><td>"
	dot += strconv.FormatInt(superboot.VirtualTreeFree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_detalle_directorio_free</td><td>"
	dot += strconv.FormatInt(superboot.DirectoryDetailFree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_inodos_free</td><td>"
	dot += strconv.FormatInt(superboot.InodesFree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_bloques_free</td><td>"
	dot += strconv.FormatInt(superboot.BlocksFree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_date_creacion</td><td>"
	dot += string(superboot.CreationDate[:])
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_date_ultimo_montaje</td><td>"
	dot += string(superboot.LastAssemblyDate[:])
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_montajes_count</td><td>"
	dot += strconv.FormatInt(superboot.MontageCount, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_bitmap_arbol_directorio</td><td>"
	dot += strconv.FormatInt(superboot.PrDirectoryTreeBitmap, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_arbol_directorio</td><td>"
	dot += strconv.FormatInt(superboot.PrDirectoryTree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_bitmap_detalle_directorio</td><td>"
	dot += strconv.FormatInt(superboot.PrDirectoryDetailBitmap, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_detalle_directorio</td><td>"
	dot += strconv.FormatInt(superboot.PrDirectoryDetail, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_bitmap_tabla_inodo</td><td>"
	dot += strconv.FormatInt(superboot.PrInodeTableBitmap, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_tabla_inodo</td><td>"
	dot += strconv.FormatInt(superboot.PrInodeTable, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_bitmap_bloques</td><td>"
	dot += strconv.FormatInt(superboot.PrBlocksBitmap, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_bloques</td><td>"
	dot += strconv.FormatInt(superboot.PrBlocks, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_ap_log</td><td>"
	dot += strconv.FormatInt(superboot.PrLog, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_size_struct_arbol_directorio</td><td>"
	dot += strconv.FormatInt(superboot.SizeDirectoryTree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_size_struct_detalle_directorio</td><td>"
	dot += strconv.FormatInt(superboot.SizeDirectoryDetail, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_size_struct_inodo</td><td>"
	dot += strconv.FormatInt(superboot.SizeInode, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_size_struct_bloque</td><td>"
	dot += strconv.FormatInt(superboot.SizeBlock, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_first_free_bit_arbol_directorio</td><td>"
	dot += strconv.FormatInt(superboot.FirstFreeBitDirectoryTree, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_first_free_bit_detalle_directorio</td><td>"
	dot += strconv.FormatInt(superboot.FirstFreeBitDirectoryDetail, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_first_free_bit_tabla_inodo</td><td>"
	dot += strconv.FormatInt(superboot.FirstFreeBitInodeTable, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_first_free_bit_bloques</td><td>"
	dot += strconv.FormatInt(superboot.FirstFreeBitBlocks, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td>sb_magic_num</td><td>"
	dot += string(superboot.MagicNum[:])
	dot += "</td></tr>\n"
	dot += "</table>\n>];\n"
	dot += "}"
	return dot
}

//Reporte de arbol virtual de directorio
func reportVirtualDirectoryTree(file *os.File, sb superBoot) string {
	var dot string = "digraph REP_VDT{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n ranksep = 2;\n\n"
	//Empezamos a escribir el reporte
	dot += vdtModel(file, sb, 0)
	dot += "}"
	return dot
}

//Modelo de un arbol de directorio virtual completo
func vdtModel(file *os.File, sb superBoot, index int64) string {
	vdt := getVirtualDirectotyTree(file, sb.PrDirectoryTree, index)
	dot := vdtTable(vdt, index)
	//Creamos los subdirecotrios
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] != -1 {
			dot += vdtModel(file, sb, vdt.Subdirectories[i])
			dot += "N"
			dot += strconv.FormatInt(index, 10)
			dot += ":"
			dot += strconv.Itoa(i + 1)
			dot += " -> N"
			dot += strconv.FormatInt(vdt.Subdirectories[i], 10)
			dot += ":0;\n"
		}
	}
	//Creamos el indirecto si es que existe
	if vdt.PrVirtualDirectoryTree != -1 {
		dot += vdtModel(file, sb, vdt.PrVirtualDirectoryTree)
		dot += "N"
		dot += strconv.FormatInt(index, 10)
		dot += ":8-> N"
		dot += strconv.FormatInt(vdt.PrVirtualDirectoryTree, 10)
		dot += ":0;\n"
	}
	return dot
}

//Reporte completo de todo los directorios con su detalle y archivos
func reportTreeComplete(file *os.File, sb superBoot) string {
	var dot string = "digraph REP_VDT{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n ranksep = 2;\n\n"
	//Empezamos a escribir el reporte
	dot += completeModel(file, sb, 0)
	dot += "}"
	return dot
}

//Modelo completo de todo los directorios con su detalle y archivos
func completeModel(file *os.File, sb superBoot, index int64) string {
	vdt := getVirtualDirectotyTree(file, sb.PrDirectoryTree, index)
	dot := vdtTable(vdt, index)
	//Creamos los subdirecotrios
	for i := 0; i < len(vdt.Subdirectories); i++ {
		if vdt.Subdirectories[i] != -1 {
			dot += completeModel(file, sb, vdt.Subdirectories[i])
			dot += "N"
			dot += strconv.FormatInt(index, 10)
			dot += ":"
			dot += strconv.Itoa(i + 1)
			dot += " -> N"
			dot += strconv.FormatInt(vdt.Subdirectories[i], 10)
			dot += ":0;\n"
		}
	}
	//Creamos el indirecto si es que existe
	if vdt.PrVirtualDirectoryTree != -1 {
		dot += vdtModel(file, sb, vdt.PrVirtualDirectoryTree)
		dot += "N"
		dot += strconv.FormatInt(index, 10)
		dot += ":8-> N"
		dot += strconv.FormatInt(vdt.PrVirtualDirectoryTree, 10)
		dot += ":0;\n"
	}
	//Creamos los detalles de directorios
	dot += ddModel(file, sb, index, vdt.DirectoryName)
	//Asignamos el detalle de directorio
	dot += "N"
	dot += strconv.FormatInt(index, 10)
	dot += ":7"
	dot += " -> D"
	dot += strconv.FormatInt(vdt.PrDirectoryDetail, 10)
	dot += ":0;\n"
	//Construimos los inodos
	dot += inodesModel(file, sb, vdt.PrDirectoryDetail)
	return dot
}

//Reporte de arbol de directorio que muestra su detalle del directorio
func reportDirectoryTree(file *os.File, sb superBoot, path string) (string, error) {
	//Revismos que la ruta a insertar sea correcta
	if path[0] != '/' {
		fmt.Println("[ERROR] El path no es valido.")
		return "", fmt.Errorf("ERROR")
	}
	//Obtenemos las carpetas
	folders := strings.Split(path, "/")
	folders = folders[1:]
	var dot string = "digraph REP_TREEDIRECTORY{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n ranksep = 2;\n\n"
	//Recuperamos el arbol de directorio de '/'
	root := getVirtualDirectotyTree(file, sb.PrDirectoryTree, 0)
	//Empezamos a escribir el reporte
	aux, _ := buildDirectoryTree(file, sb, root, folders)
	dot += aux
	dot += "}"
	return dot, nil
}

//Funcion que construye el arbol de directorio con su detalle de directorio
func buildDirectoryTree(file *os.File, sb superBoot, root virtualDirectoryTree, folders []string) (string, int64) {
	var index int64
	var aux string
	var foldername [20]byte
	var pr int64
	//Obtenemos la raiz
	dot := vdtTable(root, 0)
	//Si existe mas subdirectorios realizamos la iteracion o en dado caso es el detalle de la raiz '/'
	if len(folders) > 0 && folders[0] != "" {
		index, aux, foldername, pr = directoryTreeModel(file, &sb, root, folders, 0)
	} else {
		index, aux, foldername, pr = root.PrDirectoryDetail, "", root.DirectoryName, 0
	}
	//Obtenido el indice del detalle de directorio procedemos a crear su grafico
	if index != -1 {
		//Se incluyen los subdirectorios, si es que lo habian
		dot += aux
		//Se crea el grafico de la estructura
		dot += ddModel(file, sb, index, foldername)
		//Asignamos el detalle de directorio
		dot += "N"
		dot += strconv.FormatInt(pr, 10)
		dot += ":7"
		dot += " -> D"
		dot += strconv.FormatInt(index, 10)
		dot += ":0;\n"
	}
	return dot, index
}

//Funcion que te devuelve una tabla que representa la estructura de un vdt
func vdtTable(vdt virtualDirectoryTree, bm int64) string {
	dot := "N"
	dot += strconv.FormatInt(bm, 10)
	dot += "[color=\"#99ccff\"  label=<\n"
	dot += "<table border=\"0\" cellborder=\"1\" cellpadding=\"10\">\n"
	dot += "\t<tr><td bgcolor=\"#99ccff\" colspan=\"2\" PORT=\"0\">"
	dot += strings.Replace(string(vdt.DirectoryName[:]), "\x00", "", -1)
	dot += "</td></tr>\n"
	//Subdirectorios
	for i := 0; i < len(vdt.Subdirectories); i++ {
		dot += "\t<tr><td>aptr"
		dot += strconv.Itoa(i + 1)
		dot += "</td><td PORT=\""
		dot += strconv.Itoa(i + 1)
		dot += "\">"
		dot += strconv.FormatInt(vdt.Subdirectories[i], 10)
		dot += "</td></tr>\n"
	}
	//Fecha de creacion
	dot += "\t<tr><td>Creacion</td><td>"
	dot += string(vdt.CreatedAt[:])
	dot += "</td></tr>\n"
	//Detalle de directorio
	dot += "\t<tr><td bgcolor=\"#7ab648\">detalle_D</td><td PORT=\"7\">"
	dot += strconv.FormatInt(vdt.PrDirectoryDetail, 10)
	dot += "</td></tr>\n"
	//Apuntador indirecto
	dot += "\t<tr><td bgcolor=\"#99ccff\">aptr_ind</td><td PORT=\"8\">"
	dot += strconv.FormatInt(vdt.PrVirtualDirectoryTree, 10)
	dot += "</td></tr>\n"
	dot += "</table>\n>];\n"
	return dot
}

//Funcion del modelo treeFile
func directoryTreeModel(file *os.File, sb *superBoot, vdt virtualDirectoryTree, folders []string, bm int64) (int64, string, [20]byte, int64) {
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
		dot := vdtTable(aux, index)
		for i := 0; i < len(vdt.Subdirectories); i++ {
			if vdt.Subdirectories[i] != -1 && vdt.Subdirectories[i] == index {
				dot += "N"
				dot += strconv.FormatInt(bm, 10)
				dot += ":"
				dot += strconv.Itoa(i + 1)
				dot += " -> N"
				dot += strconv.FormatInt(vdt.Subdirectories[i], 10)
				dot += ":0;\n"
			}
		}
		if len(folders) > 0 {
			j, g, a, inx := directoryTreeModel(file, sb, aux, folders, index)
			dot += g
			return j, dot, a, inx
		}
		return aux.PrDirectoryDetail, dot, auxVDT, index
	}
	return -1, "", [20]byte{}, -1
}

//Funcion que genera la estructura de todo un detalle de directorio de una directorio
func ddModel(file *os.File, sb superBoot, index int64, foldername [20]byte) string {
	dot := "D"
	dot += strconv.FormatInt(index, 10)
	dot += "[color=\"#7ab648\"  label=<\n"
	dot += "<table border=\"0\" cellborder=\"1\" cellpadding=\"10\">\n"
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, index)
	dot += "\t<tr><td bgcolor=\"#7ab648\" colspan=\"2\" PORT=\"0\">"
	dot += "DD: " + strings.Replace(string(foldername[:]), "\x00", "", -1)
	dot += "</td></tr>\n"
	//Archivos contenidos en el directo
	for i := 0; i < len(dd.Files); i++ {
		dot += "\t<tr><td>"
		name := strings.Replace(string(dd.Files[i].Name[:]), "\x00", "", -1)
		if name != "" {
			dot += name
		} else {
			dot += "----------"
		}
		dot += "</td><td bgcolor=\"#fcc438\" PORT=\""
		dot += strconv.Itoa(i + 1)
		dot += "\">"
		dot += strconv.FormatInt(dd.Files[i].PrInode, 10)
		dot += "</td></tr>\n"
	}
	//Apuntador indirecto
	dot += "\t<tr><td bgcolor=\"#7ab648\">aptr_ind</td><td  PORT=\"6\">"
	dot += strconv.FormatInt(dd.PrDirectoryDetail, 10)
	dot += "</td></tr>\n"
	dot += "</table>\n>];\n"
	//Creamos el indirecto si es que existe
	if dd.PrDirectoryDetail != -1 {
		dot += ddModel(file, sb, dd.PrDirectoryDetail, foldername)
		dot += "D"
		dot += strconv.FormatInt(index, 10)
		dot += ":6-> D"
		dot += strconv.FormatInt(dd.PrDirectoryDetail, 10)
		dot += ":0;\n"
	}
	return dot
}

//Reporte del arbol de directorio donde este contenido en su detalle cierto archivo y que muestra los inodos que lo representan y su contenido en bloques
func reportTreeFile(file *os.File, sb superBoot, path string) (string, error) {
	//Revismos que la ruta a insertar sea correcta
	if path[0] != '/' {
		fmt.Println("[ERROR] El path no es valido.")
		return "", fmt.Errorf("ERROR")
	}
	//Obtenemos las carpetas
	folders := strings.Split(path, "/")
	folders = folders[1:]
	//Obtenemos el nombre del archivo a representar
	if !strings.HasSuffix(strings.ToLower(folders[len(folders)-1]), ".txt") {
		fmt.Println("[ERROR] El file a buscar no es valido.")
		return "", fmt.Errorf("ERROR")
	}
	var filename [20]byte
	copy(filename[:], folders[len(folders)-1])
	folders = folders[:len(folders)-1]
	//Empezamos a escribir el reporte
	var dot string = "digraph REP_TREEFILE{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n ranksep = 2;\n\n"
	//Recuperamos el arbol de directorio de '/'
	root := getVirtualDirectotyTree(file, sb.PrDirectoryTree, 0)
	//Construimos el arbol de directorio y su detalle
	aux, bm := buildDirectoryTree(file, sb, root, folders)
	dot += aux
	//[-] Construimos los inodos
	//Obtenemos el detalle de directorio
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, bm)
	//Recuperamos el puntero del inodo donde se encuentra el archivo
	nInode, nDD := searchFile(file, sb, dd, filename)
	if nInode != -1 {
		dot += inodeModel(file, sb, nInode)
		//Asignamos el inodo al detalle de directoio
		dot += "D"
		dot += strconv.FormatInt(bm, 10)
		dot += ":"
		dot += strconv.Itoa(nDD + 1)
		dot += " -> I"
		dot += strconv.FormatInt(nInode, 10)
		dot += ":0;\n"
	} else {
		fmt.Println("[ERROR] El archivo no fue encontrado.")
		return "", fmt.Errorf("ERROR")
	}
	dot += "}"
	return dot, nil
}

//Funcion que genera la estructura de todos los inodos que conforman un archivo
func inodeModel(file *os.File, sb superBoot, index int64) string {
	dot := "I"
	dot += strconv.FormatInt(index, 10)
	dot += "[color=\"#fcc438\"  label=<\n"
	dot += "<table border=\"0\" cellborder=\"1\" cellpadding=\"10\">\n"
	inode := getInode(file, sb.PrInodeTable, index)
	dot += "\t<tr><td bgcolor=\"#fcc438\" colspan=\"2\" PORT=\"0\">"
	dot += "INODO: " + strconv.FormatInt(inode.Count, 10)
	dot += "</td></tr>\n"
	//Bloques del inodo
	for i := 0; i < len(inode.Blocks); i++ {
		dot += "\t<tr><td>aptr"
		dot += strconv.Itoa(i + 1)
		dot += "</td><td bgcolor=\"#ffbbb1\" PORT=\""
		dot += strconv.Itoa(i + 1)
		dot += "\">"
		dot += strconv.FormatInt(inode.Blocks[i], 10)
		dot += "</td></tr>\n"
	}
	//Tamaño del archivo
	dot += "\t<tr><td>Tamaño</td><td  PORT=\"5\">"
	dot += strconv.FormatInt(inode.SizeFile, 10)
	dot += "</td></tr>\n"
	//Cantidad de bloques
	dot += "\t<tr><td>Bloques</td><td  PORT=\"6\">"
	dot += strconv.FormatInt(inode.AllocatedBlock, 10)
	dot += "</td></tr>\n"
	//Apuntador indirecto
	dot += "\t<tr><td bgcolor=\"#fcc438\">aptr_ind</td><td  PORT=\"7\">"
	dot += strconv.FormatInt(inode.PrIndirect, 10)
	dot += "</td></tr>\n"
	dot += "</table>\n>];\n"
	//Creamos los bloques de datos
	for i := 0; i < len(inode.Blocks); i++ {
		if inode.Blocks[i] != -1 {
			dot += blockModel(file, sb, inode.Blocks[i])
			dot += "I"
			dot += strconv.FormatInt(inode.Count, 10)
			dot += ":"
			dot += strconv.Itoa(i + 1)
			dot += " -> B"
			dot += strconv.FormatInt(inode.Blocks[i], 10)
			dot += ":0;\n"
		}
	}
	//Creamos el indirecto si es que existe
	if inode.PrIndirect != -1 {
		dot += inodeModel(file, sb, inode.PrIndirect)
		dot += "I"
		dot += strconv.FormatInt(index, 10)
		dot += ":7-> I"
		dot += strconv.FormatInt(inode.PrIndirect, 10)
		dot += ":0;\n"
	}
	return dot
}

//Funcion que genera multiples inodos de varios archivos
func inodesModel(file *os.File, sb superBoot, index int64) string {
	//Obtenemos el detalle de directorio
	dd := getDirectotyDetail(file, sb.PrDirectoryDetail, index)
	//Recuperamos el puntero del inodo donde se encuentra el archivo
	dot := ""
	for i := 0; i < len(dd.Files); i++ {
		nInode, nDD := searchFile(file, sb, dd, dd.Files[i].Name)
		if nInode != -1 {
			dot += inodeModel(file, sb, nInode)
			//Asignamos el inodo al detalle de directoio
			dot += "D"
			dot += strconv.FormatInt(index, 10)
			dot += ":"
			dot += strconv.Itoa(nDD + 1)
			dot += " -> I"
			dot += strconv.FormatInt(nInode, 10)
			dot += ":0;\n"
		}
	}
	if dd.PrDirectoryDetail != -1 {
		dot += inodesModel(file, sb, dd.PrDirectoryDetail)
	}
	return dot
}

//Funcion que genera la estructura de un bloque de datos de un inodo
func blockModel(file *os.File, sb superBoot, index int64) string {
	dot := "B"
	dot += strconv.FormatInt(index, 10)
	dot += "[color=\"#ffbbb1\"  label=<\n"
	dot += "<table border=\"0\" cellborder=\"1\" cellpadding=\"10\">\n"
	block := getBlock(file, sb.PrBlocks, index)
	//Numero de bloque
	dot += "\t<tr><td bgcolor=\"#ffbbb1\" PORT=\"0\">Bloque</td><td bgcolor=\"#ffbbb1\">"
	dot += strconv.FormatInt(index, 10)
	dot += "</td></tr>\n"
	dot += "\t<tr><td colspan=\"2\" PORT=\"1\">"
	dot += strings.Replace(string(block.Data[:]), "\x00", "", -1)
	dot += "</td></tr>\n"
	dot += "</table>\n>];\n"
	return dot
}

//Reporte de la bitacora del sistema de archivos
func reportLog(file *os.File, sb superBoot) string {
	//Empezamos a escribir el reporte
	var dot string = "digraph REP_TREEFILE{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n ranksep = 2;\n\n"
	//Obtenemos la bitacora inicial
	bita, _ := getBitacora(file, sb.PrLog, 0)
	for i := 1; bita.TransactionDate != [19]byte{}; i++ {
		dot += logModel(bita, int64(i))
		//Obtenemos la siguiente bitacora
		bita, _ = getBitacora(file, sb.PrLog, int64(i))
	}
	dot += "}"
	return dot
}

//Funcion que genera la estructura de un log de cambios de la bitacora
func logModel(bita bitacora, index int64) string {
	dot := "L"
	dot += strconv.FormatInt(index, 10)
	dot += "[color=\"#ffbbb1\"  label=<\n"
	dot += "<table border=\"0\" cellborder=\"1\" cellpadding=\"10\">\n"
	//Numero de log
	dot += "\t<tr><td bgcolor=\"#ffbbb1\" colspan=\"6\">Log "
	dot += strconv.FormatInt(index, 10)
	dot += "</td></tr>\n"
	//Atributos
	dot += "\t<tr><td>Operacion</td><td>Tipo</td><td>Path</td><td>Contenido</td><td>Fecha log</td><td>Size</td></tr>\n"
	//Contenido de la bitacora
	dot += "\t<tr>"
	dot += "<td>"
	dot += strings.Replace(string(bita.Operation[:]), "\x00", "", -1)
	dot += "</td>"
	dot += "<td>"
	dot += string(bita.Type)
	dot += "</td>"
	dot += "<td>"
	dot += strings.Replace(string(bita.Name[:]), "\x00", "", -1)
	dot += "</td>"
	dot += "<td>"
	dot += strings.Replace(string(bita.Content[:]), "\x00", "", -1)
	dot += "</td>"
	dot += "<td>"
	dot += string(bita.TransactionDate[:])
	dot += "</td>"
	dot += "<td>"
	dot += strconv.FormatInt(bita.Size, 10)
	dot += "</td>"
	dot += "</tr>\n"
	dot += "</table>\n>];\n"
	return dot
}

//Funcion que recorre todo el detalle de directorio para encontrar un archivo
func searchFile(file *os.File, sb superBoot, dd directoryDetail, filename [20]byte) (int64, int) {
	for i := 0; i < len(dd.Files); i++ {
		if dd.Files[i].Name == filename {
			return dd.Files[i].PrInode, i
		}
	}
	if dd.PrDirectoryDetail != -1 {
		aux := getDirectotyDetail(file, sb.PrDirectoryDetail, dd.PrDirectoryDetail)
		return searchFile(file, sb, aux, filename)
	}
	return -1, -1
}

//Funcion para crear reportes de bitmaps
func createReportBM(path string, bitmap []byte) {
	//Creamos el archivo
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data := ""
	//Creamos la data que representara el BITMAP
	for i, bm := range bitmap {
		if i%20 == 0 {
			data += "\n|"
		}
		if bm == '1' {
			data += "1|"
		} else {
			data += "0|"
		}
	}
	// Escribir la data del bm en el archivo
	err = ioutil.WriteFile(path, []byte(data), 0644)
	if err != nil {
		panic(err)
	}
}
