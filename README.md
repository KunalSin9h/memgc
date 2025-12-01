# memgc

Simple continuous memory info aggregator from **GoLang** `gc trace`

A single cycle of GC Trace looks like:

```ocaml
gc 3 @3.182s 0%: 0.015+0.59+0.096 ms clock, 0.19+0.10/1.3/3.0+1.1 ms cpu, 4->4->2 MB, 5 MB goal, 12 P
```

This line is printed every time the Go garbage collector runs and shows detailed metrics about that specific **GC cycle**.

## Data Aggregation

Installation:

```bash
go install github.com/kunalsin9h/memgc@latest
```

Usage:

```bash
GODEBUG=gctrace=1 ./bin 2> >(memgc --csv data.csv)
```

This will output a `data.csv` file.

## Visualization

Load the `data.csv` file in graph plotting tools like [Draxlr](https://www.draxlr.com/tools/line-chart-generator/)

### CPUPercent

Shows "Total percentage of CPU time spent in GC since the program started"

<img width="1404" height="840" alt="image" src="https://github.com/user-attachments/assets/19eba1ed-90fa-4489-b96b-5c70a6149e6e" />

> If this number keeps increasing, it means GC overhead is increasing - possibly due to memory leaks or excessive allocations.

### HeapInUseBefore

Tells how much heap memory in use before that GC cycle. 

<img width="1409" height="833" alt="image" src="https://github.com/user-attachments/assets/bf082a55-e38f-4543-bdb7-5a8385c2c73a" />

### HeapMarkedLive

**Live Heap** =  Memory that is **actually reachable and in use** by your program after GC completes.

<img width="1404" height="840" alt="image" src="https://github.com/user-attachments/assets/986664a5-5a48-483d-afe2-b119ee216dc4" />

> If this value is keeping increasing and not degreasing (i.e next GC cycle is not cleaning it) that means we have memory leak

## Reference

- [Continuous Memory Plotting from GC Trace](https://www.cloudquery.io/blog/a-very-happy-golang-memory-profiling-story-at-cloudquery)

- [GC Trace Format](https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html)
