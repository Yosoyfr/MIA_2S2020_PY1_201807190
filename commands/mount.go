package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
)

//Estructura de un objeto mount que representara la particon montada Primaria
type primaryOrExtedendedPart struct {
	partition partition
	number    int
	id        string
}

//Estructura de un objeto mount que representara la particon montada Logica
type logicPart struct {
	partition extendedBootRecord
	number    int
	id        string
}

//Estructura de un objeto mount que representara todas las particiones de un disco que sean montadas
type mounted struct {
	path   string
	letter byte
	parts  []primaryOrExtedendedPart
	logics []logicPart
}

//Lista de los discos que han sido montados con sus respectivos
var discsMounted []mounted

//Funcion Mount, añade particiones que fueron montadas por el comando
func Mount(path string, name string) {
	//Obtenemos el mbr del disco
	file, mbr, err := readFile(path)
	if err != nil {
		return
	}
	defer file.Close()
	//Indice de la particion encontrada
	index := -1
	//Letra que se asigna por disco
	var discLetter byte = 'a'
	//Disco montando anteriormente para usarlo temporalmente
	disc := mounted{}
	//Buscamos en la lista de discos que fueron montados
	for i, discs := range discsMounted {
		if path == discs.path {
			//Se encontro el disco que alguna vez fue montado
			discLetter = discs.letter
			disc = discs
			index = i
			break
		}
		discLetter = byte(int(discLetter) + 1)
	}
	//	Obtenemos el nombre a asignar
	var realName [16]byte
	copy(realName[:], name)
	//Verificamos que no haya sido montada esta particion
	if existPart(disc, realName) {
		fmt.Println("Alert: Esta particion ya ha sido montada")
		return
	}
	//Variable que almacena el estado si se encontro o no la particion
	findPart := fmt.Errorf("NOT FOUND")
	findEBR := fmt.Errorf("NOT FOUND")
	//Variable que almacena temporalmente la particion encontrada
	partition := partition{}
	ebr := extendedBootRecord{}
	//Recorremos la lista de particiones del MBR
	//fmt.Println(string(discLetter))
	for _, part := range mbr.Partitions {
		if part.Name == realName {
			//Se encontro la particion
			findPart = nil
			partition = part
			break
		}
		//Buscamos en las particiones logicas
		if part.Type == 'E' {
			indexEBR := part.Start
			for i := 1; true; i++ {
				file.Seek(indexEBR, 0)
				//Se obtiene la data del archivo binario
				data := readNextBytes(file, int64(binary.Size(ebr)))
				buffer := bytes.NewBuffer(data)
				err := binary.Read(buffer, binary.BigEndian, &ebr)
				if err != nil {
					log.Fatal("binary.Read failed", err)
				}
				//Verificamos si existe
				if ebr.Name == realName {
					findPart = nil
					findEBR = nil
					break
				}
				//Si ya no hay siguientes
				if ebr.Next == -1 {
					break
				}
				indexEBR = ebr.Next
			}
		}
	}
	//En dado caso no se encuentre la particion
	if findPart != nil {
		fmt.Println("[ERROR] Este particion no fue encontrada en el disco.")
		return
	}
	//Si el disco no ha sido montado lo montamos
	if disc.path == "" {
		//fmt.Println("No ha sido montado")
		disc.letter = discLetter
		disc.path = path
	}
	//En dado caso vamos a montar una particion logica
	if findEBR == nil {
		part := logicPart{partition: ebr}
		if len(disc.logics) == 0 {
			part.number = 1
		} else {
			part.number = disc.logics[len(disc.logics)-1].number + 1
		}
		part.id = "vd" + string(disc.letter) + strconv.Itoa(part.number)
		disc.logics = append(disc.logics, part)
	} else {
		fmt.Println(partition, ebr)
	}
	if index == -1 {
		discsMounted = append(discsMounted, disc)
	} else {
		discsMounted[index] = disc
	}
	fmt.Println(discsMounted)
}

func existPart(disc mounted, name [16]byte) bool {
	//Revisamos si no existe en las primarias o extendidas
	for _, part := range disc.parts {
		if part.partition.Name == name {
			return true
		}
	}
	//Revisamos si no existe en las logicas
	for _, part := range disc.logics {
		if part.partition.Name == name {
			return true
		}
	}
	return false
}
