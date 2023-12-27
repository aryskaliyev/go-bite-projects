# Concurrency in Go

### Race Conditions
A race condition occurs when two or more operations must execute in the correct order, but the program has not been written so that this order is guaranteed to be maintained.

#### Example:
```go
var data int
go func() {
	data++
}()
if data == 0 {
	fmt.Printf("the value is %v.\n", data)	
}
```
