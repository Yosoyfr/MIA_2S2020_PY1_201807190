package interpreter

import (
	"fmt"
	"strings"

	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

var symbols []string        // Tokens que representan simbolos
var keywords []string       // Comandos reservados para acciones
var Tokens []string         // Todos los Tokens identificados
var tokenIds map[string]int // Mapa de Tokens para identificarlos
var Lexer *lex.Lexer        // Lexer es el objeto para construir el Scanner

func init() {
	initTokens()
	var err error
	Lexer, err = initLexer()
	if err != nil {
		fmt.Println("Hay un error")
		panic(err)
	}
}

func initTokens() {
	symbols = []string{
		/*
			Fase 1 del sistema de archivos
		*/
		"#",
		"->",
		"\\*",
	}
	keywords = []string{
		/*
			Fase 1 del sistema de archivos
		*/
		//Comandos a ejecutar
		"EXEC",
		"PAUSE",
		"MKDISK",
		"RMDISK",
		"FDISK",
		"MOUNT",
		"UNMOUNT",
		//Parametros que reciben los comandos
		"-SIZE",
		"-PATH",
		"-NAME",
		"-UNIT",
		"-TYPE",
		"-FIT",
		"-DELETE",
		"-ADD",
		//Valores que puede tomar el -UNIT
		"B",
		"K",
		"M",
		//Valores que puede tomar el -TYPE
		"P",
		"E",
		"L",
		//Valores que puede tomar el -FIT
		"BF",
		"FF",
		"WF",
		//Valores que puede tomar el -DELETE
		"FAST",
		"FULL",
	}
	Tokens = []string{
		/*
			Fase 1 del sistema de archivos
		*/
		"COMMENT",
		"ID",
		"ROUTE",
		"NUMBER",
		"IDN",
	}
	Tokens = append(Tokens, keywords...)
	Tokens = append(Tokens, symbols...)
	tokenIds = make(map[string]int)
	for i, tok := range Tokens {
		tokenIds[tok] = i
	}
}

func token(name string) lex.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(tokenIds[name], string(m.Bytes), m), nil
	}
}

func skip(*lex.Scanner, *machines.Match) (interface{}, error) {
	return nil, nil
}

func initLexer() (*lex.Lexer, error) {
	lexer := lex.NewLexer()

	for _, lit := range symbols {
		r := "\\" + strings.Join(strings.Split(lit, ""), "\\")
		lexer.Add([]byte(r), token(lit))
	}
	for _, name := range keywords {
		lexer.Add([]byte(strings.ToLower(name)), token(name))
	}
	/*
		Fase 1 del sistema de archivos
	*/
	lexer.Add([]byte(`#[^\n]*`), token("COMMENT"))
	lexer.Add([]byte(`([a-z]|[A-Z]|_)([a-z]|[A-Z]|[0-9]|_|\.)*`), token("ID"))
	lexer.Add([]byte(`"([^\\"]|(\\.))*"`), token("ROUTE"))
	lexer.Add([]byte(`/([a-z]|[A-Z]|[0-9]|_|/|-|\.)*`), token("ROUTE"))
	lexer.Add([]byte(`[0-9]+`), token("NUMBER"))
	lexer.Add([]byte(`-id[0-9]+`), token("IDN"))
	lexer.Add([]byte("( |\t|\n|\r)+"), skip)
	err := lexer.Compile()
	if err != nil {
		return nil, err
	}
	return lexer, nil
}
