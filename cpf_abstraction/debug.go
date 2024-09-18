package cpf_abstraction

import (
	cpf "github.com/nacioboi/go_cpf/cpf_release"
)

func Debug_Printf(format string, a ...interface{}) {
	cpf.Printf(format, a...)
}
