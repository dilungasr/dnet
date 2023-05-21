package logger

import (
	"fmt"
	"time"
)

func Log(info any) {
	fmt.Printf("[%v] %v\n", time.Now().String(), info)
}
