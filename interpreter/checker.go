package interpreter

import (
	"fmt"

	"github.com/timtadh/lexmachine"
)

//Posibles parametros
type param struct {
	size      string
	path      string
	name      string
	unit      string
	Type      string
	fit       string
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
	aux := param{unit: "K"}

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
			fmt.Println("Se ejecuta comando", "----------------------------------------")
			fmt.Println(aux)
			aux = param{unit: "K"}
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
		aux.path = string(parameter.Lexeme)
	} else if paramType == "SIZE" {
		if aux.size != "" {
			fmt.Println("Error: Ya existe un SIZE asignado")
			return aux, fmt.Errorf("Error")
		}
		if tokens[parameter.Type] != "NUMBER" {
			fmt.Println("Error: No es un numero valido")
			return aux, fmt.Errorf("Error")
		}
		aux.size = string(parameter.Lexeme)
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
		aux.unit = tokens[parameter.Type]
	} else if paramType == "TYPE" {
		if aux.Type != "" {
			fmt.Println("Error: Ya existe un TYPE asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.Type = tokens[parameter.Type]
	} else if paramType == "FIT" {
		if aux.fit != "" {
			fmt.Println("Error: Ya existe un FIT asignado")
			return aux, fmt.Errorf("Error")
		}
		aux.fit = tokens[parameter.Type]
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
