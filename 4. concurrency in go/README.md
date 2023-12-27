# Concurrency in Go

### Race Conditions
- A race condition occurs when two or more operations must execute in the correct order, but the program has not been written so that this order is guaranteed to be maintained.

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

### Deadlocks, Livelocks and Starvation
- A deadlocked program is one in which all concurrent processes are waiting on one another.

#### Example:
```go
type value struct {
	mu sync.Mutex
	value int
}

var wg sync.WaitGroup
printSum := func(v1, v2 *value) {
	defer wg.Done()
	v1.mu.Lock()
	defer v1.mu.Unlock()

	time.Sleep(2 * time.Second)
	v2.mu.Lock()
	defer v2.mu.Unlock()

	fmt.Printf("sum=%v\n", v1.value + v2.value)
}

var a, b value
wg.Add(2)
go printSum(&a, &b)
go printSum(&b, &a)
wg.Wait()
```

#### *Coffman Conditions* must be present for deadlocks to arise:
- Mutual Exclusion
	- A concurrenct process holds exclusive rights to a resource at any one time.
- Wait For Condition
	- A concurrent process must simultaneously hold a resource and be waiting for an additional resouce.
- No Preemption
	- A resource held by a concurrent process can only be released by that process.
- Circular Wait
	- A concurrent process must be waiting on a chain of other concurrent processes, which are in turn waiting on it.
