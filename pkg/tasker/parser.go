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
		log.Printf("[parser] can't open file: %s", opts.File)
		exit(1)
	}
	defer file.Close()

	lines := []string{}
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		ln := strings.TrimLeft(scan.Text(), " \t")
		// Skip empty or comment
		if ln != "" && ln[0] != '#' {
			lines = append(lines, ln)
		}
	}

	if err := scan.Err(); err != nil {
		if len(lines) == 0 {
			log.Printf("[parser] error reading taskfile: %v", err)
			exit(1)
		}

		log.Println(err)
	}

	return linesToTasks(lines)
}

// var cronRe = regexp.MustCompile(`^((?:[^\s]+\s+){5,6}(?:\d{4})?)(?:\s+)?(.*)`)
var aliasRe = regexp.MustCompile(`^(@(?:annually|yearly|monthly|weekly|daily|hourly|5minutes|10minutes|15minutes|30minutes|always|everysecond))(?:\s+)?(.*)`)
var segRe = regexp.MustCompile(`(?i),|/\d+$|^\d+-\d+$|^([0-7]|sun|mon|tue|wed|thu|fri|sat)(L|W|#\d)?$|-([0-7]|sun|mon|tue|wed|thu|fri|sat)$|\d{4}`)

func linesToTasks(lines []string) []Task {
	var tasks []Task

	gron := gronx.New()
	for _, line := range lines {
		var match []string
		if line[0] == '@' {
			match = aliasRe.FindStringSubmatch(line)
		} else {
			match = parseLine(line)
		}

		if len(match) > 2 && gron.IsValid(match[1]) {
			tasks = append(tasks, Task{strings.Trim(match[1], " \t"), match[2]})
			continue
		}

		log.Printf("[parser] can't parse cron expr: %s", line)
	}

	return tasks
}

func parseLine(line string) (match []string) {
	wasWs, expr, cmd := false, "", ""
	i, nseg, llen := 0, 0, len(line)-1
	match = append(match, line)

	for ; i < llen && nseg <= 7; i++ {
		isWs := strings.ContainsAny(line[i:i+1], "\t ")
		if nseg >= 5 {
			seg, ws := "", line[i-1:i]
			for i < llen && !strings.ContainsAny(line[i:i+1], "\t ") {
				i, seg = i+1, seg+line[i:i+1]
			}
			if isCronPart(seg) {
				expr, nseg = expr+ws+seg, nseg+1
			} else if seg != "" {
				cmd += seg
				break
			}
		} else {
			expr += line[i : i+1]
		}
		if isWs && !wasWs {
			nseg++
		}
		wasWs = isWs
	}
	cmd += line[i:]
	if nseg >= 5 && strings.TrimSpace(cmd) != "" {
		match = append(match, expr, cmd)
	}
	return
}

func isCronPart(seg string) bool {
	return seg != "" && seg[0] != '/' && (seg[0] == '*' || seg[0] == '?' || segRe.MatchString(seg))
}
