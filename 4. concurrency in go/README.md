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

#### Example: Using *Broadcast* method to notify all registered handlers
```go
	type Button struct {
		Clicked *sync.Cond
	}
	button := Button{ Clicked: sync.NewCond(&sync.Mutex{}) }

	subscribe := func(c *sync.Cond, fn func()) {
		var goroutineRunning sync.WaitGroup
		goroutineRunning.Add(1)
		go func() {
			goroutineRunning.Done()
			c.L.Lock()
			defer c.L.Unlock()
			c.Wait()
			fn()
		}()
		goroutineRunning.Wait()
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing window.")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialog box!")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse clicked.")
		clickRegistered.Done()
	})

	button.Clicked.Broadcast()

	clickRegistered.Wait()
```

### Once
- *sync.Once* is a type that utilizes some *sync* primitives internally to ensure that only one call to *Do* ever calls the function passed in -- even on different goroutines.

#### Example:
```go
	var count int

	increment := func() {
		count++
	}

	var once sync.Once

	var increments sync.WaitGroup
	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
		}()
	}

	increments.Wait()
	fmt.Printf("Count is %d\n", count)
```

- *sync.Once* only counts the number of times *Do* is called, not how many times unique functions passed into *Do* are called.

#### Example:
```go
	var count int
	increment := func() { count++ }
	decrement := func() { count-- }

	var once sync.Once
	once.Do(increment)
	once.Do(decrement)

	fmt.Printf("Count: %d\n", count)
```

### Pool
- *Pool* is a concurrent-safe implementation of the object pool pattern.
- The pool pattern is a way to create and make available a fixed number, or pool, of things for use. It's commonly used to constrain the creation of things that are expensive (e.g. database connections) so that only a fixed number of them are ever created, but an indeterminate number of operations can still request accesss to these things. In the case of Go's *sync.Pool*, this data type can be safely used by multiple goroutines.

#### Example:
```go
        myPool := &sync.Pool{
                New: func() interface{} {
                        fmt.Println("Creating new instance.")
                        return struct{}{}
                },
        }

        myPool.Get()
        instance := myPool.Get()
        myPool.Put(instance)
        myPool.Get()
```

- Why use pool and not just instantiate objects as you go? Go has a garbage collector, so the instantiated objects will be automatically cleaned up.
#### Example:
```go
	var numCalcsCreated int
	calcPool := &sync.Pool {
		New: func() interface{} {
			numCalcsCreated += 1
			mem := make([]byte, 1024)
			return &mem
		},
	}

	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	const numWorkers = 1024 * 1024
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			mem := calcPool.Get().(*[]byte)
			defer calcPool.Put(mem)
		}()
	}

	wg.Wait()
	fmt.Printf("%d calculators were created.\n", numCalcsCreated)
```

- Another common situation where a *Pool* is useful is for warming a cache of pre-allocated objects for operations that must run as quickly as possible. In this case, instead of trying to guard the host machine's memory by constraining the number of objects created, we're trying to guard consumers' time by front-loading the time it takes to get a reference to another object. This is very common when writing high-throughput network servers that attempt to respond to requests as quickly as possible.

- When working with a *Pool*, just remember the following points:
	- When instantiating *sync.Pool*, give it a *New* member variable that is thread-safe when called.
	- When you receive an instance from *Get*, make no assumptions regarding the state of the object you receive back.
	- Make sure to call *Put* when you're finished with the object you pulled out of the pool. Otherwise, the *Pool* is useless. Usually this is done with *defer*.
	- Objects in the pool must be roughly uniform in makeup.

### Channels
- Channels are one of the synchronization primitives in Go derived from Hoare's CSP (Communicating Sequential Processes). While they are used to synchronize access of the memory, they are best used to communicate information between goroutines.

- Like a river, a channel serves as a conduit for a stream of information; values may be passed along the channel, and then read out downstream.

- Channels are the glue that binds goroutines together.

#### Example: Declaring a channel of empty interface
```go
	var dataStream chan interface{} // Declaring a channel of empty interface
	dataStream = make(chan interface{}) // Instantiating the channel using built-in *make* function
```

- Channels can also be declared to only support a unidirectional flow of data - that is, you can define a channel that only supports sending or receiving information.

#### Example: To both declare and instantiate a channel that can only read, place *<-* operator on the lefthand side
```go
	var dataStream <-chan interface{}
	dataStream := make(<-chan interface{})
```

