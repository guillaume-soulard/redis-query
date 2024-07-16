package main

func ParseArguments(str string) (args []string) {
	args = make([]string, 0)
	arg := ""
	inQuote := false
	for _, char := range str {
		if char == ' ' && !inQuote {
			args = append(args, arg)
			arg = ""
		} else if char == '"' {
			if !inQuote {
				inQuote = true
			} else {
				inQuote = false
				args = append(args, arg)
				arg = ""
			}
		} else {
			arg += string(char)
		}
	}
	if arg != "" {
		args = append(args, arg)
	}
	return args
}
