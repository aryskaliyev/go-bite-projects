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

### Goroutines
- A goroutine is a function that is running concurrently (not necessarily in parallel!) alongside other code.

- Concurrency is a property of the code; parallelism is a property of the running program. We do not write parallel code, only concurrent code that we *hope* will be run in parallel.

- Goroutines - are unique to Go - they're a higher level of abstrction known as *coroutines*. Coroutines are simply concurrent subroutines (functions, closures, or methods in Go) that are *nonpreemptive* - that is, they cannot be interrupted. Instead, coroutines have multiple points throughout which allow for suspension or reentry.

#### Example:

```go
	var wg sync.WaitGroup
	sayHello := func() {
		defer wg.Done()
		fmt.Println("hello")
	}
	wg.Add(1)
	go sayHello()
	wg.Wait()
```

#### Example: Goroutines are not garbage collected

```go
	memConsumed := func() uint64 {
                runtime.GC()
                var s runtime.MemStats
                runtime.ReadMemStats(&s)
                return s.Sys
        }

        var c <-chan interface{}
        var wg sync.WaitGroup
        noop := func() { wg.Done(); <-c }

        const numGoroutines = 1e4
        wg.Add(numGoroutines)
        before := memConsumed()
        for i := numGoroutines; i > 0; i-- {
                go noop()
        }
        wg.Wait()
        after := memConsumed()
        fmt.Printf("%.3fkb\n", float64(after-before)/numGoroutines/1000)
```

### The *sync* package
- The *sync* package contains the concurrency primitives that are most useful for low-level memory access synchronization.

#### WaitGroup
- *WaitGroup* is a great way to wait for a set of concurrent operations to complete when you either don't care about the result of the concurrent operation, or you have other means of collecting their results.

#### Example:
```go
        var wg sync.WaitGroup

        wg.Add(1)
        go func() {
                defer wg.Done()
                fmt.Println("1st goroutine sleeping...")
                time.Sleep(1)
        }()

        wg.Add(1)
        go func() {
                defer wg.Done()
                fmt.Println("2nd goroutine sleeping...")
                time.Sleep(2)
        }()

        wg.Wait()
        fmt.Println("All goroutines complete.")
```

#### Mutex and RWMutex
- *Mutex* stands for "mutual exclusion" and is a way to guard critical sections of your program. A critical section is an area of your program that requires exclusive access to a shared resource. A *Mutex* provides a concurrent-safe way to express exclusive access to these shared resources.

- Critical sections are so named because they reflect a bottleneck in your program. It is somewhat expensive to enter and exit a critical section, and so generally people attempt to minimize the time spent in critical sections.

#### Example:
```go
        var count int
        var lock sync.Mutex

        increment := func() {
                lock.Lock()
                defer lock.Unlock()
                count++
                fmt.Printf("Incrementing: %d\n", count)
        }

        decrement := func() {
                lock.Lock()
                defer lock.Unlock()
                count--
                fmt.Printf("Decrementing: %d\n", count)
        }

        var arithmetic sync.WaitGroup
        for i := 0; i <= 5; i++ {
                arithmetic.Add(1)
                go func() {
                        defer arithmetic.Done()
                        increment()
                }()
        }

        for i := 0; i <= 5; i++ {
                arithmetic.Add(1)
                go func() {
                        defer arithmetic.Done()
                        decrement()
                }()
        }

        arithmetic.Wait()
        fmt.Println("Arithmetic complete.")
```

- The *sync.RWMutex* is conceptually the same thing as *Mutex*: it guards access to memory; however, *RWMutex* gives you a little bit more control over the memory. You can request a lock for reading, in which case you will be granted access unless the lock is being held for writing. This means that an arbitrary number of readers can hold a reader lock so long as nothing else is holding a writer lock.

#### Example:
```go
producer := func(wg *sync.WaitGroup, l sync.Locker) {
                defer wg.Done()
                for i := 5; i > 0; i-- {
                        l.Lock()
                        l.Unlock()
                        time.Sleep(1)
                }
        }

        observer := func(wg *sync.WaitGroup, l sync.Locker) {
                defer wg.Done()
                l.Lock()
                defer l.Unlock()
        }

        test := func(count int, mutex, rwMutex sync.Locker) time.Duration {
                var wg sync.WaitGroup
                wg.Add(count+1)
                beginTestTime := time.Now()
                go producer(&wg, mutex)
                for i := count; i > 0; i-- {
                        go observer(&wg, rwMutex)
                }

                wg.Wait()
                return time.Since(beginTestTime)
        }

        tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
        defer tw.Flush()

        var m sync.RWMutex
        fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
        for i := 0; i < 20; i++ {
                count := int(math.Pow(2, float64(i)))
                fmt.Fprintf(
                        tw,
                        "%d\t%v\t%v\n",
                        count,
                        test(count, &m, m.RLocker()),
                        test(count, &m, &m),
                )
        }
```

### Cond
- `...a rendezvous point for goroutines waiting for or announcing the occurrence of an event.` An "event" is any arbitrary signal between two or more goroutines that carries no information other than the fact that it has occurred. Very often you'll want to wait for one of these signals before continuing execution on a goroutine.

#### Example:
```go
        c := sync.NewCond(&sync.Mutex{})
        queue := make([]interface{}, 0, 10)

        removeFromQueue := func(delay time.Duration) {
                time.Sleep(delay)
                c.L.Lock()
                queue = queue[1:]
                fmt.Println("Removed from queue")
                c.L.Unlock()
                c.Signal()
        }

        for i := 0; i < 10; i++ {
                c.L.Lock()
                for len(queue) == 2 {
                        c.Wait()
                }
                fmt.Println("Adding to queue")
                queue = append(queue, struct{}{})
                go removeFromQueue(1 * time.Second)
                c.L.Unlock()
        }
```