#### Example: To declare and create a channel that can only send, place *<-* operator on the righthand side.
```go
	var dataStream chan<- interface{}
	dataStream := make(chan<- interface{})
```

- Go implicitly converts bidirectional channels to unidirectional channels when needed.
#### Example:
```go
	var receiveChan <-chan interface{}
	var sendChan chan<- interface{}
	dataStream := make(chan interface{})

	receiveChan = dataStream
	sendChan = dataStream
```

- Channels in Go are said to be *blocking*. This means that any goroutine that attempts to write to a channel that is full will wait until the channel has been emptied, and any goroutine that attempts to read from a channel that is empty will wait until at least one item is placed on it.

- If a goroutine making writes to a channel has knowledge of how many writes it will make, it can be useful to create a buffered channel whose capacity is the number of writes to be made, and then make those writes as quickly as possible.

#### Example: *Buffered Channel*
```go
        var stdoutBuff bytes.Buffer
        defer stdoutBuff.WriteTo(os.Stdout)

        intStream := make(chan int, 2)
        go func() {
                defer close(intStream)
                defer fmt.Fprintln(&stdoutBuff, "Producer Done.")
                for i := 0; i < 5; i++ {
                        fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
                        intStream <- i
                }


        }()
        for integer := range intStream {
                fmt.Fprintf(&stdoutBuff, "\tReceived %v.\n", integer)
        }
```

- Channel *ownership*. Channel owners have a write-access view into the channel (*chan* or *chan<-*), and channel utilizers only have a read-only view into the channel (*<-chan*).
- The goroutine that owns a channel should:
	- Instantiate a channel.
	- Perform writes, or pass ownership to another goroutine.
	- Close the channel.
	- Encapsulate the previous three things in this list and expose them via a reader channel.

- By assigning these responsibilities to channel owners, a few things happen:
	- Because we're the one initializing the channel, we remove the risk of deadlocking by writing to a *nil* channel.
	- Because we're the one initializing the channel, we remove the risk of *panic*ing by closing a *nil* channel.
	- Because we're the one who decides when the channel gets closed, we remove the risk of *panic*ing by writing to a closed channel.
	- Because we're the one who decides when the channel gets closed, we remove the risk of *panic*ing by closing a channel more than once.
	- We wield the type checker at compile time to prevent improper writes to our channel.

- As a consumer of a channel, ensure two things:
	- Knowing when a channel is closed.
	- Responsibly handling blocking for any reason.

#### Example:
```go
        chanOwner := func() <-chan int {
                resultStream := make(chan int, 5)
                go func() {
                        defer close(resultStream)
                        for i := 0; i <= 5; i++ {
                                resultStream <- i
                        }
                }()
                return resultStream
        }

        resultStream := chanOwner()
        for result := range resultStream {
                fmt.Printf("Received: %d\n", result)
        }
        fmt.Println("Done receiving!")
```

### The *select* statement
- The *select* statement is the glue that binds channels together; it's how we're able to compose channels together in a program to form larger abstractions.

- *select* statements can help safely bring channels together with concepts like cancellations, timouts, waiting, and default values.

#### Example: How to use *select* statements
```go
	var c1, c2 <-chan interface{}
	var c3 chan<- interface{}
	select {
	case <- c1:
		// Do something
	case <- c2:
		// Do something
	case c3<- struct{}{}:
		// Do something
	}
```

- Unlike *switch* blocks, *case* statements in *select* block aren't tested sequentially, and execution won't automatically fall through if none of the criteria are met. Instead, all channel reads and writes are considered simultaneously to see if any of them are ready: populated or closed channels in the case of reads, and channels that are not at capacity in the case of writes. If none of the channels are ready, the entrie *select* statement blocks. Then when one of the channels is ready, that operation will proceed, and its corresponding statements will execute.

#### Example:
```go
        start := time.Now()
        c := make(chan interface{})
        go func() {
                time.Sleep(5 * time.Second)
                close(c)
        }()

        fmt.Println("Blocking on read...")
        select {
        case <-c:
                fmt.Printf("Unblocked %v later.\n", time.Since(start))
        }
```

- What happens when multiple channels have something to read? The Go runtime will perform pseudo-random uniform selection over the set of case statements. This just means that of your set of case statements, each has an equal chance of being selected as all the others. By weighting the chance of each channel being utilized equally, all Go programs that utilize the *select* statement will perform well in the average case.

