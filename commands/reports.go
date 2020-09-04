package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

//Funcion manager del tipo de reporte a crear
func Reports(id string, rep string, path string) {
	var report string
	switch rep {
	case "mbr":
		report = reportMBR(id)
	case "disk":
		report = reportDisk(id)
	case "sb":
		report = reportSuperBoot(id)
	}
	err := ioutil.WriteFile("report.dot", []byte(report), 0644)
	if err != nil {
		log.Fatal(err)
	}
	extension := path[(len(path) - 3):(len(path))]
	exec.Command("dot", "-T"+extension, "report.dot", "-o", path).Output()
}

func reportMBR(id string) string {
	//Obtenemos el file y la particion a trabajar
	path, _, err := searchPartition(id)
	if err != nil {
		return ""
	}
	//Obtenemos el mbr del disco
	file, mbr, err := readFile(path)
	if err != nil {
		return ""
	}
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
			if ebr.Next == -1 {
				break
			}
			indexEBR = ebr.Next
		}
	}
	dot += "}\n"
	return dot
}

func reportDisk(id string) string {
	//Obtenemos el file y la particion a trabajar
	path, _, err := searchPartition(id)
	if err != nil {
		return ""
	}
	//Obtenemos el mbr del disco
	file, mbr, err := readFile(path)
	if err != nil {
		return ""
	}
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
					nLogics = i
					file.Seek(indexEBR, 0)
					//Se obtiene la data del archivo binario
					data := readNextBytes(file, int64(binary.Size(ebr)))
					buffer := bytes.NewBuffer(data)
					err := binary.Read(buffer, binary.BigEndian, &ebr)
					if err != nil {
						log.Fatal("binary.Read failed", err)
					}
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

func reportSuperBoot(id string) string {
	var dot string = "digraph REP_SB{\nrankdir = LR;\n node [shape=plain, fontsize=20];\n graph[dpi=120];\n\n"
	//Superboot a trabajar
	superboot := superBoot{}
	//Obtenemos el file y la particion a trabajar
	diskPath, mountedPart, err := searchPartition(id)
	if err != nil {
		return ""
	}
	file, _, err := readFile(diskPath)
	defer file.Close()
	if err != nil {
		return ""
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
	//Nos posicionamos en esa parte del archivo
	file.Seek(indexSB, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, int64(binary.Size(superboot)))
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err = binary.Read(buffer, binary.BigEndian, &superboot)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
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
