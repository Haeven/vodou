package main

import (
	"bufio"
	"flag"
	"fmt"
	"json"
	"os"
	"strconv"
)

var hadError bool

type token uint8

// 	hadRuntimeError bool

// r = newRunner(os.Stdout, os.Stderr)
// )

func main() {
	var filePath string

	// Declare command-line flag 'filepath', store in filePath variable
	flag.StringVar(&filePath, "filepath", "", "File path")
	flag.Parse()

	// If file path is defined, open file
	if filePath != "" {
		runFile(filePath)
		// ...else run command as prompt
	} else {
		runPrompt()
	}
}

// Run prompt command directly
func runPrompt() {
	inpScanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !inpScanner.Scan() {
			break
		}

		line := inpScanner.Text()
		run(line)
		hadError = false
	}
}

// Open input file
func runFile(path string) {
	file, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}
	print(file)
	run(string(file))

	if hadError {
		panic(err)
	}
}

// Begin scanning file
func run(source string) {
	scanner := NewScanner(source)
	tokens := scanner.scanTokens()

	for _, tkn := range tokens {
		s, _ := json.MarshalIndent(tkn, "", "\t")
		fmt.Printf("%s\n", s)
	}
}

// Print error to console
func report(line int, where string, msg string) {
	fmt.Println("[line %s] Error %d: %s", line, where, msg)
	hadError = true
}

type Token struct {
	TokenType token
	Lexeme    string
	Literal   interface{}
	Line      int
}

// Define new token instance
func newToken(tokenType token, lexeme string, literal interface{}, line int) Token {
	return Token{TokenType: tokenType, Lexeme: lexeme, Literal: literal, Line: line}
}

// Convert token to string
func toString(t Token) string {
	return fmt.Sprintf("%s %d %s", t.TokenType, t.Lexeme, t.Literal)
}

const (
	// single-character tokens
	tLeftParen token = iota
	tRightParen
	tLeftBrace
	tRightBrace
	tComma
	tDot
	tMinus
	tPlus
	tSemicolon
	tSlash
	tStar

	// one or two character tokens
	tBang
	tBangEqual
	tEqual
	tEqualEqual
	tGreater
	tGreaterEqual
	tLess
	tLessEqual

	// literals
	tIdentifier
	tString
	tNumber

	// keywords
	tAnd
	tClass
	tElse
	tFalse
	tFun
	tFor
	tIf
	tNil
	tOr
	tPrint
	tReturn
	tSuper
	tThis
	tTrue
	tVar
	tWhile

	tEof
)

var keywordTokens = map[string]token{
	"and":    tAnd,
	"class":  tClass,
	"else":   tElse,
	"false":  tFalse,
	"for":    tFor,
	"fun":    tFun,
	"if":     tIf,
	"nil":    tNil,
	"or":     tOr,
	"print":  tPrint,
	"return": tReturn,
	"super":  tSuper,
	"this":   tThis,
	"true":   tTrue,
	"var":    tVar,
	"while":  tWhile,
}

type Scanner struct {
	start   int
	current int
	line    int
	source  string
	tokens  []Token
}

// Instantiates new scanner,
// returns instance of Scanner starting at line 1
func NewScanner(source string) *Scanner {
	return &Scanner{source: source, line: 1}
}

func (s *Scanner) scanTokens() []Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	return s.tokens
}

// Check current character is known
func (s *Scanner) scanToken() {
	char := s.advance()
	switch char {
	case '(':
		s.addToken(tLeftParen)
	case ')':
		s.addToken(tRightParen)
	case '{':
		s.addToken(tLeftBrace)
	case '}':
		s.addToken(tRightBrace)
	case ',':
		s.addToken(tComma)
	case '.':
		s.addToken(tDot)
	case '-':
		s.addToken(tMinus)
	case '+':
		s.addToken(tPlus)
	case ';':
		s.addToken(tSemicolon)
	case '*':
		s.addToken(tStar)
	case '!':
		if s.match("=") {
			s.addToken(tBangEqual)
		} else {
			s.addToken(tBang)
		}
	case '=':
		if s.match("=") {
			s.addToken(tEqualEqual)
		} else {
			s.addToken(tEqual)
		}
	case '<':
		if s.match("=") {
			s.addToken(tLessEqual)
		} else {
			s.addToken(tLess)
		}
	case '>':
		if s.match("=") {
			s.addToken(tGreaterEqual)
		} else {
			s.addToken(tGreater)
		}
	case '/':
		if s.match("/") {
			// Comment goes until end of line
			for s.peek() != "\n" && !s.isAtEnd() {
				s.advance()
			}
		}
	case ' ':
	case '\r':
	case '\t':
		break
	case '\n':
		s.line++
	case '"':
		s.scanString()
	case 'o':
		if s.match("r") {
			s.addToken('r')
		}
	default:
		if isDigit(char) {
			s.number()
		} else if s.isAlpha(char) {
			s.identifier()
		} else {
			reportError(s.line, "Unexpected character.")
		}
	}
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	val, _ := strconv.ParseFloat(s.source[s.start:s.current], 64)
	s.addTokenWithLiteral(tNumber, val)
}

func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	s.addToken(tIdentifier)
}

func (s *Scanner) isAlpha(char rune) bool {
	return char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' ||
		char == '_'
}

func (s *Scanner) isAlphaNumeric(char rune) bool {
	return s.isAlpha(char) || isDigit(char)
}

// Check if next character is part of current lexeme
func (s *Scanner) match(expected string) bool {
	if s.isAtEnd() {
		return false
	}

	if s.source != expected {
		return false
	}

	s.current++

	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		str := '\000'
		return str
	}

	return rune(s.source[s.current])
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return '\000'
	}

	return rune(s.source[s.current+1])
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

// Basic error report
func reportError(line int, msg string) {
	report(line, "", msg)
}

// Check if scanner is at end of input file
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

// Advance scanner to next character
func (s *Scanner) advance() rune {
	curr := rune(s.source[s.current])
	s.current++
	return curr
}

func (s *Scanner) addToken(tokenType token) {
	s.addTokenWithLiteral(tokenType, nil)
}

// Add current lexeme to list of tokens
func (s *Scanner) addTokenWithLiteral(tokenType token, literal interface{}) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, newToken(tokenType, text, literal, s.line))
}

func (s *Scanner) scanString() {
	for s.peek() != '\'' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}

		s.advance()
	}

	if s.isAtEnd() {
		reportError(s.line, "Unterminated string")
	}

	s.advance()

	value := s.source[s.start+1 : s.current+1]
	s.addTokenWithLiteral(tString, value)
}
