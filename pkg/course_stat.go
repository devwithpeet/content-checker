package pkg

import (
	"fmt"
	"strings"
)

type Color string

const (
	cliRed    Color = "\x1b[31m"
	cliPurple Color = "\x1b[35m"
	cliYellow Color = "\x1b[33m"
	cliGreen  Color = "\x1b[32m"
	cliBlue   Color = "\x1b[34m"
	cliReset  Color = "\x1b[0m"
	cliBold   Color = "\x1b[1m"
)

type CourseStat struct {
	Title      string
	Total      int
	Stub       int
	Incomplete int
	Complete   int
	Errors     int
}

func (cs *CourseStat) Print(columnWidths [6]int, columnColors [6]Color, total int) {
	complete := fmt.Sprintf("%d (%.0f%%)", cs.Complete, float64(cs.Complete)/float64(cs.Total)*100)
	incomplete := fmt.Sprintf("%d (%.0f%%)", cs.Incomplete, float64(cs.Incomplete)/float64(cs.Total)*100)
	stub := fmt.Sprintf("%d (%.0f%%)", cs.Stub, float64(cs.Stub)/float64(cs.Total)*100)
	totalData := fmt.Sprintf("%d (%.0f%%)", cs.Total, float64(cs.Total)/float64(total)*100)

	fmt.Printf(
		"%s | %s | %s | %s | %s | %s\n",
		column(cs.Title, columnWidths[0], columnColors[0]),
		column(totalData, columnWidths[1], columnColors[1]),
		column(complete, columnWidths[2], columnColors[2]),
		column(incomplete, columnWidths[3], columnColors[3]),
		column(stub, columnWidths[4], columnColors[4]),
		column(cs.Errors, columnWidths[5], columnColors[5]),
	)
}

func (cs *CourseStat) PrintHead(columnWidths [6]int, columnColors [6]Color) {
	fmt.Printf(
		"%s | %s | %s | %s | %s | %s\n",
		column("Course", columnWidths[0], columnColors[0]),
		column("All", columnWidths[1], columnColors[1]),
		column("Complete", columnWidths[2], columnColors[2]),
		column("Incomplete", columnWidths[3], columnColors[3]),
		column("Stub", columnWidths[4], columnColors[4]),
		column("Errors", columnWidths[5], columnColors[5]),
	)
}

func (cs *CourseStat) Line(columnWidths [6]int) {
	for i, width := range columnWidths {
		if i == 0 {
			fmt.Print(strings.Repeat("-", width+1))

			continue
		}

		fmt.Print("+", strings.Repeat("-", width+2))
	}

	fmt.Println()
}

func (cs *CourseStat) Add(stat CourseStat) {
	cs.Total += stat.Total
	cs.Complete += stat.Complete
	cs.Incomplete += stat.Incomplete
	cs.Stub += stat.Stub
	cs.Errors += stat.Errors
}

func NewCourseStat(title string, total, stub, incomplete, complete, errors int) CourseStat {
	return CourseStat{
		Title:      title,
		Total:      total,
		Complete:   complete,
		Incomplete: incomplete,
		Stub:       stub,
		Errors:     errors,
	}
}

func PrintStats(c Courses) {
	columnWidths := [6]int{17, 10, 10, 10, 10, 6}
	columnColors := [6]Color{cliBold, cliBold, cliGreen, cliYellow, cliPurple, cliRed}
	totalStat := NewCourseStat("Total", 0, 0, 0, 0, 0)

	totalStat.PrintHead(columnWidths, columnColors)
	totalStat.Line(columnWidths)

	stats := make([]CourseStat, 0, len(c))
	for _, course := range c {
		courseAll, courseStub, courseIncomplete, courseComplete, courseErrors := course.Stats()

		newStats := NewCourseStat(course.Course, courseAll, courseStub, courseIncomplete, courseComplete, courseErrors)

		stats = append(stats, newStats)

		totalStat.Add(newStats)
	}

	for _, stat := range stats {
		stat.Print(columnWidths, columnColors, totalStat.Total)
	}

	totalStat.Line(columnWidths)
	totalStat.Print(columnWidths, columnColors, totalStat.Total)
}
