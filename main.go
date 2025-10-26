package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

/*
	GC Trace Output Format
	Ref: https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html

	gc 2553 @8.452s 14%: 0.004+0.33+0.051 ms clock, 0.056+0.12/0.56/0.94+0.61 ms cpu, 4->4->2 MB, 5 MB goal, 12 P

	gc 2553     : The 2553 GC runs since the program started
	@8.452s     : Eight seconds since the program started
	14%         : Fourteen percent of the available CPU so far has been spent in GC

	// wall-clock
	0.004ms     : STW        : Write-Barrier - Wait for all Ps to reach a GC safe-point.
	0.33ms      : Concurrent : Marking
	0.051ms     : STW        : Mark Term     - Write Barrier off and clean up.

	// CPU time
	0.056ms     : STW        : Write-Barrier
	0.12ms      : Concurrent : Mark - Assist Time (GC performed in line with allocation)
	0.56ms      : Concurrent : Mark - Background GC time
	0.94ms      : Concurrent : Mark - Idle GC time
	0.61ms      : STW        : Mark Term

	4MB         : Heap memory in-use before the Marking started
	4MB         : Heap memory in-use after the Marking finished
	2MB         : Heap memory marked as live after the Marking finished
	5MB         : Collection goal for heap memory in-use after Marking finished

	// Threads
	12P         : Number of logical processors or threads used to run Goroutines.
**/

// GCTrace represents parsed GC trace data
type GCTrace struct {
	GCNum      int     // GC number
	Timestamp  float64 // Time since program start (seconds)
	CPUPercent float64 // CPU percentage spent in GC

	// Wall-clock times (ms)
	WallSTWWriteBarrier float64
	WallConcurrent      float64
	WallSTWMarkTerm     float64

	// CPU times (ms)
	CPUSTWWriteBarrier float64
	CPUMarkAssist      float64
	CPUMarkBackground  float64
	CPUMarkIdle        float64
	CPUSTWMarkTerm     float64

	// Memory (MB)
	HeapInUseBefore int
	HeapInUseAfter  int
	HeapMarkedLive  int
	HeapGoal        int
	StacksMB        int
	GlobalsMB       int

	// Processors
	NumProcs int
}

type ParserError int

const (
	LineNotMatch ParserError = iota
	RegexCompileFail
	NoError
)

func parseGCTrace(line string) (*GCTrace, ParserError) {
	// Pattern for: gc 2553 @8.452s 14%: 0.004+0.33+0.051 ms clock, 0.056+0.12/0.56/0.94+0.61 ms cpu, 4->4->2 MB, 5 MB goal, 12 P
	pattern := `gc\s+(\d+)\s+@([\d.]+)s\s+([\d.]+)%:\s+([\d.]+)\+([\d.]+)\+([\d.]+)\s+ms\s+clock,\s+([\d.]+)\+([\d.]+)/([\d.]+)/([\d.]+)\+([\d.]+)\s+ms\s+cpu,\s+(\d+)->(\d+)->(\d+)\s+MB,\s+(\d+)\s+MB\s+goal,\s+(\d+)\s+MB\s+stacks,\s+(\d+)\s+MB\s+globals,\s+(\d+)\s+P`

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, RegexCompileFail
	}
	matches := re.FindStringSubmatch(line)

	if matches == nil {
		return nil, LineNotMatch
	}

	// Helper to parse float
	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	// Helper to parse int
	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	trace := &GCTrace{
		GCNum:               parseInt(matches[1]),
		Timestamp:           parseFloat(matches[2]),
		CPUPercent:          parseFloat(matches[3]),
		WallSTWWriteBarrier: parseFloat(matches[4]),
		WallConcurrent:      parseFloat(matches[5]),
		WallSTWMarkTerm:     parseFloat(matches[6]),
		CPUSTWWriteBarrier:  parseFloat(matches[7]),
		CPUMarkAssist:       parseFloat(matches[8]),
		CPUMarkBackground:   parseFloat(matches[9]),
		CPUMarkIdle:         parseFloat(matches[10]),
		CPUSTWMarkTerm:      parseFloat(matches[11]),
		HeapInUseBefore:     parseInt(matches[12]),
		HeapInUseAfter:      parseInt(matches[13]),
		HeapMarkedLive:      parseInt(matches[14]),
		HeapGoal:            parseInt(matches[15]),
		StacksMB:            parseInt(matches[16]),
		GlobalsMB:           parseInt(matches[17]),
		NumProcs:            parseInt(matches[18]),
	}

	return trace, NoError
}

