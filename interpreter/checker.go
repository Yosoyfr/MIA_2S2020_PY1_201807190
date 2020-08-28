package interpreter

import (
	"fmt"
	"strconv"
	"strings"

	/*
		Imports para los comandos de consola
	*/
	commands "../commands"
	"github.com/timtadh/lexmachine"
)

//Posibles parametros
type param struct {
	size      int64
	path      string
	name      string
	unit      byte
	Type      byte
	fit       byte
	delete    string
	add       string
	idn       []string
	paramType string
}

//Funcion para checkear el comando que se va a ejecutar
func CommandChecker(s *lexmachine.Scanner) {
	paramType := ""
	var paramError error
	concat := false
	aux := param{unit: 'K'}

	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		//Si existe un error lexico termina la lectura de la lista de tokens
		if err != nil {
			fmt.Println(err)
			break
		}
		//Obtenemos el token a leer
		token := tok.(*lexmachine.Token)
		//Verificamos si es alguna reserveda de los comandos
		/*
			EXEC
			PAUSE
			MKDISK
			RMDISK
			FDISK
			MOUNT
			UNMOUNT
			REP
		*/
		for _, value := range keywords {
			if tokens[token.Type] == value {
				aux.paramType = value
			}
		}

		//Verificamos si viene algun parametro para asignarlo
		switch tokens[token.Type] {
		case "-PATH":
			paramType = "PATH"
		case "-SIZE":
			paramType = "SIZE"
		case "-NAME":
			paramType = "NAME"
		case "-UNIT":
			paramType = "UNIT"
		case "-TYPE":
			paramType = "TYPE"
		case "-FIT":
			paramType = "FIT"
		case "-DELETE":
			paramType = "DELETE"
		case "-ADD":
			paramType = "ADD"
		case "IDN":
			paramType = "IDN"
		case "->":
			tokenAux, _, _ := s.Next()
			aux, paramError = paramDesigned(tokenAux.(*lexmachine.Token), paramType, aux)
		}

		//Verificamos si no existe un error en la asignacion
		if paramError != nil {
			break
		}

		//Concatenamos la siguiente fila
		if tokens[token.Type] == "\\*" {
			concat = true
		}
		//Se ejecuta el comando
		if tokens[token.Type] == "FINISHCOMMAND" && !concat && aux.paramType != "" {
			controlCommands(aux)
			aux = param{unit: 'K'}
		}
		//Reseteamos la concatenacion
		if tokens[token.Type] == "FINISHCOMMAND" {
			concat = false
		}
	}
}

func paramDesigned(parameter *lexmachine.Token, paramType string, aux param) (param, error) {
	/*
		Fase 1
	*/
	if paramType == "PATH" {
		//Verificamos si ya existe un path
		if aux.path != "" {
			fmt.Println("Error: Ya existe un PATH asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "ROUTE" {
			fmt.Println("Error: No es una ruta valida")
			return aux, fmt.Errorf("Error")
		}
		aux.path = strings.Replace(string(parameter.Lexeme), "\"", "", -1)
	} else if paramType == "SIZE" {
		if aux.size != 0 {
			fmt.Println("Error: Ya existe un SIZE asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "NUMBER" {
			fmt.Println("Error: No es un numero valido")
			return aux, fmt.Errorf("Error")
		}
		aux.size, _ = strconv.ParseInt(string(parameter.Lexeme), 10, 64)
	} else if paramType == "NAME" {
		if aux.name != "" {
			fmt.Println("Error: Ya existe un NAME asignado", string(parameter.Lexeme))
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "ID" {
			fmt.Println("Error: Se esperaba un ID")
			return aux, fmt.Errorf("Error")
		}
		aux.name = string(parameter.Lexeme)
	} else if paramType == "UNIT" {

		aux.unit = tokens[parameter.Type][0]
	} else if paramType == "TYPE" {
		if aux.Type != 0 {
			fmt.Println("Error: Ya existe un TYPE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.Type = tokens[parameter.Type][0]
	} else if paramType == "FIT" {
		if aux.fit != 0 {
			fmt.Println("Error: Ya existe un FIT asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.fit = tokens[parameter.Type][0]
	} else if paramType == "DELETE" {
		if aux.delete != "" {
			fmt.Println("Error: Ya existe un DELETE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.delete = tokens[parameter.Type]
	} else if paramType == "ADD" {
		if aux.add != "" {
			fmt.Println("Error: Ya existe un ADD asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "NUMBER" {
			fmt.Println("Error: No es un numero valido")
			return aux, fmt.Errorf("Error")
		}
		aux.add = string(parameter.Lexeme)
	} else if paramType == "IDN" {
		if tokens[parameter.Type] != "ID" {
			fmt.Println("Error: Se esperaba un ID")
			return aux, fmt.Errorf("Error")
		}
		//TODO Agregar los id al slice de ids
	}

	return aux, nil
}

func controlCommands(command param) {
	if command.Type == 0 {
		command.Type = 'P'
	}
	if command.fit == 0 {
		command.fit = 'W'
	}
	if command.unit == 0 {
		command.unit = 'K'
	}
	fmt.Println(command)
	switch command.paramType {
	case "EXEC":
		fmt.Println("Hara el exec")
	case "MKDISK":
		commands.MKDisk(command.path, command.name, command.size, command.unit)
	case "RMDISK":
		commands.RMDisk(command.path)
	case "FDISK":
		commands.FKDisk(command.path, command.size, command.unit, command.Type, command.fit, command.name)
	}
}
