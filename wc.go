package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
)

func main() {

	files, err := handleArgs()

	const sizeBuf = 100
	buf := make([]byte, sizeBuf)
	n := 0

	if err != nil {
		goto catchedError
	}

	for _, file := range files {
		for err == nil {
			n, err = file.Read(buf)
			for _, thing := range file.things {
				thing.Count(buf[:n])
			}
		}
		if err != io.EOF {
			goto catchedError
		}
		err = nil
	}
	for _, file := range files {
		fmt.Println(file.things, file)
	}

	printOverall(files)
	return

catchedError:
	fmt.Println(err)
	return
}

func handleArgs() (files []*File, err error) {
	if len(os.Args) < 2 {
		err = newArgError(InvalidNumberOfArguments)
		return
	}

	used := make(map[string]bool)
	for _, key := range keys {
		used[key] = false
	}

	for _, arg := range os.Args[1:] {
		switch {
		case arg == "--help":
			help()
			os.Exit(0)
		case arg == "--chars" || arg == "--bytes" || arg == "--words" || arg == "--lines" || arg == "--max-line-length":
			key := arg[2:]
			if !findThing(key) {
				goto catchedError
			}
			used[key] = true
		case arg[0] == '-':
			for i := 1; i < len(arg); i++ {
				key := arg[i : i+1]
				if !findThing(key) {
					goto catchedError
				}
				used[key] = true
			}
		default:
			var file *os.File
			file, err = os.Open(arg)
			if err != nil {
				goto catchedError
			}
			files = append(files, newFile(file, arg))
		}
	}
	if len(files) == 0 {
		files = append(files, newFile(os.Stdin, ""))
	}
	for _, file := range files {
		for _, key := range keys {
			if used[key] {
				file.things = append(file.things, newThing(key))
			}
		}
	}
	return

catchedError:
	if err == nil {
		err = newArgError(InvalidSyntaxOfArguments)
	}
	return
}

func help() {
	fmt.Print("Usage: ./wc [KEY]... [FILE]...                                     \n" +
		"If file wasn't mentioned, standard input is read.                        \n" +
		"Possible modes can set with keys:                                        \n" +
		"  -b, --bytes            print number of bytes                           \n" +
		"  -c, --chars            print number of characters                      \n" +
		"  -l, --lines            print number of new lines                       \n" +
		"  -m, --max-line-length  print lenght of the longest line                \n" +
		"  -w, --words            print number of words                           \n" +
		"      --help             show this manual and quite                      \n")
}

type ArgError struct {
	what string
}

const (
	InvalidNumberOfArguments = "invalid number of arguments"
	InvalidSyntaxOfArguments = "invalid syntax"
)

func newArgError(what string) *ArgError {
	return &ArgError{what}
}

func (e *ArgError) Error() (msg string) {
	msg += "You've entered invalid arguments! Please, fix it, and reload the program.\n"
	msg += "(" + e.what + ")\n"
	msg += "use '--help' for manual"
	return
}

func printOverall(files []*File) {
	if len(files) < 2 {
		return
	}
	var overall Things
	for _, thing := range files[0].things {
		overall = append(overall, newThing(thing.Key()))
	}
	values := make([]int, len(overall))
	for _, file := range files {
		for i, thing := range file.things {
			value, _ := strconv.Atoi(thing.Value())
			if overall[i].Key() == "m" {
				if values[i] < value {
					values[i] = value
				}
			} else {
				values[i] += value
			}
		}
	}
	for i := range overall {
		overall[i].SetValue(strconv.Itoa(values[i]))
	}
	fmt.Println(overall, "overall")
}

type File struct {
	file   *os.File
	path   string
	things Things
}

func newFile(file *os.File, path string) *File {
	return &File{file: file, path: path}
}

func (f *File) String() string {
	return f.path
}

func (f *File) Read(b []byte) (int, error) {
	return f.file.Read(b)
}

type Thing interface {
	Count([]byte)
	Key() string
	Value() string
	SetValue(string)
}

type Things []Thing

var keys = []string{
	"b",
	"c",
	"w",
	"l",
	"m",
	"bytes",
	"chars",
	"words",
	"lines",
	"max-line-length",
}

var newThings = map[string]NewThing{
	"b":               newBytes,
	"c":               newChars,
	"w":               newWords,
	"l":               newLines,
	"m":               newMaxLineLength,
	"bytes":           newBytes,
	"chars":           newChars,
	"words":           newWords,
	"lines":           newLines,
	"max-line-length": newMaxLineLength,
}

type NewThing func() Thing

func newThing(key string) Thing {
	return newThings[key]()
}

func findThing(key string) bool {
	return newThings[key] != nil
}

func (things Things) String() (s string) {
	for _, key := range keys {
		for _, t := range things {
			if t.Key() == key {
				s += t.Value() + "\t"
			}
		}
	}
	return
}

type Chars struct {
	count int
}

func newChars() Thing {
	return new(Chars)
}

func (c *Chars) Count(wasRead []byte) {
	for _ = range string(wasRead) {
		c.count++
	}
}

func (c *Chars) Key() string {
	return "c"
}

func (c *Chars) Value() string {
	return strconv.Itoa(c.count)
}

func (c *Chars) SetValue(value string) {
	c.count, _ = strconv.Atoi(value)
}

type Bytes struct {
	count int
}

func newBytes() Thing {
	return new(Bytes)
}

func (b *Bytes) Count(wasRead []byte) {
	b.count += len(wasRead)
}

func (b *Bytes) Key() string {
	return "b"
}

func (b *Bytes) Value() string {
	return strconv.Itoa(b.count)
}

func (b *Bytes) SetValue(value string) {
	b.count, _ = strconv.Atoi(value)
}

type Words struct {
	count int
	flag  bool
}

func newWords() Thing {
	return new(Words)
}

func (w *Words) Count(wasRead []byte) {
	for _, char := range string(wasRead) {
		isSpace := unicode.IsSpace(char)
		if !isSpace && !w.flag {
			w.count++
		}
		w.flag = !isSpace
	}
}

func (w *Words) Key() string {
	return "w"
}

func (w *Words) Value() string {
	return strconv.Itoa(w.count)
}

func (w *Words) SetValue(value string) {
	w.count, _ = strconv.Atoi(value)
}

type Lines struct {
	count int
}

func newLines() Thing {
	return new(Lines)
}

func (l *Lines) Count(wasRead []byte) {
	l.count += bytes.Count(wasRead, []byte{'\n'})
}

func (l *Lines) Key() string {
	return "l"
}

func (l *Lines) Value() string {
	return strconv.Itoa(l.count)
}

func (l *Lines) SetValue(value string) {
	l.count, _ = strconv.Atoi(value)
}

type MaxLineLength struct {
	cur int
	max int
}

func newMaxLineLength() Thing {
	return new(MaxLineLength)
}

func (m *MaxLineLength) Count(wasRead []byte) {
	for _, char := range string(wasRead) {
		if char == '\n' {
			if m.max < m.cur {
				m.max = m.cur
			}
			m.cur = 0
		} else {
			m.cur++
		}
	}
	if m.max < m.cur {
		m.max = m.cur
	}
}

func (m *MaxLineLength) Key() string {
	return "m"
}

func (m *MaxLineLength) Value() string {
	return strconv.Itoa(m.max)
}

func (m *MaxLineLength) SetValue(value string) {
	m.max, _ = strconv.Atoi(value)
}
