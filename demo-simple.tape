# SQL Network Manager Demo
# A beautiful TUI for managing Cloud SQL authorized networks

Output demo.gif

Set FontSize 16
Set Width 1000
Set Height 600
Set Theme "Catppuccin Mocha"

# Clear screen and show command
Clear
Type "# SQL Network Manager - Manage Cloud SQL networks with style"
Enter
Sleep 1s
Type "sql-network-manager"
Enter
Sleep 3s

# Simulate loading complete - show instance list
Type "my-prod"  # Search for instance
Sleep 2s
Enter  # Select instance
Sleep 2s

# In network view, add a new network
Type "a"  # Add network
Sleep 1s
Type "Home Office"
Tab
Type "192.168.1.100/32"
Enter
Sleep 2s

# Navigate back to instance list
Escape
Sleep 1s

# Exit
Type "q"
Sleep 500ms