#### Example:
```go
        c1 := make(chan interface{}); close(c1)
        c2 := make(chan interface{}); close(c2)

        var c1Count, c2Count int
        for i := 1000; i >= 0; i-- {
                select {
                case <-c1:
                        c1Count++
                case <-c2:
                        c2Count++
                }
        }

        fmt.Printf("c1Count: %d\nc2Count: %d\n", c1Count, c2Count)
```

- What happens if there are never any channels that become ready? If there's nothing useful you can do when all the channels are blocked, but also you can't block forever, you may want to time out.

#### Example:
```go
	var c <-chan int
	select {
	case <-c:
	case <-time.After(1 * time.Second):
		fmt.Println("Timed out.")
	}
```

- What happens when no channel is ready, and we need to do something in the meantime? Like *case* statements, the *select* statement also allows for a *default* clause in case you'd like to do something if all the channels you're selecting against are blocking.

#### Example:
```go
        start := time.Now()
        var c1, c2 <-chan int
        select {
        case <-c1:
        case <-c2:
        default:
                fmt.Printf("In default after %v\n", time.Since(start))
        }
```

- Usually you'll see a *default* clause used in conjunction with a for-select loop. This allows a goroutine to make progress on work while waiting for another goroutine to report a result.

#### Example: 
```go
        done := make(chan interface{})
        go func() {
                time.Sleep(5 * time.Second)
                close(done)
        }()

        workCounter := 0
        loop: // labelling a for loop
        for {
                select {
                case <-done:
                        break loop
                default:
                }

                // Simulate work
                workCounter++
                time.Sleep(1 * time.Second)
        }

        fmt.Printf("Achieved %v cycles of work before signalled to stop.\n", workCounter)
```

- There is a special case for empty *select* statements: *select* statements with no *case* clauses: `select {}` -- they will block forever.

## Concurrency Patterns in Go

### Confinement

- Safe operation options:
	- Synchronization primitives for sharing memory (e.g., *sync.Mutex*)
	- Synchronization via communicating (e.g., channels)

- Implicitly safe options within multiple concurrent processes:
	- Immutable data
		- Immutable data is ideal because it is implicitly concurrent-safe. Each concurrent process may operate on the same data, but it may not modify it. If it wants to create new data, it must create a new copy of the data with the desired modifications.
	- Data protected by confinement
		- Confinement is the simple yet powerful idea of ensuring information is only ever available from *one* concurrent process. When this is achieved, a concurrent program is implicitly safe and no synchronization is needed. There are two kinds of confinement possible: ad hoc and lexical.

#### Example: Ad Hoc Confinement
```go
        data := make([]int, 4)

        loopData := func(handleData chan<- int) {
                defer close(handleData)
                for i := range data {
                        handleData <- data[i]
                }
        }

        handleData := make(chan<- int)
        go loopData(handleData)

        for num := range handleData {
                fmt.Println(num)
        }
```

- Lexical confinement involves using lexical scope to expose only the correct data and concurrency primitives for multiple concurrent processes to use. It makes it impossible to do the wrong thing.

#### Example: Lexical Confinement
```go
	chanOwner := func() <-chan int {
		results := make(chan int, 5)
		go func() {
			defer close(results)
			for i := 0; i <= 5; i++ {
				results <- i
			}
		}()
		return results
	}

	consumer := func(results <-chan int) {
		for result := range results {
			fmt.Printf("Received: %d\n", result)
		}
		fmt.Println("Done receiving!")
	}

	results := chanOwner()
	consumer(results)
```

#### Example: Lexical Confinement 
```go
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()

		var buff bytes.Buffer
		for _, b := range data {
			fmt.Fprintf(&buff, "%c", b)
		}
		fmt.Println(buff.String())
	}

	var wg sync.WaitGroup
	wg.Add(2)
	data := []byte("golang")
	go printData(&wg, data[:3])
	go printData(&wg, data[3:])

	wg.Wait()
```

### The *for-select* Loop

#### Example:
```go
	for { // Either loop infinitely or range over something
		select {
		// Do some work with channels
		}
	}
```

#### Example: Sending iteration variables on a channel
```go
	for _, s := range []string{"a", "b", "c"} {
		select {
		case <-done:
			return
		case stringStream <- s:
		}
	}
```

