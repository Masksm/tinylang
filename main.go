package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type tokenKind int

const (
	tEOF tokenKind = iota
	tNumber
	tString
	tIdentifier
	tLet
	tPrint
	tIf
	tElse
	tWhile
	tTrue
	tFalse
	tLeftParen
	tRightParen
	tLeftBrace
	tRightBrace
	tSemicolon
	tComma
	tPlus
	tMinus
	tStar
	tSlash
	tPercent
	tBang
	tEqual
	tEqualEqual
	tBangEqual
	tLess
	tLessEqual
	tGreater
	tGreaterEqual
	tAnd
	tOr
)

type token struct {
	kind tokenKind
	text string
	line int
}

type lexer struct {
	source string
	pos    int
	line   int
}

func lex(source string) ([]token, error) {
	l := &lexer{source: source, line: 1}
	var tokens []token
	for l.pos < len(l.source) {
		c := l.source[l.pos]
		if c == ' ' || c == '\t' || c == '\r' {
			l.pos++
			continue
		}
		if c == '\n' {
			l.line++
			l.pos++
			continue
		}
		if c == '#' {
			for l.pos < len(l.source) && l.source[l.pos] != '\n' {
				l.pos++
			}
			continue
		}
		start, line := l.pos, l.line
		if isDigit(c) {
			l.pos++
			for l.pos < len(l.source) && (isDigit(l.source[l.pos]) || l.source[l.pos] == '.') {
				l.pos++
			}
			tokens = append(tokens, token{tNumber, l.source[start:l.pos], line})
			continue
		}
		if isAlpha(c) {
			l.pos++
			for l.pos < len(l.source) && isAlphaNumeric(l.source[l.pos]) {
				l.pos++
			}
			word := l.source[start:l.pos]
			kind := map[string]tokenKind{
				"let": tLet, "print": tPrint, "if": tIf, "else": tElse,
				"while": tWhile, "true": tTrue, "false": tFalse,
			}[word]
			if kind == 0 {
				kind = tIdentifier
			}
			tokens = append(tokens, token{kind, word, line})
			continue
		}
		if c == '"' {
			l.pos++
			var value strings.Builder
			closed := false
			for l.pos < len(l.source) {
				ch := l.source[l.pos]
				l.pos++
				if ch == '"' {
					closed = true
					break
				}
				if ch == '\n' {
					l.line++
				}
				if ch == '\\' && l.pos < len(l.source) {
					escaped := l.source[l.pos]
					l.pos++
					switch escaped {
					case 'n':
						value.WriteByte('\n')
					case 't':
						value.WriteByte('\t')
					case '"', '\\':
						value.WriteByte(escaped)
					default:
						return nil, fmt.Errorf("line %d: unknown escape \\%c", line, escaped)
					}
				} else {
					value.WriteByte(ch)
				}
			}
			if !closed {
				return nil, fmt.Errorf("line %d: unterminated string", line)
			}
			tokens = append(tokens, token{tString, value.String(), line})
			continue
		}
		l.pos++
		switch c {
		case '(':
			tokens = append(tokens, token{tLeftParen, "(", line})
		case ')':
			tokens = append(tokens, token{tRightParen, ")", line})
		case '{':
			tokens = append(tokens, token{tLeftBrace, "{", line})
		case '}':
			tokens = append(tokens, token{tRightBrace, "}", line})
		case ';':
			tokens = append(tokens, token{tSemicolon, ";", line})
		case ',':
			tokens = append(tokens, token{tComma, ",", line})
		case '+':
			tokens = append(tokens, token{tPlus, "+", line})
		case '-':
			tokens = append(tokens, token{tMinus, "-", line})
		case '*':
			tokens = append(tokens, token{tStar, "*", line})
		case '/':
			tokens = append(tokens, token{tSlash, "/", line})
		case '%':
			tokens = append(tokens, token{tPercent, "%", line})
		case '!':
			if l.match('=') { tokens = append(tokens, token{tBangEqual, "!=", line}) } else { tokens = append(tokens, token{tBang, "!", line}) }
		case '=':
			if l.match('=') { tokens = append(tokens, token{tEqualEqual, "==", line}) } else { tokens = append(tokens, token{tEqual, "=", line}) }
		case '<':
			if l.match('=') { tokens = append(tokens, token{tLessEqual, "<=", line}) } else { tokens = append(tokens, token{tLess, "<", line}) }
		case '>':
			if l.match('=') { tokens = append(tokens, token{tGreaterEqual, ">=", line}) } else { tokens = append(tokens, token{tGreater, ">", line}) }
		case '&':
			if !l.match('&') { return nil, fmt.Errorf("line %d: expected &&", line) }
			tokens = append(tokens, token{tAnd, "&&", line})
		case '|':
			if !l.match('|') { return nil, fmt.Errorf("line %d: expected ||", line) }
			tokens = append(tokens, token{tOr, "||", line})
		default:
			return nil, fmt.Errorf("line %d: unexpected character %q", line, c)
		}
	}
	tokens = append(tokens, token{kind: tEOF, line: l.line})
	return tokens, nil
}

