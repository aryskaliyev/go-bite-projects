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
	- A concurrent process holds exclusive rights to a resource at any one time.
- Wait For Condition
	- A concurrent process must simultaneously hold a resource and be waiting for an additional resouce.
- No Preemption
	- A resource held by a concurrent process can only be released by that process.
- Circular Wait
	- A concurrent process must be waiting on a chain of other concurrent processes, which are in turn waiting on it.

- Livelocks are programs that are actively performing concurrent operations, but these operations do nothing to move the state of the program forward.

#### Example:
```go
cadence := sync.NewCond(&sync.Mutex{})
go func() {
	for range time.Tick(1 * time.Millisecond) {
		cadence.Broadcast()
	}
}()

takeStep := func() {
	cadence.L.Lock()
	cadence.Wait()
	cadence.L.Unlock()
}

tryDir := func(dirName string, dir *int32, out *bytes.Buffer) bool {
	fmt.Fprintf(out, " %v", dirName)
	atomic.AddInt32(dir, 1)
	takeStep()
	if atomic.LoadInt32(dir) == 1 {
		fmt.Fprintf(out, ". Success!")
		return true
	}
	takeStep()
	atomic.AddInt32(dir, -1)
	return false
}

var left, right int32
tryLeft := func(out *bytes.Buffer) bool { return tryDir("left", &left, out)}
tryRight := func(out *bytes.Buffer) bool { return tryDir("right", &right, out)}

walk := func(walking *sync.WaitGroup, name string) {
	var out bytes.Buffer
	defer func() { fmt.Println(out.String()) }()
	defer walking.Done()
	fmt.Fprintf(&out, "%v is trying to scoot:", name)
	for i := 0; i < 5; i++ {
		if tryLeft(&out) || tryRight(&out) {
			return
		}
	}
	fmt.Fprintf(&out, "\n%v tosses her hands up in exasperation!", name)
}

var peopleInHallway sync.WaitGroup
peopleInHallway.Add(2)
go walk(&peopleInHallway, "Alice")
go walk(&peopleInHallway, "Barbara")
peopleInHallway.Wait()
```

- Starvation is any situation where a concurrent process cannot get all the resources it needs to perform work.

```go
var wg sync.WaitGroup
	var sharedLock sync.Mutex
	const runtime = 1 * time.Second

	greedyWorker := func() {
		defer wg.Done()

		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(3 * time.Nanosecond)
			sharedLock.Unlock()
			count++
		}

		fmt.Printf("Greedy worker was able to execute %v work loops\n", count)
	}

	politeWorker := func() {
		defer wg.Done()

		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			count++
		}

		fmt.Printf("Polite worker was able to execute %v work loops\n", count)
	}

	wg.Add(2)
	go greedyWorker()
	go politeWorker()

	wg.Wait()
```