#### Example: Looping infinitely waiting to be stopped
```go
	for {
		select {
		case <-done:
			return
		default:
		}
		
		// Do non-preemtable work
	}

	for {
		select {
		case <-done:
			return
		default:
			// Do non-preemptable work
		}
	}
```

### Preventing Goroutine Leaks
- Goroutines are cheap and easy to create, the runtime handles multiplexing the goroutines onto any number of operating system threads so that we don't often have to worry about that level of abstraction. But they *do* cost resources, and goroutines are not garbage collected by the runtime.

- Goroutines represent units of work that may or may not run in parallel with each other. The goroutine has a few paths to termination:
	- When it has completed its work.
	- When it cannot continue its work due to an unrecoverable error.
	- When it's told to stop working.

- The main goroutine passes a *nil* channel into *doWork*. Therefore, the *strings* channel will never actually gets any strings written onto int, and the goroutine containing *doWork* will remain in memory for the lifetime of this process (we would even deadlock if we joined the goroutine within *doWork* and the main goroutine). See the example below.

#### Example: Goroutine Leak
```go
	doWork := func(strings <-chan string) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for s := range strings {
				// Do something interesting
				fmt.Println(s)
			}
		}()
		return completed
	}

	doWork(nil)
	// Perhaps more work is done here
	fmt.Println("Done.")
```

- To successfully mitigate the goroutine leak, we need to establish a signal between the parent goroutine and its children that allows the parent to signal cancellation to its children. By convention, this signal is usually a read-only channel named *done*. The parent goroutine passes this channel to the child goroutine and then closes the channel when it wants to cancel the child goroutine.

#### Example: Goroutines receiving on a channel
```go
	doWork := func(
		done <-chan interface{},
		strings <-chan string,
	) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(terminated)
			for {
				select {
				case s := <-strings:
					// Do something interesting
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()
		return terminated
	}

	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		// Cancel the operation after 1 second.
		time.Sleep(1 * time.Second)
		fmt.Println("Canceling doWork goroutine...")
		close(done)
	}()

	<-terminated
	fmt.Println("Done.")

```

#### Example: Goroutine blocked on attempting to write a value to a channel
```go
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)
			for {
				select {
				case randStream <- rand.Int():
				case <-done:
					return
				}
			}
		}()

		return randStream
	}

	done := make(chan interface{})
	randStream := newRandStream(done)
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}
	close(done)

	// Simulate ongoing work
	time.Sleep(1 * time.Second)
```

- *If a goroutine is responsible for creating a goroutine, it is also responsible for ensuring it can stop the goroutine.*
- How we ensure goroutines are able to be stopped can differ depending on the type and purpose of goroutine, but they all build on the foundation of passing in a *done* channel.

### The *or-channel*
- If you can't know the number of *done* channels you're working with at runtime, or if you prefer a one-liner, you can combine these channels together using the *or-channel* pattern.

#### Example: Composite *done* channel through recursion and goroutines
```go
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
			case 0:
				return nil
			case 1:
				return channels[0]
		}

		orDone := make(chan interface{})
		go func() {
			defer close(orDone)

			switch len(channels) {
			case 2:
				select {
					case <-channels[0]:
					case <-channels[1]:
				}
			default:
				select {
					case <-channels[0]:
					case <-channels[1]:
					case <-channels[2]:
					case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}

	sig := func(after time.Duration) <-chan interface{} {
		fmt.Println(after)
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2 * time.Hour),
		sig(5 * time.Minute),
		sig(1 * time.Second),
		sig(1 * time.Hour),
		sig(1 * time.Minute),
	)
	fmt.Printf("done after %v\n", time.Since(start))
```

### Error Handling
- "Who should be responsible for handling the error?"
- Separation of concerns: in general, concurrent processes should send their errors to another part of the program that has complete information about the state of the program, and can make a more informed decision about what to do.

- The key thing is that we've coupled the potential result with the potential error. The main takeeway is that errors should be considered first-class citizens when constructing values to return from goroutines. If your goroutine can produce errors, those errors should be tightly coupled with your result type, and passed along through the same lines of communication -- just like regular synchronous functions.

