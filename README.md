# TinyLang

TinyLang is a small interpreted programming language implemented in Go. The goal is to build a real language incrementally, learning how lexers, parsers, abstract syntax trees, and interpreters work.

## Current features

- Variables: `let score = 10;`
- Reassignment: `score = score + 5;`
- Numbers, strings, and booleans
- Arithmetic, comparisons, and boolean operators
- Parentheses and operator precedence
- `print(...)`
- `if / else` blocks
- `while` loops
- Helpful errors for invalid syntax and unknown variables

## Run TinyLang

You need Go 1.22 or newer.

Run a source file:

```sh
go run . examples/hello.tiny
```

Start the interactive prompt:

```sh
go run .
```

At the prompt, enter one statement per line. Type `:quit` to exit.

Run the tests:

```sh
go test ./...
```

## Example

```text
let name = "TinyLang";
let count = 1;
while count <= 3 {
    print(name + " says hello #" + count);
    count = count + 1;
}
if count == 4 {
    print("Loop completed.");
} else {
    print("Something went wrong.");
}
```

## Roadmap

- [x] Lexer and parser
- [x] Variables and expression evaluation
- [x] Conditionals and loops
- [x] Automated tests
- [ ] Functions and return values
- [ ] Better interactive prompt and multi-line input
- [ ] Source locations in error messages
- [ ] Bytecode compiler or virtual machine

TinyLang is a learning project, not a security boundary. Do not execute untrusted code in a privileged environment.
