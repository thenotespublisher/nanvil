package main

import "github.com/nspcc-dev/neo-go/pkg/interop/runtime"

func _deploy(_ any, isUpdate bool) {
	if !isUpdate {
		runtime.Log("nsmith go example deployed")
	}
}

// GetValue is invoked by scripts/test-nsmith-examples.sh.
func GetValue() string {
	return "nsmith-go-ok"
}