#### Example:
```go
	type Result struct {
		Error error
		Response *http.Response
	}

	checkStatus := func(
		done <-chan interface{},
		urls ...string,
	) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)
			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
					case <-done:
						return
					case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	errCount := 0
	urls := []string{"a", "https://www.google.com", "https://badhost", "https://www.youtube.com", "b", "c", "d"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)
			errCount++
			if errCount >= 3 {
				fmt.Println("Too many errors, breaking!")
				break
			}
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
```

### Pipelines
- A *pipeline* is just another tool you can use to form an abstraction in your system.
- A pipeline is nothing more than a series of things that take data in, perform an operation on it, and pass the data back out. Each of these operations is called a *stage* of the pipeline.
- A stage is just something that takes data in, performs a transformation on it, and sends the data back out.


#### Example: 
```go
	// A pipeline stage example
	multiply := func(values []int, multiplier int) []int {
		multipliedValues := make([]int, len(values))
		for i, v := range values {
			multipliedValues[i] = v * multiplier
		}
		return multipliedValues
	}

	// A pipeline stage example
	add := func(values []int, additive int) []int {
		addedValues := make([]int, len(values))
		for i, v := range values {
			addedValues[i] = v + additive
		}
		return addedValues
	}

	ints := []int{1, 2, 3, 4}
	for _, v := range multiply(add(multiply(ints, 2), 1), 2) {
		fmt.Println(v)
	}
```
- Properties of pipeline stage:
	- A stage consumes and returns the same type.
	- A stage must be reified (define variables that have a type of a function signature) by the language so that it may be passed around.

### Best Practices for Constructing Pipelines

#### Example:
```go
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for _, i := range integers {
				select {
					case <-done:
						return
					case intStream <- i:
				}
			}
		}()
		return intStream
	}

	multiply := func(
		done <-chan interface{},
		intStream <-chan int,
		multiplier int,
	) <-chan int {
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
					case <-done:
						return
					case multipliedStream <- i * multiplier:
				}
			}
		}()
		return multipliedStream
	}

	add := func(
		done <-chan interface{},
		intStream <-chan int,
		additive int,
	) <-chan int {
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
					case <-done:
						return
					case addedStream <- i + additive:
				}
			}
		}()
		return addedStream
	}

	done := make(chan interface{})
	defer close(done)

	intStream := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 2)

	for v := range pipeline {
		fmt.Println(v)
	}
```

- Using channels allows us two things:
	- At the end of our pipeline, we use a range statement to extract the values, and at each stage we can safely execute concurrently because our inputs and outputs are safe in concurrent contexts.
	- Each stage of the pipeline is executing concurrently. This means that any stage only needs wait for its inputs, to be able to send its outputs.

- At the beginning of the pipeline, we've established that we must convert discrete values into a channel. There are two points in this process that *must* be preemptable:
	- Creation of the discrete value that is not nearly instantaneous.
	- Sending of the discrete value on its channel.

### Some Handy Generators
- A generator for a pipeline is any function that converts a set of discrete values into a stream of values on a channel.

#### Example:
```go
	repeat := func(
		done <-chan interface{},
		values ...interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
						case <-done:
							return
						case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}

	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
					case <-done:
						return
					case takeStream <- <- valueStream:
				}
			}
		}()
		return takeStream
	}

	done := make(chan interface{})
	defer close(done)

	for num := range take(done, repeat(done, 1), 10) {
		fmt.Printf("%v ", num)
	}
```

#### Example:
```go
	take := func(
                done <-chan interface{},
                valueStream <-chan interface{},
                num int,
        ) <-chan interface{} {
                takeStream := make(chan interface{})
                go func() {
                        defer close(takeStream)
                        for i := 0; i < num; i++ {
                                select {
                                        case <-done:
                                                return
                                        case takeStream <- <- valueStream:
                                }
                        }
                }()
                return takeStream
        }

	repeatFn := func(
		done <-chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				select {
					case <-done:
						return
					case valueStream <- fn():
				}
			}
		}()
		return valueStream
	}

	done := make(chan interface{})
	defer close(done)

	rand := func() interface{} { return rand.Int() }

	for num := range take(done, repeatFn(done, rand), 10) {
		fmt.Println(num)
	}
```