func (l *lexer) match(want byte) bool {
	if l.pos >= len(l.source) || l.source[l.pos] != want { return false }
	l.pos++
	return true
}
func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isAlpha(c byte) bool { return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isAlphaNumeric(c byte) bool { return isAlpha(c) || isDigit(c) }

type expr interface{ eval(*environment) (any, error) }
type literalExpr struct{ value any }
func (e literalExpr) eval(*environment) (any, error) { return e.value, nil }
type variableExpr struct{ name string; line int }
func (e variableExpr) eval(env *environment) (any, error) {
	for scope := env; scope != nil; scope = scope.parent {
		if v, ok := scope.values[e.name]; ok {
			return v, nil
		}
	}
	return nil, fmt.Errorf("line %d: unknown variable %q", e.line, e.name)
}
type unaryExpr struct{ op tokenKind; right expr; line int }
func (e unaryExpr) eval(env *environment) (any, error) {
	v, err := e.right.eval(env); if err != nil { return nil, err }
	switch e.op {
	case tMinus:
		n, ok := v.(float64); if !ok { return nil, fmt.Errorf("line %d: unary '-' needs a number", e.line) }; return -n, nil
	case tBang:
		b, ok := v.(bool); if !ok { return nil, fmt.Errorf("line %d: '!' needs a boolean", e.line) }; return !b, nil
	}
	return nil, fmt.Errorf("line %d: unsupported unary operator", e.line)
}
type binaryExpr struct{ left expr; op tokenKind; right expr; line int }
func (e binaryExpr) eval(env *environment) (any, error) {
	a, err := e.left.eval(env); if err != nil { return nil, err }
	if e.op == tAnd {
		ab, ok := a.(bool); if !ok { return nil, fmt.Errorf("line %d: && needs booleans", e.line) }
		if !ab { return false, nil }
		b, err := e.right.eval(env); if err != nil { return nil, err }; bb, ok := b.(bool); if !ok { return nil, fmt.Errorf("line %d: && needs booleans", e.line) }; return bb, nil
	}
	if e.op == tOr {
		ab, ok := a.(bool); if !ok { return nil, fmt.Errorf("line %d: || needs booleans", e.line) }
		if ab { return true, nil }
		b, err := e.right.eval(env); if err != nil { return nil, err }; bb, ok := b.(bool); if !ok { return nil, fmt.Errorf("line %d: || needs booleans", e.line) }; return bb, nil
	}
	b, err := e.right.eval(env); if err != nil { return nil, err }
	if e.op == tPlus {
		if x, ok := a.(string); ok { return x + stringify(b), nil }
		if x, ok := a.(float64); ok { y, ok := b.(float64); if ok { return x + y, nil } }
		return nil, fmt.Errorf("line %d: '+' needs two numbers or a string on the left", e.line)
	}
	if e.op == tEqualEqual { return equal(a,b), nil }
	if e.op == tBangEqual { return !equal(a,b), nil }
	x, xok := a.(float64); y, yok := b.(float64)
	if !xok || !yok { return nil, fmt.Errorf("line %d: operator %q needs numbers", e.line, tokenName(e.op)) }
	switch e.op {
	case tMinus: return x-y, nil
	case tStar: return x*y, nil
	case tSlash: if y == 0 { return nil, fmt.Errorf("line %d: division by zero", e.line) }; return x/y, nil
	case tPercent: if y == 0 { return nil, fmt.Errorf("line %d: remainder by zero", e.line) }; return float64(int64(x)%int64(y)), nil
	case tLess: return x<y, nil
	case tLessEqual: return x<=y, nil
	case tGreater: return x>y, nil
	case tGreaterEqual: return x>=y, nil
	}
	return nil, fmt.Errorf("line %d: unsupported operator", e.line)
}
func equal(a,b any) bool {
	switch x := a.(type) {
	case float64: y, ok := b.(float64); return ok && x == y
	case string: y, ok := b.(string); return ok && x == y
	case bool: y, ok := b.(bool); return ok && x == y
	}
	return false
}
func stringify(v any) string {
	switch x := v.(type) {
	case float64: return strconv.FormatFloat(x, 'f', -1, 64)
	case string: return x
	case bool: if x { return "true" }; return "false"
	default: return fmt.Sprint(v)
	}
}

type stmt interface{ exec(*environment) error }
type environment struct{ values map[string]any; parent *environment; output func(string) }
func newEnvironment(output func(string)) *environment { return &environment{values: map[string]any{}, output: output} }
type letStmt struct{ name string; value expr; line int }
func (s letStmt) exec(env *environment) error {
	if _, exists := env.values[s.name]; exists { return fmt.Errorf("line %d: variable %q already declared in this scope", s.line, s.name) }
	v, err := s.value.eval(env); if err != nil { return err }; env.values[s.name] = v; return nil
}
type assignStmt struct{ name string; value expr; line int }
func (s assignStmt) exec(env *environment) error {
	v, err := s.value.eval(env); if err != nil { return err }
	for scope := env; scope != nil; scope = scope.parent { if _, ok := scope.values[s.name]; ok { scope.values[s.name] = v; return nil } }
	return fmt.Errorf("line %d: cannot assign unknown variable %q", s.line, s.name)
}
type printStmt struct{ value expr }
func (s printStmt) exec(env *environment) error { v, err := s.value.eval(env); if err != nil { return err }; env.output(stringify(v)); return nil }
type expressionStmt struct{ value expr }
func (s expressionStmt) exec(env *environment) error { _, err := s.value.eval(env); return err }
type blockStmt struct{ statements []stmt }
func (s blockStmt) exec(env *environment) error { child := &environment{values: map[string]any{}, parent: env, output: env.output}; for _, st := range s.statements { if err := st.exec(child); err != nil { return err } }; return nil }
type ifStmt struct{ condition expr; thenBranch stmt; elseBranch stmt; line int }
func (s ifStmt) exec(env *environment) error {
	v, err := s.condition.eval(env); if err != nil { return err }; b, ok := v.(bool); if !ok { return fmt.Errorf("line %d: if condition must be boolean", s.line) }
	if b { return s.thenBranch.exec(env) }; if s.elseBranch != nil { return s.elseBranch.exec(env) }; return nil
}
type whileStmt struct{ condition expr; body stmt; line int }
func (s whileStmt) exec(env *environment) error {
	for { v, err := s.condition.eval(env); if err != nil { return err }; b, ok := v.(bool); if !ok { return fmt.Errorf("line %d: while condition must be boolean", s.line) }; if !b { return nil }; if err := s.body.exec(env); err != nil { return err } }
}

type parser struct{ tokens []token; pos int }
func parse(tokens []token) ([]stmt, error) {
	p := &parser{tokens: tokens}
	var result []stmt
	for !p.check(tEOF) { before := p.pos; s, err := p.statement(); if err != nil { return nil, err }; result = append(result, s); if p.pos == before { return nil, fmt.Errorf("line %d: parser made no progress", p.peek().line) } }
	return result, nil
}
func (p *parser) statement() (stmt, error) {
	if p.match(tLet) { name, err := p.consume(tIdentifier, "expected a variable name after let"); if err != nil { return nil, err }; if _, err = p.consume(tEqual, "expected '=' after variable name"); err != nil { return nil, err }; value, err := p.expression(); if err != nil { return nil, err }; if _, err = p.consume(tSemicolon, "expected ';' after declaration"); err != nil { return nil, err }; return letStmt{name.text,value,name.line},nil }
	if p.match(tPrint) {
		if _, err := p.consume(tLeftParen, "expected '(' after print"); err != nil { return nil, err }
		value, err := p.expression(); if err != nil { return nil, err }
		if _, err = p.consume(tRightParen, "expected ')' after print value"); err != nil { return nil, err }
		if _, err = p.consume(tSemicolon, "expected ';' after print"); err != nil { return nil, err }
		return printStmt{value},nil
	}
	if p.match(tIf) {
		line := p.previous().line
		if _, err := p.consume(tLeftParen, "expected '(' after if"); err != nil { return nil, err }
		condition, err := p.expression(); if err != nil { return nil, err }
		if _, err = p.consume(tRightParen, "expected ')' after condition"); err != nil { return nil, err }
		thenBranch, err := p.block(); if err != nil { return nil, err }
		var elseBranch stmt
		if p.match(tElse) { elseBranch, err = p.block(); if err != nil { return nil, err } }
		return ifStmt{condition,thenBranch,elseBranch,line},nil
	}
	if p.match(tWhile) {
		line := p.previous().line
		if _, err := p.consume(tLeftParen, "expected '(' after while"); err != nil { return nil, err }
		condition, err := p.expression(); if err != nil { return nil, err }
		if _, err = p.consume(tRightParen, "expected ')' after condition"); err != nil { return nil, err }
		body, err := p.block(); if err != nil { return nil, err }; return whileStmt{condition,body,line},nil
	}
	if p.check(tIdentifier) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].kind == tEqual {
		name := p.advance(); p.advance(); value, err := p.expression(); if err != nil { return nil, err }; if _, err = p.consume(tSemicolon, "expected ';' after assignment"); err != nil { return nil, err }; return assignStmt{name.text,value,name.line},nil
	}
	value, err := p.expression(); if err != nil { return nil, err }
	if _, err = p.consume(tSemicolon, "expected ';' after expression"); err != nil { return nil, err }
	return expressionStmt{value},nil
}
func (p *parser) block() (stmt,error) {
	if _, err := p.consume(tLeftBrace, "expected '{' before block"); err != nil { return nil, err }
	var statements []stmt
	for !p.check(tRightBrace) && !p.check(tEOF) { s, err := p.statement(); if err != nil { return nil, err }; statements = append(statements,s) }
	if _, err := p.consume(tRightBrace, "expected '}' after block"); err != nil { return nil, err }
	return blockStmt{statements},nil
}
func (p *parser) expression() (expr,error) { return p.or() }
func (p *parser) or() (expr,error) { e,err:=p.and(); if err!=nil{return nil,err}; for p.match(tOr) { op:=p.previous(); r,er:=p.and(); if er!=nil{return nil,er}; e=binaryExpr{e,op.kind,r,op.line} }; return e,nil }
func (p *parser) and() (expr,error) { e,err:=p.equality(); if err!=nil{return nil,err}; for p.match(tAnd) { op:=p.previous(); r,er:=p.equality(); if er!=nil{return nil,er}; e=binaryExpr{e,op.kind,r,op.line} }; return e,nil }
func (p *parser) equality() (expr,error) { e,err:=p.comparison(); if err!=nil{return nil,err}; for p.match(tEqualEqual,tBangEqual) { op:=p.previous(); r,er:=p.comparison(); if er!=nil{return nil,er}; e=binaryExpr{e,op.kind,r,op.line} }; return e,nil }
func (p *parser) comparison() (expr,error) { e,err:=p.term(); if err!=nil{return nil,err}; for p.match(tLess,tLessEqual,tGreater,tGreaterEqual) { op:=p.previous(); r,er:=p.term(); if er!=nil{return nil,er}; e=binaryExpr{e,op.kind,r,op.line} }; return e,nil }
func (p *parser) term() (expr,error) { e,err:=p.factor(); if err!=nil{return nil,err}; for p.match(tPlus,tMinus) { op:=p.previous(); r,er:=p.factor(); if er!=nil{return nil,er}; e=binaryExpr{e,op.kind,r,op.line} }; return e,nil }
func (p *parser) factor() (expr,error) { e,err:=p.unary(); if err!=nil{return nil,err}; for p.match(tStar,tSlash,tPercent) { op:=p.previous(); r,er:=p.unary(); if er!=nil{return nil,er}; e=binaryExpr{e,op.kind,r,op.line} }; return e,nil }
func (p *parser) unary() (expr,error) { if p.match(tBang,tMinus) { op:=p.previous(); r,err:=p.unary(); if err!=nil{return nil,err}; return unaryExpr{op.kind,r,op.line},nil }; return p.primary() }
func (p *parser) primary() (expr,error) {
	t:=p.advance()
	switch t.kind {
	case tNumber: n,err:=strconv.ParseFloat(t.text,64); if err!=nil{return nil,fmt.Errorf("line %d: invalid number %q",t.line,t.text)}; return literalExpr{n},nil
	case tString: return literalExpr{t.text},nil
	case tTrue: return literalExpr{true},nil
	case tFalse: return literalExpr{false},nil
	case tIdentifier: return variableExpr{t.text,t.line},nil
	case tLeftParen: e,err:=p.expression(); if err!=nil{return nil,err}; if _,err=p.consume(tRightParen,"expected ')' after expression"); err!=nil{return nil,err}; return e,nil
	default: return nil,fmt.Errorf("line %d: expected an expression, got %q",t.line,t.text)
	}
}
func (p *parser) match(kinds ...tokenKind) bool { for _,k:=range kinds { if p.check(k) { p.advance(); return true } }; return false }
func (p *parser) check(k tokenKind) bool { return p.peek().kind==k }
func (p *parser) advance() token { t:=p.peek(); if !p.check(tEOF) { p.pos++ }; return t }
func (p *parser) peek() token { if p.pos>=len(p.tokens) { return token{kind:tEOF,line:1} }; return p.tokens[p.pos] }
func (p *parser) previous() token { if p.pos==0{return token{}}; return p.tokens[p.pos-1] }
func (p *parser) consume(k tokenKind,msg string) (token,error) { if p.check(k){return p.advance(),nil}; t:=p.peek(); return token{},fmt.Errorf("line %d: %s (got %q)",t.line,msg,t.text) }
func tokenName(k tokenKind) string { switch k { case tMinus:return "-"; case tStar:return "*"; case tSlash:return "/"; case tPercent:return "%"; case tLess:return "<"; case tLessEqual:return "<="; case tGreater:return ">"; case tGreaterEqual:return ">=" }; return "operator" }

func run(source string, output func(string)) error {
	tokens, err := lex(source); if err != nil { return err }
	statements, err := parse(tokens); if err != nil { return err }
	env := newEnvironment(output)
	for _, s := range statements { if err := s.exec(env); err != nil { return err } }
	return nil
}
func main() {
	if len(os.Args) > 1 {
		data, err := os.ReadFile(os.Args[1]); if err != nil { fmt.Fprintln(os.Stderr,err); os.Exit(1) }
		if err := run(string(data),func(s string){fmt.Println(s)}); err != nil { fmt.Fprintln(os.Stderr,err); os.Exit(1) }
		return
	}
	fmt.Println("TinyLang interactive prompt. Enter statements; type :quit to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	var source strings.Builder
	for {
		fmt.Print("tiny> ")
		if !scanner.Scan() { break }
		line := scanner.Text()
		if strings.TrimSpace(line)==":quit" { break }
		source.WriteString(line+"\n")
		// Each submitted line is parsed and run as a small program. Variables persist
		// only within that submitted program in this first version.
		if err := run(source.String(),func(s string){fmt.Println(s)}); err != nil { fmt.Println("error:",err) }
		source.Reset()
	}
}
