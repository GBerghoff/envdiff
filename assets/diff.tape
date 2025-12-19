# VHS demo tape for envdiff diff
# Run: vhs diff.tape

Output diff.gif

Set Shell "bash"
Set FontSize 16
Set Width 900
Set Height 600
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 75ms

# Compare and save to file, then render
Type "../envdiff compare local.json ci.json -o diff.json && ../envdiff render diff.json"
Enter
Sleep 5s
