package main

import "fmt"

const (
	infoColor    = "\033[1;34m%s\033[0m"
	warningColor = "\033[1;33m%s\033[0m"
	errorColor   = "\033[1;31m%s\033[0m"
	traceColor   = "\033[0;36m%s\033[0m"
)

// LogError logs in a red color. Meant for error messages
func LogError(message string) {
	fmt.Printf(errorColor, "[ERROR]: "+message)
	fmt.Println()
}

// LogWarning logs in a yellow color. Meant for unexpected or suspicious behaviour messages
func LogWarning(message string) {
	fmt.Printf(warningColor, "[WARN]:  "+message)
	fmt.Println()
}

// LogInfo logs in a dark blue color. Meant for control messages
func LogInfo(message string) {
	fmt.Printf(infoColor, "[INFO]:  "+message)
	fmt.Println()
}

// LogTrace logs in a light blue color. Meant for debugging messages
func LogTrace(message string) {
	fmt.Printf(traceColor, "[TRACE]: "+message)
	fmt.Println()
}