#### Example:
```go
	repeat := func(
		done <-chan interface{},
		values ...interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
						case <-done:
							return
						case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}

	take := func(
                done <-chan interface{},
                valueStream <-chan interface{},
                num int,
        ) <-chan interface{} {
                takeStream := make(chan interface{})
                go func() {
                        defer close(takeStream)
                        for i := 0; i < num; i++ {
                                select {
                                        case <-done:
                                                return
                                        case takeStream <- <- valueStream:
                                }
                        }
                }()
                return takeStream
        }

	toString := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) <-chan string {
		stringStream := make(chan string)
		go func() {
			defer close(stringStream)
			for v := range valueStream {
				select {
					case <-done:
						return
					case stringStream <- v.(string):
				}
			}
		}()
		return stringStream
	}

	done := make(chan interface{})
	defer close(done)

	var message string
	for token := range toString(done, take(done, repeat(done, "I", "am."), 5)) {
		message += token
	}

	fmt.Printf("message: %s...", message)
```

### Fan-Out, Fan-In

- *Fan-out* is a term to describe the process of starting multiple goroutines to handle input from the pipeline, and *fan-in* is a term to describe the process of combining multiple results into one channel.

- You might consider fanning out one of your stages if **both** of the following apply:
	- It doesn't rely on values that the stage had calculated before.
	- It takes a long time to run.

- The property of order-independence is important because you have no guarantee in what order concurrent copies of your stage will run, nor in what order they will return.

#### Example:
```go
	toInt := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()
		return intStream
	}

	repeatFn := func(
		done <-chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				select {
				case <-done:
					return
				case valueStream <- fn():
				}
			}
		}()
		return valueStream
	}

	take := func(
		done <-chan interface{},
		valueStream <-chan int,
		num int,
	) <-chan int {
		takeStream := make(chan int)
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <- valueStream:
				}
			}
		}()
		return takeStream
	}

	primeFinder := func(
		done <-chan interface{},
		intStream <-chan int,
	) <-chan int {
		primeStream := make(chan int)
		go func() {
			defer close(primeStream)
			for v := range intStream {
				isPrime := true
				for i := v-1; i >= 2; i-- {
					if v%i == 0 {
						isPrime = false
						break
					}
				}
				if isPrime {
					select {
					case <-done:
						return
					case primeStream <- v:
					}
				}
			}

		}()
		return primeStream
	}

	rand := func() interface{} { return rand.Intn(50000000) }

	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	randIntStream := toInt(done, repeatFn(done, rand))
	fmt.Println("Primes:")
	for prime := range take(done, primeFinder(done, randIntStream), 10) {
		fmt.Printf("\t%d\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
```

- Fanning in means *multiplexing* or joining together multiple streams of data into a single stream.

- Fanning in involves creating the multiplexed channel consumers will read from, and then spinning up one goroutine for each incoming channel, and one goroutine to close the multiplexed channel when the incoming channels have all been closed. Since we're going to be creating a goroutine that is waiting on *N* other goroutines to complete, it makes sense to create a *sync.WaitGroup* to coordinate things. The *multiplex* function also notifies the *WaitGroup* that it's done.

#### Example:
```go
	fanIn := func(
		done <-chan interface{},
		channels ...<-chan interface{},
	) <-chan interface{} {
		var wg sync.WaitGroup
		multiplexedStream := make(chan interface{})

		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}

		// Select from all the channels
		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		// Wait for all the reads to complete
		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()

		return multiplexedStream
	}

	repeatFn := func(
		done <-chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				select {
				case <-done:
					return
				case valueStream <- fn():
				}
			}
		}()
		return valueStream
	}

	toInt := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()
		return intStream
	}

	primeFinder := func(
		done <-chan interface{},
		intStream <-chan int,
	) <-chan interface{} {
		primeStream := make(chan interface{})
		go func() {
			defer close(primeStream)
			for v := range intStream {
				isPrime := true
				for i := v-1; i >= 2; i-- {
					if v%i == 0 {
						isPrime = false
						break
					}
				}
				if isPrime {
					select {
					case <-done:
						return
					case primeStream <- v:
					}
				}
			}
		}()
		return primeStream
	}

	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <- valueStream:
				}
			}
		}()
		return takeStream
	}


	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	rand := func() interface{} { return rand.Intn(50000000) }

	randIntStream := toInt(done, repeatFn(done, rand))

	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)
	finders := make([]<-chan interface{}, numFinders)
	fmt.Println("Primes:")
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	for prime := range take(done, fanIn(done, finders...), 10) {
		fmt.Printf("\t%d\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
```