// ToCSVHeader returns CSV header row
func (t *GCTrace) ToCSVHeader() []string {
	return []string{
		"GCNum",
		"Timestamp",
		"CPUPercent",
		"WallSTWWriteBarrier",
		"WallConcurrent",
		"WallSTWMarkTerm",
		"CPUSTWWriteBarrier",
		"CPUMarkAssist",
		"CPUMarkBackground",
		"CPUMarkIdle",
		"CPUSTWMarkTerm",
		"HeapInUseBefore",
		"HeapInUseAfter",
		"HeapMarkedLive",
		"HeapGoal",
		"StacksMB",
		"GlobalsMB",
		"NumProcs",
	}
}

// ToCSVRow returns CSV data row
func (t *GCTrace) ToCSVRow() []string {
	return []string{
		strconv.Itoa(t.GCNum),
		strconv.FormatFloat(t.Timestamp, 'f', -1, 64),
		strconv.FormatFloat(t.CPUPercent, 'f', -1, 64),
		strconv.FormatFloat(t.WallSTWWriteBarrier, 'f', -1, 64),
		strconv.FormatFloat(t.WallConcurrent, 'f', -1, 64),
		strconv.FormatFloat(t.WallSTWMarkTerm, 'f', -1, 64),
		strconv.FormatFloat(t.CPUSTWWriteBarrier, 'f', -1, 64),
		strconv.FormatFloat(t.CPUMarkAssist, 'f', -1, 64),
		strconv.FormatFloat(t.CPUMarkBackground, 'f', -1, 64),
		strconv.FormatFloat(t.CPUMarkIdle, 'f', -1, 64),
		strconv.FormatFloat(t.CPUSTWMarkTerm, 'f', -1, 64),
		strconv.Itoa(t.HeapInUseBefore),
		strconv.Itoa(t.HeapInUseAfter),
		strconv.Itoa(t.HeapMarkedLive),
		strconv.Itoa(t.HeapGoal),
		strconv.Itoa(t.StacksMB),
		strconv.Itoa(t.GlobalsMB),
		strconv.Itoa(t.NumProcs),
	}
}

func main() {
	csvOutput := flag.String("csv", "data.csv", "GC Trace output in CSV format (e.g., data.csv). If empty, prints to stdout.")
	flag.Parse()

	csvFile, err := os.Create(*csvOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CSV file: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := csvFile.Close(); err != nil {
			fmt.Printf("Erorr: %s\n", err.Error())
		}
	}()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	scanner := bufio.NewScanner(os.Stdin)
	headerWritten := false

	for scanner.Scan() {
		line := scanner.Text()
		gcTrace, err := parseGCTrace(line)

		switch err {
		case LineNotMatch:
			fmt.Println(line)
			fmt.Println("Not Matching")
			continue
		case RegexCompileFail:
			fmt.Println("Error regex compile")
			os.Exit(1)
		}

		// Write CSV header on first successful parse
		if !headerWritten {
			if err := csvWriter.Write(gcTrace.ToCSVHeader()); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing CSV header: %v\n", err)
				os.Exit(1)
			}
			headerWritten = true
		}

		// Write CSV row
		if err := csvWriter.Write(gcTrace.ToCSVRow()); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing CSV row: %v\n", err)
			os.Exit(1)
		}

		// Flush immediately after each row
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing CSV row: %v\n", err)
			os.Exit(1)
		}

		// Sync to disk for maximum durability
		if err := csvFile.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing CSV file: %v\n", err)
			os.Exit(1)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
