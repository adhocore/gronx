package tasker

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/adhocore/gronx"
)

// MustParseTaskfile either parses taskfile from given Option.
// It fails hard in case any error.
func MustParseTaskfile(opts Option) []Task {
	file, err := os.Open(opts.File)
	if err != nil {
		log.Fatalf("[parser] can't open file: %s", opts.File)
	}
	defer file.Close()

	lines := []string{}
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		ln := strings.TrimLeft(scan.Text(), " \t")
		if ln != "" && ln[0] != '#' {
			lines = append(lines, ln)
		}
	}

	if err := scan.Err(); err != nil {
		if len(lines) == 0 {
			log.Fatalf("[parser] error reading taskfile: %v", err)
		}

		log.Println(err)
	}

	return linesToTasks(lines)
}

var cronRe = regexp.MustCompile(`^((?:[^\s]+\s+){5}(?:\d{4})?)(?:\s+)?(.*)`)
var aliasRe = regexp.MustCompile(`^(@(?:annually|yearly|monthly|weekly|daily|hourly|5minutes|10minutes|15minutes|30minutes|always))(?:\s+)?(.*)`)

func linesToTasks(lines []string) []Task {
	var tasks []Task

	gron := gronx.New()
	for _, line := range lines {
		var match []string
		if line[0] == '@' {
			match = aliasRe.FindStringSubmatch(line)
		} else {
			match = cronRe.FindStringSubmatch(line)
		}
		if len(match) > 2 && gron.IsValid(match[1]) {
			tasks = append(tasks, Task{strings.Trim(match[1], " \t"), match[2]})
			continue
		}

		log.Printf("[parser] can't parse cron expr: %s", line)
	}

	return tasks
}
