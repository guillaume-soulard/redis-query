package main

type SubCommand interface {
	Accept(parameters *Parameters) bool
	Execute(parameters *Parameters) (err error)
}

var subCommands = []SubCommand{
	ListEnvSubCommand{},
	DescribeEnvSubCommand{},
	DelEnvSubCommand{},
	LoadEnvSubCommand{},
	SetEnvSubCommand{},
	ExecSubCommand{},
	LoopSubCommand{},
	ScanSubCommand{},
	QuerySubCommand{},
	HelpSubcommand{},
}

func Run(parameters *Parameters) {
	var err error
	for _, command := range subCommands {
		if command.Accept(parameters) {
			if err = command.Execute(parameters); err != nil {
				PrintErrorAndExit(err)
			}
		}
	}
}
