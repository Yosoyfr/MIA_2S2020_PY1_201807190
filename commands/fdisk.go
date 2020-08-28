package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

//Struct del EBR
type extendedBootRecord struct {
	Status byte
	Fit    byte
	Start  int64
	Size   int64
	Next   int64
	Name   [16]byte
}

//Struct de una particion del MBR
type partition struct {
	Status byte
	Type   byte
	Fit    byte
	Start  int64
	Size   int64
	Name   [16]byte
}

//Funcion para leer el archivo binario que representa el disco
func readFile(disc string) (*os.File, masterBootRecord) {
	//Se abre el archivo
	file, err := os.OpenFile(disc, os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	//Se instancia un struct de mbr
	mbr := masterBootRecord{}
	var size int64 = int64(binary.Size(mbr))
	file.Seek(0, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err = binary.Read(buffer, binary.BigEndian, &mbr)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	return file, mbr
}

func readNextBytes(file *os.File, number int64) []byte {
	bytes := make([]byte, number)
	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

/*
Funcion para crear particiones ya se de tipo_
P: Particion Primaria
E: Particion Extendida
L: Particoin Logica
*/
func FKDisk(path string, size int64, unit byte, typeF byte, fit byte, name string) {
	//Obtenemos el mbr del disco
	file, mbr := readFile(path)
	defer file.Close()
	/*
		[VALIDACIONES]
		Al obtener el mbr, realizamos los analisis:
		1.El tamaño de la particion tiene que ser menor o igual al tamaño disponible del disco
		2.No exista mas de una particion extendida
		3.No se pueden crear mas de 4 particiones contando [1 extendida y 3 primarias o 4 primarias y pueden existir N logicas dentro de una extendida] (Disponibilidad de particiones)
		4.Si se desea crear una particion logica, debe existir una particion extendida
		5.Debe existir un IDENTIFICADOR unico para cada particion
	*/
	//	Obtenemos el nombre a asignar
	var newName [16]byte
	copy(newName[:], name)
	// Obtenemos el tamaño disponible
	sizeAvailable := mbr.Size - int64(binary.Size(mbr))
	// Variable que identifica si existe una particion extendida en el disco
	existExtended := false
	//Variable que almacenara temporalmente la posicion de la particion extendida
	auxExtended := -1
	for i, part := range mbr.Partitions {
		//Espacio disponible
		if part.Status == 1 {
			sizeAvailable = sizeAvailable - part.Size
		}
		//Existe una particion extendida
		if part.Type == 'E' {
			auxExtended = i
			existExtended = true
		}
		//[VALIDACION 5]
		if part.Name == newName {
			fmt.Println("[ERROR] Este nombre ya fue asignado a una particion.")
			return
		}
	}
	//Obtenemos el tamaño de la particion a crear
	partitionSize, err := unitCalc(size, unit)
	if err != nil {
		fmt.Println("[ERROR] la unidad declarada no es valida.")
		return
	}
	//[CREACION DE LOGICAS]
	if typeF == 'L' {
		//[VALIDACION 4]
		if !existExtended {
			fmt.Println("[ERROR] No existe una particion extendida creada.")
			return
		}
		//[LOGICA] Proceso para la creacion de la  particion
		extended := mbr.Partitions[auxExtended]
		//EBR aux
		prevEBR := extendedBootRecord{}
		//[VALIDACION] El tamaño de la particion debe ser mayor al del EBR
		if partitionSize <= int64(binary.Size(prevEBR)) {
			fmt.Println("[ERROR] El espacio solicitado no es valido para crear una particion lógica.")
			return
		}
		//Obtenemos la posicion en la que empieza la particion extendida
		index := extended.Start
		//Nos ubicamos en la particion extendida en el disco
		file.Seek(index, 0)
		//Se obtiene la data del archivo binarios
		data := readNextBytes(file, 42)
		buffer := bytes.NewBuffer(data)
		//Se asigna al ebr declarado para leer la informacion de la particion extendida
		err = binary.Read(buffer, binary.BigEndian, &prevEBR)
		if err != nil {
			log.Fatal("binary.Read failed", err)
		}
		//Espacio disponible en la particion extendida
		sizeAvailableExteded := extended.Size
		fmt.Println(sizeAvailableExteded)

		//Si en dado caso es la inicial, seteamos los valores
		if prevEBR.Status == 0 {
			sizeAvailableExteded = sizeAvailableExteded - int64(binary.Size(prevEBR))
			//[VALIDACION] Espacio disponible en las extendidas
			if partitionSize > sizeAvailableExteded {
				fmt.Println("[ERROR] El espacio requerido no se encuentra disponible en la particion extendida.")
				return
			}
			prevEBR.Fit = fit
			prevEBR.Name = newName
			prevEBR.Size = partitionSize
			prevEBR.Status = 1
			prevEBR.Start = index + int64(binary.Size(prevEBR))
			//Empezamos el proceso de guardar la data del struct EBR
			file.Seek(index, 0)
			var binaryDisc bytes.Buffer
			binary.Write(&binaryDisc, binary.BigEndian, &prevEBR)
			writeNextBytes(file, binaryDisc.Bytes())

		} else {
			//Buscamos el ultimo insertado para editarlo
			for prevEBR.Next != -1 {
				//[VALIDACION 5]
				if prevEBR.Name == newName {
					fmt.Println("[ERROR] Este nombre ya fue asignado a una particion.")
					return
				}
				//Obtenemos el espacio disponible en la particion extendida
				sizeAvailableExteded = sizeAvailableExteded - prevEBR.Size
				//Nos ubicamos en la posicion donde empieza el siguiente ebr
				file.Seek(prevEBR.Next, 0)
				//Se obtiene la data del archivo binarios
				data = readNextBytes(file, int64(binary.Size(prevEBR)))
				buffer = bytes.NewBuffer(data)
				//Se asigna al ebr declarado para leer la informacion de la particion extendida
				err = binary.Read(buffer, binary.BigEndian, &prevEBR)
				if err != nil {
					log.Fatal("binary.Read failed", err)
				}
			}
			//[VALIDACION 5]
			if prevEBR.Name == newName {
				fmt.Println("[ERROR] Este nombre ya fue asignado a una particion.")
				return
			}
			//Obtenemos el espacio disponible en la particion extendida
			sizeAvailableExteded = sizeAvailableExteded - prevEBR.Size
			//[VALIDACION] Espacio disponible en las extendidas
			if partitionSize > sizeAvailableExteded {
				fmt.Println("[ERROR] El espacio requerido no se encuentra disponible en la particion extendida.")
				return
			}
			//Se debe modificar el siguiente del anterior
			prevEBR.Next = prevEBR.Start + (prevEBR.Size - int64(binary.Size(prevEBR)))
			//Se crea nuevo EBR que sera el siguiente
			nextEBR := extendedBootRecord{Fit: fit, Name: newName, Size: partitionSize, Status: 1, Next: -1, Start: prevEBR.Next + int64(binary.Size(prevEBR))}
			//Empezamos el proceso de guardar la data de la lista de EBR's
			file.Seek(prevEBR.Start-int64(binary.Size(prevEBR)), 0)
			var binaryPrevEBR bytes.Buffer
			binary.Write(&binaryPrevEBR, binary.BigEndian, &prevEBR)
			writeNextBytes(file, binaryPrevEBR.Bytes())
			file.Seek(prevEBR.Next, 0)
			var binaryNextEBR bytes.Buffer
			binary.Write(&binaryNextEBR, binary.BigEndian, &nextEBR)
			writeNextBytes(file, binaryNextEBR.Bytes())
		}
		return
	}
	//[VALIDACION 1]
	if partitionSize > sizeAvailable {
		fmt.Println("[ERROR] El espacio requerido no se encuentra disponible.")
		return
	}
	//[VALIDACION 2]
	if typeF == 'E' && existExtended {
		fmt.Println("[ERROR] Una particion extendida ya fue creada en el disco.")
		return
	}
	//[CREAR PARTICION EXTENDIDA O PRIMARIA]
	//Variable que guardara si existe alguna particion disponible
	index := -1
	//[VALIDACION 3]
	for i, part := range mbr.Partitions {
		if part.Status == 0 {
			index = i
			break
		}
	}
	//[VALIDACION 3]
	if index == -1 {
		fmt.Println("[ERROR] No existe una particion disponible en el disco.")
		return
	}
	//Creamos la nueva particion
	newPartition := partition{Size: partitionSize, Type: typeF, Status: 1, Fit: fit}
	newPartition.Name = newName
	//[START] Byte donde empieza la particion
	if index == 0 {
		newPartition.Start = int64(binary.Size(mbr))
	} else {
		newPartition.Start = mbr.Partitions[index-1].Start + mbr.Partitions[index-1].Size
	}
	//Se la asignamos al mbr
	mbr.Partitions[index] = newPartition
	//Escribimos de nuevo el mbr
	file.Seek(0, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct MBR
	var binaryMBR bytes.Buffer
	binary.Write(&binaryMBR, binary.BigEndian, &mbr)
	writeNextBytes(file, binaryMBR.Bytes())
	//[EXTENDIDA] Se crea el EBR incial
	if typeF == 'E' {
		ebr := extendedBootRecord{Next: -1, Start: newPartition.Start + 42}
		file.Seek(newPartition.Start, 0)
		//Empezamos el proceso de guardar la data del struct EBR
		var binaryEBR bytes.Buffer
		binary.Write(&binaryEBR, binary.BigEndian, &ebr)
		writeNextBytes(file, binaryEBR.Bytes())
	}
}

//Funcion para calcular el valor de un tamaño a partir de la unidad definida
func unitCalc(size int64, unit byte) (int64, error) {
	switch unit {
	case 'B':
		return size, nil
	case 'K':
		return 1024 * size, nil
	case 'M':
		return 1024 * 1024 * size, nil
	}
	return 0, fmt.Errorf("ERROR")
}
