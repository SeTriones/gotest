package main

import (
    "fmt"
    "os"
    "strconv"
    "sync"
    "time"
)

func recoverWriteToClosedChannel() {
    if recover() == nil {
        return
    }
    fmt.Fprintf(os.Stderr, "write to closed channel\n")
}

func main() {
    done := make(chan bool)
    result := make(chan string, 10)

    wg := sync.WaitGroup{}
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            defer recoverWriteToClosedChannel()
            str := strconv.Itoa(idx)
            if idx == 3 {
                time.Sleep(2 * time.Second)
            }
            result <- str
            fmt.Fprintf(os.Stderr, "write %v\n", idx)
        }(i)
    }
    go func() {
        wg.Wait()
        done <- true
    }()

    select {
    case <-done:
        fmt.Fprintf(os.Stderr, "DONE!\n")
    case <-time.After(1 * time.Second):
        fmt.Fprintf(os.Stderr, "TIMEOUT!\n")
    }
    fmt.Errorf("close channle\n")
    close(result)

    for val := range result {
        fmt.Fprintf(os.Stderr, "val=%s\n", val)
    }

    <-done
}
