package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

//Struct del MBR
type masterBootRecord struct {
	Size          int64
	CreatedAt     [19]byte
	DiskSignature int64
	Partitions    [4]partition
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

//Struct del EBR
type extendedBootRecord struct {
	Status byte
	Fit    byte
	Start  int64
	Size   int64
	Next   int64
	Name   [16]byte
}

//Funcion que crea el archivo binario de cierto tamaño
func createBinaryFile(name string, size int64, unit int) (file *os.File, newSize int64) {
	//Se crea un archivo binario de extension dsk
	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	if unit == 0 {
		newSize = size * 1024
	} else {
		newSize = size * 1024 * 1024
	}
	aux := make([]byte, newSize)
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, aux)
	writeNextBytes(file, binaryDisc.Bytes())
	return
}

//Funcion para crear el disco a partir de un comando MKDISK
func MKDisk(name string, size int64, unit int) {
	//Se crea el archivo binario que emula el disco
	file, newSize := createBinaryFile(name, size, unit)

	//Al no generar errores en la creacion del archivo binario que emulara el disco, empezamos agregar las respectivas estructuras
	mbr := masterBootRecord{Size: newSize,
		DiskSignature: rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	copy(mbr.CreatedAt[:], timestamp)

	//Guardamos el contenido del mbr generado en memorio
	mk := &mbr
	file.Seek(0, 0)
	//Empezamos el proceso de guardar en binario la data en memoria del struct MBR
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, mk)
	writeNextBytes(file, binaryDisc.Bytes())

}

//Funcion para escribir los bytes en el archivo binario
func writeNextBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

//Funcion para leer el archivo binario que representa el disco
func ReadFile(disc string) {
	//Se abre el archivo
	file, err := os.Open(disc)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	//Se instancia un struct de mbr
	mbr := masterBootRecord{}
	var size int = int(unsafe.Sizeof(mbr))
	file.Seek(0, 0)
	//Se obtiene la data del archivo binarios
	data := readNextBytes(file, size)
	buffer := bytes.NewBuffer(data)
	//Se asigna al mbr declarado para leer la informacion de ese disco
	err = binary.Read(buffer, binary.BigEndian, &mbr)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
	//fmt.Println(mbr)
	fmt.Printf("Tamaño: %d\nFecha de creacion: %s\nSignature: %d\n", mbr.Size, mbr.CreatedAt, mbr.DiskSignature)
}

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)
	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}
