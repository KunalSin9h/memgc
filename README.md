# memgc

Simple continuous memory info aggregator from **GoLang** `gc trace`

Installation:

```bash
go install github.com/kunalsin9h/memgc@latest
```

Usage:

```bash
GODEBUG=gctrace=1 ./bin 2> >(memgc --csv data.csv)
```

This will output a `data.csv` file.

## Reference

- [Continuous Memory Plotting from GC Trace](https://www.cloudquery.io/blog/a-very-happy-golang-memory-profiling-story-at-cloudquery)
- [GC Trace Format](https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html)
