Output demo.gif

Set FontSize 24
Set Width 1600
Set Height 800

Sleep 1  # Ensure readiness

# Run kQuery
Type "go run cmd/client/client.go"
Enter

Sleep 2

# First query
Type "What are the pods consuming the most CPU in my cluster?"
Enter

Wait@20s /.*Enter query \('exit' to leave\):.*/

Sleep 10  # Allow time to view the output

# Second query
Type "What is the average memory usage of pods in my cluster?"
Enter

Wait@20s /.*Enter query \('exit' to leave\):.*/

Sleep 12  # Allow time to view the output

# Exit kQuery
Type "exit"
Enter

Sleep 2