- A **naive** implementation of the fan-in, fan-out algorithm only works if the order in which results arrive is unimportant.

### The *or-done*-channel

- At times you will be working with channels from disparate parts of your system. Unlike with pipelines, you can't make any assertions about how a channel will behave when code you're working with is canceled via its *done* channel. That is to say, you don't know if the fact that your goroutine was canceled means the channel you're reading from will have been canceled.

#### Example:
```go
	for val := range myChan {
		// Do something with val
	}

	// -----------------------------
	
	loop:
	for {
		select {
		case <-done:
			break loop
		case maybeVal, ok := <-myChan:
			if ok == false {
				return // or maybe break from for
			}
			// Do something with maybeVal
		}
	}
```

#### Example:
```go
        orDone := func(done, c <-chan interface{}) <-chan interface{} {
                valStream := make(chan interface{})
                go func() {
                        defer close(valStream)
                        for {
                                select {
                                case <-done:
                                        return
                                case v, ok := <-c:
                                        if ok == false {
                                                return
                                        }
                                        select {
                                        case valStream <- v:
                                        case <-done: 
                                        }       
                                }
                        }
                }()
                return valStream
        }

	// -----------------------------

	for val := range orDone(done, myChan) {
		// Do something with val
	}	
```

### The *tee*-channel

- Sometimes you may want to split values coming in from a channel so that you can send them off into two separate areas of your codebase. Imagine a channel of user commands: you might want to take in a stream of user commands on a channel, send them to something that executes them, and also send them to something that logs the commands for later auditing.

#### Example: Pass a channel to read from, and it will return two separate channels that will get the same value
```go
	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <- valueStream:
				}
			}
		}()
		return takeStream
	}

	repeat := func(
		done <-chan interface{},
		values ...interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for _, v := range values {
				select {
				case <-done:
					return
				case valueStream <- v:
				}
			}
		}()
		return valueStream
	}

	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	tee := func(
		done <-chan interface{},
		in <-chan interface{},
	) (<-chan interface{}, <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})
		go func() {
			defer close(out1)
			defer close(out2)
			for val := range orDone(done, in) {
				var out1, out2 = out1, out2
				for i := 0; i < 2; i++ {
					select {
					case <-done:
					case out1 <-val:
						out1 = nil
					case out2 <-val:
						out2 = nil
					}
				}
			}
		}()
		return out1, out2
	}

	done := make(chan interface{})
	defer close(done)

	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 20))

	for val1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", val1, <-out2)
	}
```

### The *bridge*-channel
- In some cases, you may find yourself wanting to consume values from a sequence of channels: `<-chan <-chan interface{}`. If we define a function that can destructure the channel of channels into a simple channel - a technique called *bridging* the channels - this will make much easier for the consumer to focus on the problem at hand.

#### Example:
```go
	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	bridge := func(
		done <-chan interface{},
		chanStream <-chan <-chan interface{},
	) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				var stream <-chan interface{}
				select {
				case maybeStream, ok := <-chanStream:
					if ok == false {
						return
					}
					stream = maybeStream
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					select {
					case valStream <- val:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}

	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
```

### Queuing
- Sometimes it's useful to begin accepting work for your pipeline even though the pipeline is not yet ready for more. This process is called *queuing*. All this means is that once your stage has completed some work, it stores it in a temporary location in memory so that other stages can retrieve it later, and your stage doesn't need to hold a reference to it. While introducing queuing into your system is very useful, it's usually one of the last techniques you want to employ when optimizing your program. Adding queuing prematurely can hide synchronization issues such as deadlocks and livelocks, and further, as your program converges toward correctness, you may find that you need more or less queuing.

- Queuing will almost never speed up the total runtime of your program; it will only allow the program to behave differently.

- The utility of introducing a queue isn't that the runtime of one of stages has been reduced, but rather that the time it's in a *blocking state* is reduced. This allows the stage to continue doing its job.

- The true utility of queues is to *decouple stages* so that the runtime of one stage has no impact on the runtime of another. Decoupling stages in this manner then cascades to alter the runtime behavior of the system as a whole, which can be either good or bad depending on your system.

- Situations in which queuing *can* increase the overall performance of your system:
	- If batching requests in a stage saves time.
	- If delays in a stage produce a feedback loop into system.

- Queuing should be implemented either:
	- At the entrance to your pipeline.
	- In stages where batching will lead to higher efficiency.

