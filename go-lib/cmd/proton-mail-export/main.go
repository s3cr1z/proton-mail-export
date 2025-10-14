package main

import (
        "fmt"
        "os"

        "github.com/ProtonMail/export-tool/internal/tui"
)

func main() {
        if err := tui.Run(); err != nil {
                fmt.Printf("Error: %v\n", err)
                os.Exit(1)
        }
}
