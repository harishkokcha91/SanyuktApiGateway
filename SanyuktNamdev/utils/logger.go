package utils

import (
	"log"
	"os"
)

// Logger is the standard logger for the application
var Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)