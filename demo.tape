# PIAM Admin Network Configurator Demo
# This tape demonstrates the PIAM ANC TUI application

Output demo.gif

Set FontSize 16
Set Width 1400
Set Height 900
Set Theme "Catppuccin Mocha"
Set PlaybackSpeed 0.75

# Launch the application
Type "./piam-anc"
Enter
Sleep 2s

# Show loading screen
Sleep 3s

# Show main resource list
Screenshot main-screen.png
Sleep 2s

# Start fuzzy search for 'dev'
Type "/"
Sleep 0.5
Type "dev"
Sleep 1.5
Screenshot search-dev.png

# Navigate to provision-dev-postgres-1 (should be first result)
Down
Down
Down
Sleep 1s
Screenshot selected-postgres.png

# Select the database instance
Enter
Sleep 2s

# Show network view with existing networks
Screenshot network-view.png
Sleep 2s

# Add new network
Type "a"
Sleep 1s

# Show add network form (should auto-populate with user info)
Screenshot add-network-form.png
Sleep 1s

# Clear the name field and enter demo data
Ctrl+A
Type "Demo Network"
Tab
Sleep 0.5

# Clear IP field and enter demo IP
Ctrl+A
Type "10.11.12.13/32"
Sleep 1s
Screenshot filled-form.png

# Submit the network addition
Enter
Sleep 3s

# Show updated network list with new entry
Screenshot updated-networks.png
Sleep 2s

# Show console link
Screenshot console-link.png
Sleep 1s

# Go back to main list
Escape
Sleep 1s

# Show final main screen
Screenshot final-main.png
Sleep 1s

# Exit the application
Type "q"
Sleep 1s