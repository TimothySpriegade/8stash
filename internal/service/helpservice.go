package service

import "fmt"

const spacer = "------------------------------------------------------------------------------------------------------------------------------"
const formatString = "  %-20s %s\n"

func PrintHelp() {
    fmt.Println("Welcome to 8stash! A simple tool for stashing work-in-progress on remote branches.")
    fmt.Println(spacer)

    fmt.Println("Usage:")
    fmt.Println("  8stash [command] [arguments]")
    fmt.Println()

    fmt.Println("Available Commands:")
    fmt.Printf(formatString, "push", "Save current work-in-progress to a new stash branch (default command).")
    fmt.Printf(formatString, "pop <number?>", "Apply a stash, commit, and delete the remote stash branch.")
    fmt.Printf(formatString, "list", "List all available 8stash branches on the remote.")
    fmt.Printf(formatString, "drop <number>", "Delete a specific remote stash branch.")
    fmt.Printf(formatString, "cleanup", "Delete all remote stashes older than the configured retention period.")
    fmt.Printf(formatString, "help", "Show this help message.")
    fmt.Println(spacer)

    fmt.Println("Configuration:")
    fmt.Println("  8Stash can be configured via a `.8stash.yaml` file in your repository root.")
    fmt.Println("  Key options include:")
    fmt.Println("    - branch_prefix: Customize the prefix for stash branches (e.g., 'wip/').")
    fmt.Println("    - retention_days: Set the age for the 'cleanup' command.")
    fmt.Println("    - naming: Configure stash ID format (e.g., numeric length or UUID).")
    fmt.Println()
    fmt.Println("  For more details on configuration, see the README.md file.")
    fmt.Println(spacer)
}
