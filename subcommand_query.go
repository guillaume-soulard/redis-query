package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Query struct {
	Block Block `@@`
}

type Block struct {
	Commands []Command `"{" @@* "}"`
}

type Command struct {
	Name *string ` @Command `
}

var (
	rqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Command`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{"Punct", `[{}]`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`String`, `'[^']*'|"[^"]*"`},
		{"whitespace", `\s+`},
	})
	parser = participle.MustBuild[Query](
		participle.Lexer(rqlLexer),
		participle.Unquote("String"),
		participle.UseLookahead(2),
		participle.CaseInsensitive("Command"),
	)
)

func Parse(query string) (block *Query, err error) {
	block, err = parser.ParseString("", query)
	return block, err
}

func executeQuery(params Parameters) {

}