### The *context* package
- *context* package:
```go
	var Canceled = errors.New("context canceled")
	var DeadlineExceeded error = deadlineExceededError{}

	type CancelFunc
	type Context

	func Background() Context
	func TODO() Context
	func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
	func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
	func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
	func WithValue(parent Context, key, val interface{}) Context
```

- *Context* type is the type that will flow through your system much like a *done* channel does.

- *Context* type:
```go
	type Context interface {
		// Deadline returns the time when work done on behalf of this
		// context should be canceled. Deadline returns ok==false when no
		// deadline is set. Successive calls to Deadline return the same results.
		Deadline() (deadline time.Time, ok bool)

		// Done returns a channel that's closed when work done on behalf
		// of this context should be canceled. Done may return nil if this
		// context can never be canceled. Successive calls to Done return
		// the same value.
		Done() <-chan struct{}

		// Err returns a non-nil error value after Done is closed. Err
		// returns Canceled if the context was canceled or
		// DeadlineExceeded if the context's deadline passed. No other
		// values for Err are defined. After Done is closed, successive
		// calls to Err return the same value.
		Err() error

		// Value returns the value associated with this context for key,
		// or nil if no value is associated with key. Successive calls to
		// Value with the same key returns the same result.
		Value(key interface{}) interface {}
	}
```

- A *Done* method which returns a channel that's closed when our function is to be preemted (interrupted or stopped).
- A *Deadline* function to indicate if a goroutine will be canceled after a certain time.
- An *Err* method that will return non-nil if the goroutine was canceled.

- The *context* package serves two primary purposes:
	- To provide an API for canceling branches of your call-graph.
	- To provide a data-bag for transporting request-scoped data through your call-graph.

- **Cancellation** in a function has three aspects:
	- A goroutine's parent may want to cancel it.
	- A goroutine may want to cancel its children.
	- Any blocking operations within a goroutine need to be preemptable so that it may be canceled.

- Some *context* package functions:
```go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
```

- *WithCancel* returns a new *Context* that closes its *done* channel when the returned *cancel* function is called.
- *WithDeadline* returns a new *Context* that closes its *done* channeld when the machine's clock advances past the given *deadline*.
- *WithTimeout* returns a new *Context* that closes its *done* channel after the given *timeout* duration.

- Instances of *context.Context* may look equivalent from the outside, but internally they may change at every stack-frame. For this reason, it's important to always pass instances of *Context* into your functions.

- To start the chain, the *context* package provides you with two functions to create empty instances of *Context*:
```go
func Background() Context
func TODO() Context
```

- *Background* simply returns an empty *Context*.
- *TODO* is not meant for use in production, but also returns an empty *Context*; *TODO*'s intended purpose is to serve as a placeholder for when you don't know which *Context* to utilize, or if you expect your code to be provided with a *Context*, but the upstream code hasn't yet furnished one.

#### Example: *done* channel pattern
```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	done := make(chan interface{})
	defer close(done)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Wait()
}

func printGreeting(done <-chan interface{}) error {
	greeting, err := genGreeting(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewell(done <-chan interface{}) error {
	farewell, err := genFarewell(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreeting(done <-chan interface{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func genFarewell(done <-chan interface{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(done <-chan interface{}) (string, error) {
	select {
	case <-done:
		return "", fmt.Errorf("canceled")
	case <-time.After(1*time.Minute):
	}
	return "EN/US", nil
}
```

#### Example: using the *context* package instead of a *done* channel
```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
}

func printGreeting(ctx context.Context) error {
	greeting, err := genGreeting(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewell(ctx context.Context) error {
	farewell, err := genFarewell(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreeting(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsuppoerted locale")
}

func genFarewell(ctx context.Context) (string, error) {
	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(1*time.Minute):
	}
	return "EN/US", nil
}
```

#### Example: Demonstrates using the *context.Context*'s *Deadline* method
```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
}

func printGreeting(ctx context.Context) error {
	greeting, err := genGreeting(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewell(ctx context.Context) error {
	farewell, err := genFarewell(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreeting(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func genFarewell(ctx context.Context) (string, error) {
	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(ctx context.Context) (string, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if deadline.Sub(time.Now().Add(1*time.Minute)) <= 0 {
			return "", context.DeadlineExceeded
		}
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(1*time.Minute):
	}
	return "EN/US", nil
}
```
