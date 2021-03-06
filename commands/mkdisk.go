package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

//Struct del MBR
type masterBootRecord struct {
	Size          int64
	CreatedAt     [19]byte
	DiskSignature int64
	Partitions    [4]partition
}

//Funcion que crea el archivo binario de cierto tamaño
func createBinaryFile(name string, size int64, unit byte) (file *os.File, newSize int64) {
	//Se crea un archivo binario de extension dsk
	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	//Obtenemos el tamaño de la particion a crear
	newSize, err = unitCalc(size, unit)
	if err != nil {
		fmt.Println("[ERROR] la unidad declarada no es valida.")
		return
	}
	aux := make([]byte, newSize)
	var binaryDisc bytes.Buffer
	binary.Write(&binaryDisc, binary.BigEndian, aux)
	writeNextBytes(file, binaryDisc.Bytes())
	return
}

//Funcion para crear el disco a partir de un comando MKDISK
func MKDisk(path string, name string, size int64, unit byte) {
	//Se verifica si existe el directorio o si no se crea
	createDirectory(path)
	path = path + "/" + name
	//Se crea el archivo binario que emula el disco
	file, newSize := createBinaryFile(path, size, unit)
	defer file.Close()
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
	fmt.Println("[-] El disco ha sido creado con exito.")
}

//Funcion para escribir los bytes en el archivo binario
func writeNextBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

//Funcion para crear directorio si es que no existe
func createDirectory(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("Error: El sistema no puedo crear el directorio")
			return
		}
	}
}
