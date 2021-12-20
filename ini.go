// Package ini provides functions for parsing INI configuration files.
package ini

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	sectionRegex  = regexp.MustCompile(`^\[(.*)\]$`)
	assignRegex   = regexp.MustCompile(`^([^=]+)=(.*)$`)
	descRegex     = regexp.MustCompile(`(?m)(?i)^\[(description)\]$`)
	TimerSections = make(TimeMap)
)

// ErrSyntax is returned when there is a syntax error in an INI file.
type ErrSyntax struct {
	Line   int
	Source string // The contents of the erroneous line, without leading or trailing whitespace
}

func (e ErrSyntax) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

// A File represents a parsed INI file.
type File map[string]Section

// A Section represents a single section of an INI file.
type Section map[string]string

// Returns a named Section. A Section will be created if one does not already exist for the given name.
func (f File) Section(name string) Section {
	section := f[name]
	if section == nil {
		section = make(Section)
		f[name] = section
	}
	return section
}

// 根据名称返回Section，如果找不到则返回nil
func (f File) GetSection(name string) Section {
	section := f[name]
	return section
}

type TimeMap map[int]string

// 专用函数，用于统计section名称为纯数字的段落数量
func (f File) TimeSectionCount() int {
	TimerSections = make(TimeMap)
	for k, _ := range f {
		i, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		TimerSections[i] = k
	}
	return len(TimerSections)
}

// Looks up a value for a key in a section and returns that value, along with a boolean result similar to a map lookup.
func (f File) Get(section, key string) (value string, ok bool) {
	if s := f[section]; s != nil {
		value, ok = s[key]
	}
	return
}

// Loads INI data from a reader and stores the data in the File.
func (f File) Load(in io.Reader) (err error) {
	bufin, ok := in.(*bufio.Reader)
	if !ok {
		bufin = bufio.NewReader(in)
	}
	return parseFile(bufin, f)
}

// Loads INI data from a named file and stores the data in the File.
func (f File) LoadFile(file string) (err error) {
	in, err := os.Open(file)
	if err != nil {
		return
	}
	defer in.Close()
	return f.Load(in)
}

func parseFile(in *bufio.Reader, file File) (err error) {
	section := ""
	lineNum := 0
	for done := false; !done; {
		var line string
		if line, err = in.ReadString('\n'); err != nil {
			if err == io.EOF {
				done = true
			} else {
				return
			}
		}
		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}

		if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)
			file.Section(section)[key] = val
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
			// Create the section if it does not exist
			file.Section(section)
		} else {
			return ErrSyntax{lineNum, line}
		}

	}
	return nil
}

// Loads and returns a File from a reader.
func Load(in io.Reader) (File, error) {
	file := make(File)
	err := file.Load(in)
	return file, err
}

// Loads and returns an INI File from a file on disk.
func LoadFile(filename string) (File, error) {
	file := make(File)
	err := file.LoadFile(filename)
	return file, err
}

// 专用函数，读取模型描述的信息
func LoadModDesc(file string) (rst map[string]string, err error) {
	rst = make(map[string]string)
	in, err := os.Open(file)
	if err != nil {
		return
	}
	defer in.Close()
	bufin := bufio.NewReader(in)
	err = parseFileDesc(bufin, rst)
	return
}

// 专用函数。只读描述section
func parseFileDesc(in *bufio.Reader, descmap map[string]string) (err error) {
	found := false
	lineNum := 0
	for done := false; !done; {
		var line string
		if line, err = in.ReadString('\n'); err != nil {
			if err == io.EOF {
				done = true
			} else {
				return
			}
		}
		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}

		if !found {
			if len(descRegex.FindStringIndex(line)) > 0 {
				// 找到desc section
				found = true
			}
			continue
		}

		if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)
			descmap[key] = val
		} else {
			// 下一小节，结束
			break
		}

	}
	return nil
}
