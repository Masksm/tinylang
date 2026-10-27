package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestArithmeticAndPrecedence(t *testing.T) {
	var got []string
	err := run("print(2 + 3 * 4); print((2 + 3) * 4);", func(s string) { got = append(got, s) })
	if err != nil { t.Fatal(err) }
	want := []string{"14", "20"}
	if !reflect.DeepEqual(got,want) { t.Fatalf("got %v, want %v",got,want) }
}

func TestVariablesAndLoop(t *testing.T) {
	var got []string
	source := "let n = 1; while (n <= 3) { print(n); n = n + 1; }"
	if err := run(source,func(s string){got=append(got,s)}); err != nil { t.Fatal(err) }
	want := []string{"1","2","3"}
	if !reflect.DeepEqual(got,want) { t.Fatalf("got %v, want %v",got,want) }
}

func TestIfElseAndStrings(t *testing.T) {
	var got []string
	source := "let name = \"world\"; if (name == \"world\") { print(\"hello \" + name); } else { print(\"no\"); }"
	if err := run(source,func(s string){got=append(got,s)}); err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(got,[]string{"hello world"}) { t.Fatalf("got %v",got) }
}

func TestDivisionByZero(t *testing.T) {
	err := run("print(8 / 0);",func(string){})
	if err == nil || !strings.Contains(err.Error(),"division by zero") { t.Fatalf("expected division by zero error, got %v",err) }
}

func TestUnknownVariable(t *testing.T) {
	err := run("print(missing);",func(string){})
	if err == nil || !strings.Contains(err.Error(),"unknown variable") { t.Fatalf("expected unknown variable error, got %v",err) }
}

func TestBooleanOperators(t *testing.T) {
	var got []string
	if err := run("print(true && !false); print(false || true);",func(s string){got=append(got,s)}); err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(got,[]string{"true","true"}) { t.Fatalf("got %v",got) }
}
