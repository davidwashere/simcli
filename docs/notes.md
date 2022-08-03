# Notes

## time.Sleep minimum
`time.Sleep` on Windows could not sleep less than 16 ms

The table below shows real time a run was taking vs expected best case for printing a 3844 line file with a given delay

delay | real | expected (best case)
--- | --- | ---
1  | 0m59.638s | 3.844
5  | 0m59.677s | 19.220
10 | 0m59.699s | 38.440
11 | 0m59.730s | 42.284
12 | 0m59.711s | 46.128
13 | 0m59.661s | 49.972
14 | 0m59.743s | 53.816
15 | 1m8.219s  | 57.66
20 | 1m59.426s | 76.88 (1m16.880s)

To address the discrepancy a line `batch` (number of lines to print prior to sleeping) was added for any delays < 16ms, the math isn't perfect but the resulting performance was closer to whats expected. 

```go
batch := 1
if t.Delay > 0 && t.Delay <= 15 {
    batch = 16 - t.Delay
}

// each batch is followed bo a 16ms sleep
```

