package config

import "fmt"

var version = "0.0.0"

func Version() string {
	return fmt.Sprintf("grx/%s", version)
}
