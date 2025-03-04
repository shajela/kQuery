package envutils

import (
	"fmt"
	"os"
)

func LookupEnv(k string) (string, error) {
	v, b := os.LookupEnv(k)
	if !b || v == `""` || len(v) == 0 {
		return "", fmt.Errorf("Please specify value for '%s'", k)
	}
	return v, nil
}
