package interpreter

import (
	"log"
	"strings"

	"github.com/timtadh/lexmachine"
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

var symbols []string        // Tokens que representan simbolos
var keywords []string       // Comandos reservados para acciones
var params []string         //Parametros reservados para los comandos
var tokens []string         // Todos los Tokens identificados
var tokenIds map[string]int // Mapa de Tokens para identificarlos
var lexer *lex.Lexer        // Lexer es el objeto para construir el Scanner

//Funcion para interpretar la entrada a partir del lexmachine generado
func ScanInput(input string) *lexmachine.Scanner {
	s, err := lexer.Scanner([]byte(input))
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func init() {
	initTokens()
	var err error
	lexer, err = initLexer()
	if err != nil {
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
		"?",
		"*",
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
		"REP",
		/*
			Fase 2 del sistema de archivos
		*/
		"MKFS",
		"LOGIN",
		"LOGOUT",
		"MKGRP",
		"RMGRP",
		"MKUSR",
		"RMUSR",
		"CHMOD",
		"MKFILE",
		"CAT",
		"RM",
		"EDIT",
		"REN",
		"MKDIR",
		"CP",
		"MV",
		"FIND",
		"CHOWN",
		"CHGRP",
	}
	tokens = []string{
		/*
			Fase 1 del sistema de archivos
		*/
		"COMMENT",
		"ID",
		"ROUTE",
		"PARAMETER",
		"NUMBER",
		"IDN",
		"FINISHCOMMAND",
		/*
			Fase 2 del sistema de archivos
		*/
		"FILEN",
	}
	tokens = append(tokens, keywords...)
	tokens = append(tokens, params...)
	tokens = append(tokens, symbols...)
	tokenIds = make(map[string]int)
	for i, tok := range tokens {
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
	/*
		Fase 1 del sistema de archivos
	*/
	for _, lit := range symbols {
		r := "\\" + strings.Join(strings.Split(lit, ""), "\\")
		lexer.Add([]byte(r), token(lit))
	}
	for _, name := range keywords {
		lexer.Add([]byte(strings.ToLower(name)), token(name))
	}
	for _, name := range params {
		lexer.Add([]byte(strings.ToLower(name)), token(name))
	}
	lexer.Add([]byte(`#[^\n]*`), skip)
	lexer.Add([]byte(`([a-z]|[A-Z]|_)([a-z]|[A-Z]|[0-9]|_|\.)*`), token("ID"))
	lexer.Add([]byte(`"([^\\"]|(\\.))*"`), token("ROUTE"))
	lexer.Add([]byte(`/([a-z]|[A-Z]|[0-9]|_|/|-|\.)*`), token("ROUTE"))
	lexer.Add([]byte(`[0-9]+`), token("NUMBER"))
	lexer.Add([]byte(`-([a-z]|[A-Z])+`), token("PARAMETER"))
	lexer.Add([]byte(`-id[0-9]+`), token("IDN"))
	lexer.Add([]byte("( |\t|\r)+"), skip)
	lexer.Add([]byte("\n"), token("FINISHCOMMAND"))
	lexer.Add([]byte(`-file[0-9]+`), token("FILEN"))
	err := lexer.Compile()
	if err != nil {
		return nil, err
	}
	return lexer, nil
}
