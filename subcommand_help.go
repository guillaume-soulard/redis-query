package main

type HelpSubcommand struct{}

func (h HelpSubcommand) Accept(parameters *Parameters) bool {
	return !parameters.SetEnv.Cmd.Happened() &&
		!parameters.DelEnv.Cmd.Happened() &&
		!parameters.ListEnv.Cmd.Happened() &&
		!parameters.DescribeEnv.Cmd.Happened() &&
		!parameters.Scan.Cmd.Happened() &&
		!parameters.Loop.Cmd.Happened() &&
		!parameters.Command.Cmd.Happened() &&
		!parameters.Query.Cmd.Happened() &&
		!parameters.Connect.Cmd.Happened() &&
		!parameters.Migrate.Cmd.Happened() &&
		!parameters.Rdb.Cmd.Happened()
}

func (h HelpSubcommand) Execute(parameters *Parameters) (err error) {
	Print(parameters.Parser.Usage(nil))
	return err
}
