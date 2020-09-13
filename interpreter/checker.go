package interpreter

import (
	"bufio"
	"fmt"
	"os"
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
	size         int64
	path         string
	name         string
	unit         byte
	Type         byte
	fit          byte
	delete       string
	add          string
	idn          []string
	id           string
	pathComplete bool
	txt          string
	filen        []string
	ruta         string
	paramType    string
}

//Tipos de type que puede tener una particion
var partitionsType []byte = []byte{'P', 'E', 'L'}

//Tipos de unidades que tener un tamaño
var unitType []byte = []byte{'B', 'K', 'M'}

//Funcion para leer los archivos con extension ".mia"
func ReadMIAFile(route string) string {
	var output string
	file, err := os.Open(route)
	if err != nil {
		fmt.Println("[ERROR]: El sistema no puede encontrar el archivo especificado.")
		return output
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}
	return output
}

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
		if aux.paramType == "" && string(token.Lexeme) != "\n" {
			fmt.Println("[ERROR]: Este comando \"", string(token.Lexeme), "\" no esta admitido en el sistema.")
			return
		}
		//Verificamos si viene algun parametro para asignarlo
		if tokens[token.Type] == "PARAMETER" {
			paramType = string(token.Lexeme)
			paramType = strings.Replace(paramType, "-", "", -1)
			paramType = strings.ToUpper(paramType)
		} else if tokens[token.Type] == "IDN" {
			paramType = "IDN"
		} else if tokens[token.Type] == "FILEN" {
			paramType = "FILEN"
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
		if paramType == "P" {
			aux.pathComplete = true
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
			fmt.Println("[ERROR]: Ya existe un PATH asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "ROUTE" {
			fmt.Println("[ERROR]: No es una ruta valida")
			return aux, fmt.Errorf("Error")
		}
		aux.path = strings.Replace(string(parameter.Lexeme), "\"", "", -1)
	} else if paramType == "SIZE" {
		if aux.size != 0 {
			fmt.Println("[ERROR]: Ya existe un SIZE asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "NUMBER" {
			fmt.Println("[ERROR]: No es un numero valido")
			return aux, fmt.Errorf("Error")
		}
		aux.size, _ = strconv.ParseInt(string(parameter.Lexeme), 10, 64)
	} else if paramType == "NAME" {
		if aux.name != "" {
			fmt.Println("[ERROR]: Ya existe un NAME asignado", string(parameter.Lexeme))
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "ID" && tokens[parameter.Type] != "ROUTE" {
			fmt.Println("[ERROR]: Se esperaba un ID o una cadena")
			return aux, fmt.Errorf("Error")
		} 
		aux.name = strings.Replace(string(parameter.Lexeme), "\"", "", -1)
	} else if paramType == "UNIT" {
		aux.unit = strings.ToUpper(string(parameter.Lexeme))[0]
	} else if paramType == "TYPE" {
		if aux.Type != 0 {
			fmt.Println("[ERROR]: Ya existe un TYPE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.Type = strings.ToUpper(string(parameter.Lexeme))[0]
	} else if paramType == "FIT" {
		if aux.fit != 0 {
			fmt.Println("[ERROR]: Ya existe un FIT asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.fit = strings.ToUpper(string(parameter.Lexeme))[0]
	} else if paramType == "DELETE" {
		if aux.delete != "" {
			fmt.Println("[ERROR]: Ya existe un DELETE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.delete = strings.ToUpper(string(parameter.Lexeme))
	} else if paramType == "ADD" {
		if aux.add != "" {
			fmt.Println("[ERROR]: Ya existe un ADD asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "NUMBER" {
			fmt.Println("[ERROR]: No es un numero valido")
			return aux, fmt.Errorf("Error")
		}
		aux.add = string(parameter.Lexeme)
	} else if paramType == "IDN" {
		if tokens[parameter.Type] != "ID" {
			fmt.Println("[ERROR]: Se esperaba un ID")
			return aux, fmt.Errorf("Error")
		}
		aux.idn = append(aux.idn, string(parameter.Lexeme))
	} else if paramType == "ID" {
		if tokens[parameter.Type] != "ID" {
			fmt.Println("[ERROR]: Se esperaba un ID")
			return aux, fmt.Errorf("Error")
		}
		aux.id = string(parameter.Lexeme)
	} else if paramType == "P" {
		fmt.Println("[ERROR]: Este comando no recibe parametros")
		return aux, fmt.Errorf("Error")
	} else if paramType == "CONT" {
		if tokens[parameter.Type] != "ROUTE" {
			fmt.Println("[ERROR]: Se esperaba una cadena.")
			return aux, fmt.Errorf("Error")
		}
		aux.txt = strings.Replace(string(parameter.Lexeme), "\"", "", -1)
	} else if paramType == "FILEN" {
		if tokens[parameter.Type] != "ROUTE" {
			fmt.Println("[ERROR]: Se esperaba una ruta.")
			return aux, fmt.Errorf("Error")
		}
		aux.filen = append(aux.filen, strings.Replace(string(parameter.Lexeme), "\"", "", -1))
	} else if paramType == "RUTA" {
		if tokens[parameter.Type] != "ROUTE" {
			fmt.Println("[ERROR]: Se esperaba una ruta.")
			return aux, fmt.Errorf("Error")
		}
		aux.ruta = strings.Replace(string(parameter.Lexeme), "\"", "", -1)
	} else {
		fmt.Println("[ERROR]: En la lectura del comando", aux.paramType, "el parametro \"", paramType, "\" no esta permitido.")
		return aux, fmt.Errorf("Error")
	}
	return aux, nil
}

func controlCommands(command param) {
	if command.Type == 0 || command.Type == 'F' {
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
		if requiredParameters([]string{"PATH"}, command) != nil {
			return
		}
		input := ReadMIAFile(command.path)
		CommandChecker(ScanInput(input))
		fmt.Println("[-] El archivo ha sido ejecutado con exito.")
	case "PAUSE":
		systemPaused()
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
			if requiredParameters([]string{"PATH", "NAME"}, command) != nil {
				return
			}
			if command.delete == "FAST" {
				commands.FDiskDelete(command.path, false, command.name)
			} else if command.delete == "FULL" {
				commands.FDiskDelete(command.path, true, command.name)
			} else {
				fmt.Println("[ERROR]: El comando FDISK proporciona un error en su estructura al tratar de eliminar")
				return
			}
		} else if command.delete == "" && command.add == "" {
			if requiredParameters([]string{"SIZE", "PATH", "NAME"}, command) != nil {
				return
			}
			commands.FDisk(command.path, command.size, command.unit, command.Type, command.fit, command.name)
		} else {
			fmt.Println("[ERROR]: El comando FDISK proporciona un error en su estructura")
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
	case "MKFS":
		if requiredParameters([]string{"ID"}, command) != nil {
			return
		}
		commands.Mkfs(command.id, "full")
	case "MKDIR":
		if requiredParameters([]string{"ID", "PATH"}, command) != nil {
			return
		}
		commands.Mkdir(command.id, command.path, command.pathComplete)
	case "MKFILE":
		if requiredParameters([]string{"ID", "PATH"}, command) != nil {
			return
		}
		commands.Mkfile(command.id, command.path, command.pathComplete, command.size, command.txt)
	case "CAT":
		if requiredParameters([]string{"ID", "FILEN"}, command) != nil {
			return
		}
		commands.Cat(command.id, command.filen)
	case "REN":
		if requiredParameters([]string{"ID", "PATH", "NAME"}, command) != nil {
			return
		}
		commands.Ren(command.id, command.path, command.name)
	case "REP":
		if requiredParameters([]string{"NAME", "PATH", "ID"}, command) != nil {
			return
		}
		commands.Reports(command.id, strings.ToUpper(command.name), command.path, command.ruta)
	}
}

//Funcion que se encarga de verficiar que todos los parametros obligatorios esten contenidos
func requiredParameters(params []string, command param) error {
	//Recorremos la lista de parametros requeridos
	for _, parameter := range params {
		switch parameter {
		case "SIZE":
			if command.size == 0 {
				fmt.Println("[ERROR]: El comando a ejecutar necesita un size.")
				return fmt.Errorf("Error")
			}
		case "PATH":
			if command.path == "" {
				fmt.Println("[ERROR]: El comando a ejecutar necesita un path.")
				return fmt.Errorf("Error")
			}
		case "NAME":
			if command.name == "" {
				fmt.Println("[ERROR]: El comando a ejecutar necesita un nombre.")
				return fmt.Errorf("Error")
			}
		case "IDN":
			if len(command.idn) == 0 {
				fmt.Println("[ERROR]: El comando a ejecutar necesita un ID por lo menos.")
				return fmt.Errorf("Error")
			}
		case "ID":
			if command.id == "" {
				fmt.Println("[ERROR]: El comando a ejecutar necesita un ID por lo menos.")
				return fmt.Errorf("Error")
			}
		case "CAT":
			if len(command.filen) == 0 {
				fmt.Println("[ERROR]: El comando a ejecutar necesita un file por lo menos.")
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
		fmt.Println("[ERROR]: El tipo de particion no es valido.")
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
	fmt.Println("[ERROR]: El tipo de unidad no es valido.")
	return command, fmt.Errorf("Error")
}

func systemPaused() {
	fmt.Print("[PAUSE] El sistema a pausado toda ejecucion, presione ENTER para continuar con la ejecucion.")
	bufio.NewScanner(os.Stdin).Scan()
}
