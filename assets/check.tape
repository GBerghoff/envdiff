# VHS demo tape for envdiff
# Run: vhs demo.tape

Output check.gif

Set Shell "bash"
Set FontSize 16
Set Width 900
Set Height 600
Set Theme "Dracula"
Set Padding 20

# Typing speed (slower = more readable)
Set TypingSpeed 75ms

# Show the config file
Type "cat envdiff.yaml"
Enter
Sleep 4s

# Clear and run the check
Type "clear"
Enter
Sleep 1s

Type "../envdiff check -f ../envdiff.yaml"
Enter
Sleep 5s
