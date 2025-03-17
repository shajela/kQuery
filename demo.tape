Output demo2.gif

Set FontSize 18
Set Width 1600
Set Height 800

Sleep 1  # Ensure readiness

# Run kQuery
Type "go run cmd/client/client.go"
Enter

Sleep 2

# First query
Type "Which pods are consuming the most CPU?"
Enter

Wait@20s /.*Enter query \('exit' to leave\):.*/

Sleep 8  # Allow time to view the output

# Exit kQuery
Type "exit"
Enter

Sleep 2
