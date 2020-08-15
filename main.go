package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	/*
		Imports para los comandos de consola
	*/
	commands "./commands"
	/*
		imports para el interprete
	*/
	lexer "./interpreter"
	lex "github.com/timtadh/lexmachine"
)

//Funcion Main
func main() {
	/*
		fmt.Println("Prueba de creacion de un disco ----------")
		commands.MKDisk("disc_3.dsk", 25, 1)
		commands.ReadFile("disc_3.dsk")
		commands.RMDisk("disc_2.dsk")
	*/
	commands.ReadFile("disc_3.dsk")

	fmt.Println("Prueba del interpreter ----------")
	input := readMIAFile("input.mia")
	s, err := lexer.Lexer.Scanner([]byte(strings.ToLower(input)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Type    |  Position  | Lexeme ")
	fmt.Println("--------+------------+------------")
	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		if err != nil {
			fmt.Println("Hay un error")
			log.Fatal(err)
		}
		token := tok.(*lex.Token)
		fmt.Printf("%-7v | %v:%v-%v:%v | %-10v\n",
			lexer.Tokens[token.Type],
			token.StartLine,
			token.StartColumn,
			token.EndLine,
			token.EndColumn,
			string(token.Lexeme))
	}
}

//Funcion para leer los archivos con extension ".mia"
func readMIAFile(route string) string {
	var output string
	file, err := os.Open(route)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}
	return output
}
