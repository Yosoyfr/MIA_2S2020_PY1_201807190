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

//Tipos de type que puede tener una particion
var partitionsType []byte = []byte{'P', 'E', 'L'}

//Tipos de unidades que tener un tamaño
var unitType []byte = []byte{'B', 'K', 'M'}

//Funcion para checkear el comando que se va a ejecutar
func CommandChecker(s *lexmachine.Scanner) {
	paramType := ""
	var paramError error
	concat := false
	aux := param{}

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
			if strings.EqualFold(string(token.Lexeme), value) {
				aux.paramType = value
			}
		}
		//Verificamos si viene algun parametro para asignarlo
		if tokens[token.Type] == "PARAMETER" {
			paramType = string(token.Lexeme)
			paramType = strings.Replace(paramType, "-", "", -1)
			paramType = strings.ToUpper(paramType)
		} else if tokens[token.Type] == "IDN" {
			paramType = "IDN"
		}
		//Se asigna el parametro a la estructura del comando
		if tokens[token.Type] == "->" {
			tokenAux, err, _ := s.Next()
			//Si existe un error lexico termina la lectura de la lista de tokens
			if err != nil {
				fmt.Println(err)
				break
			}
			aux, paramError = paramDesigned(tokenAux.(*lexmachine.Token), paramType, aux)
			paramType = ""
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
			aux = param{}
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
		aux.unit = strings.ToUpper(string(parameter.Lexeme))[0]
	} else if paramType == "TYPE" {
		if aux.Type != 0 {
			fmt.Println("Error: Ya existe un TYPE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.Type = strings.ToUpper(string(parameter.Lexeme))[0]
	} else if paramType == "FIT" {
		if aux.fit != 0 {
			fmt.Println("Error: Ya existe un FIT asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.fit = strings.ToUpper(string(parameter.Lexeme))[0]
	} else if paramType == "DELETE" {
		if aux.delete != "" {
			fmt.Println("Error: Ya existe un DELETE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.delete = strings.ToUpper(string(parameter.Lexeme))
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
		aux.idn = append(aux.idn, string(parameter.Lexeme))
	} else {
		fmt.Println("Error: En la lectura del comando")
		return aux, fmt.Errorf("Error")
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
	command, err := validUnitTypes(command)
	if err != nil {
		return
	}
	//Ejecutamos el tipo de comando que llega
	switch command.paramType {
	case "EXEC":
		fmt.Println("Hara el exec")
	case "MKDISK":
		if requiredParameters([]string{"SIZE", "PATH", "NAME"}, command) != nil {
			return
		}
		commands.MKDisk(command.path, command.name, command.size, command.unit)
	case "RMDISK":
		if requiredParameters([]string{"PATH"}, command) != nil {
			return
		}
		commands.RMDisk(command.path)
	case "FDISK":
		if command.add != "" && command.delete == "" {
			fmt.Println("Se hara un add a una particion")
		} else if command.delete != "" && command.add == "" {
			fmt.Println("Se hara el delete de una particion")
		} else if command.delete == "" && command.add == "" {
			if requiredParameters([]string{"SIZE", "PATH", "NAME"}, command) != nil {
				return
			}
			commands.FKDisk(command.path, command.size, command.unit, command.Type, command.fit, command.name)
		} else {
			fmt.Println("Error: El comando FDISK proporciona un error en su estructura")
		}
	case "MOUNT":
		if command.path == "" && command.name == "" {
			commands.ShowMountedDisks()
		} else {
			if requiredParameters([]string{"PATH", "NAME"}, command) != nil {
				return
			}
			commands.Mount(command.path, command.name)
		}
	case "UNMOUNT":
		if requiredParameters([]string{"IDN"}, command) != nil {
			return
		}
		for _, idn := range command.idn {
			commands.Unmount(idn)
		}
	}
}

//Funcion que se encarga de verficiar que todos los parametros obligatorios esten contenidos
func requiredParameters(params []string, command param) error {
	//Recorremos la lista de parametros requeridos
	for _, parameter := range params {
		switch parameter {
		case "SIZE":
			if command.size == 0 {
				fmt.Println("Error: El comando a ejecutar necesita un size.")
				return fmt.Errorf("Error")
			}
		case "PATH":
			if command.path == "" {
				fmt.Println("Error: El comando a ejecutar necesita un path.")
				return fmt.Errorf("Error")
			}
		case "NAME":
			if command.name == "" {
				fmt.Println("Error: El comando a ejecutar necesita un nombre.")
				return fmt.Errorf("Error")
			}
		case "IDN":
			if len(command.idn) == 0 {
				fmt.Println("Error: El comando a ejecutar necesita un ID por lo menos.")
				return fmt.Errorf("Error")
			}
		}
	}
	/*
		Requerimiento de otros parametros
	*/
	err := fmt.Errorf("Error")
	//Verificar el tipo de particion
	for _, aux := range partitionsType {
		if command.Type == aux {
			err = nil
			break
		}
	}
	if err != nil {
		fmt.Println("Error: El tipo de particion no es valido.")
		return fmt.Errorf("Error")
	}
	//Todos los parametros obligatorios fueron cumplidos
	return nil
}

func validUnitTypes(command param) (param, error) {
	switch command.paramType {
	case "MKDISK":
		if command.unit == 0 {
			command.unit = 'M'
			return command, nil
		}
		for i := 1; i < len(unitType); i++ {
			if command.unit == unitType[i] {
				return command, nil
			}
		}
	case "FDISK":
		if command.unit == 0 {
			command.unit = 'K'
			return command, nil
		}
		for _, unit := range unitType {
			if command.unit == unit {
				return command, nil
			}
		}
	default:
		return command, nil
	}
	//En otro caso es error
	fmt.Println("Error: El tipo de unidad no es valido.")
	return command, fmt.Errorf("Error")
}
