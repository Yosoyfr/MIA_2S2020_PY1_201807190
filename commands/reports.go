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
func Reports(path string, rep string, ext string, destiny string) {
	var report string
	switch rep {
	case "mbr":
		report = reportMBR(path)
	case "disc":
		report = reportDisk(path)
	}
	err := ioutil.WriteFile("report.dot", []byte(report), 0644)
	if err != nil {
		log.Fatal(err)
	}
	/*
		cmd, _ := exec.Command("dot", "-T"+ext, "report.dot", ">", destiny).Output()
		ioutil.WriteFile(destiny, cmd, os.FileMode(0777))
	*/
	exec.Command("dot", "-Tpng", "report.dot", "-o", "reporte.png").Output()
}

func reportMBR(path string) string {
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

func reportDisk(path string) string {
	//Obtenemos el mbr del disco
	file, mbr, err := readFile(path)
	if err != nil {
		return ""
	}
	//Variable que concatenara todas las sentencias en lenguaje DOT para crear el reporte con GRAPHVIZ
	var dot string = "digraph REP_DISK{\n"
	dot += "DISC[\nshape=box\nlabel=<\n"
	dot += "<table border='0' cellborder='2' width='500' height=\"180\">\n"
	dot += " \t<tr><td colspan=\"5\"><b>DISK "
	dot += file.Name()
	dot += "</b></td></tr>\n"
	dot += "<tr>\n"
	dot += "<td height='200' width='100'> MBR </td>\n"
	//Informacion de las particiones
	for _, part := range mbr.Partitions {
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
			}
		} else { //Si la particion esta libre
			dot += "<td height='200' width='200'>LIBRE <br/> Utilizado: "
			//Porcentaje que ocupa esta particion primaria
			percentage := float64(part.Size) * 100 / float64(mbr.Size)
			dot += fmt.Sprintf("%f", percentage)
			dot += "%</td>\n"
		}

	}
	dot += "</tr> \n     </table>        \n>];\n\n}"
	return dot
}
