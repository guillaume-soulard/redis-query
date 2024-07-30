package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Query struct {
	Block Block `@@`
}

type Block struct {
	Commands []Command `"{" (@@)* "}"`
}

type Command struct {
	Name  *string   `@Ident`
	Args  *[]string `@String* `
	Block *Block    `[">" @@]`
}

var (
	rqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Command`, `[a-zA-Z][a-zA-Z0-9]*`},
		{"Punct", `[{}>]`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`String`, `'[^']*'|"[^"]*"`},
		{"whitespace", `\s+`},
	})
	parser = participle.MustBuild[Query](
		participle.Lexer(rqlLexer),
		participle.Unquote("String"),
		participle.UseLookahead(2),
		participle.CaseInsensitive("Ident"),
	)
)

func Parse(query string) (block *Query, err error) {
	block, err = parser.ParseString("", query)
	return block, err
}

func executeQuery(params Parameters) {

}
