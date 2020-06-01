package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Contants used for the meta tags
const (
	Version     = "1.4.1"
	Name        = "gotags"
	URL         = "https://github.com/jstemmer/gotags"
	AuthorName  = "Joel Stemmer"
	AuthorEmail = "stemmertech@gmail.com"
)

var (
	printVersion bool
	inputFile    string
	outputFile   string
	recurse      bool
	sortOutput   bool
	silent       bool
	relative     bool
	listLangs    bool
	fields       string
	extraSymbols string
)

// ContinueOnError 忽视解析错误
var flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

// Initialize flags.
// init函数： 初始化参数
func init() {
	flags.BoolVar(&printVersion, "v", false, "print version.")
	flags.StringVar(&inputFile, "L", "", `source file names are read from the specified file. If file is "-", input is read from standard in.`)
	flags.StringVar(&outputFile, "f", "", `write output to specified file. If file is "-", output is written to standard out.`)
	flags.BoolVar(&recurse, "R", false, "recurse into directories in the file list.")
	flags.BoolVar(&sortOutput, "sort", true, "sort tags.")
	flags.BoolVar(&silent, "silent", false, "do not produce any output on error.")
	flags.BoolVar(&relative, "tag-relative", false, "file paths should be relative to the directory containing the tag file.")
	flags.BoolVar(&listLangs, "list-languages", false, "list supported languages.")
	flags.StringVar(&fields, "fields", "", "include selected extension fields (only +l).")
	flags.StringVar(&extraSymbols, "extra", "", "include additional tags with package and receiver name prefixes (+q)")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "gotags version %s\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: %s [options] file(s)\n\n", os.Args[0])
		flags.PrintDefaults()
	}
}

// walkDir函数：遍历所有*.go文件
func walkDir(names []string, dir string) ([]string, error) {
	e := filepath.Walk(dir, func(path string, finfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") && !finfo.IsDir() {
			names = append(names, path)
		}
		return nil
	})

	return names, e
}

// recurseNames函数： 对指定的多个目录或文件进行确认
func recurseNames(names []string) ([]string, error) {
	var ret []string
	for _, name := range names {
		info, e := os.Stat(name)
		if e != nil || info == nil || !info.IsDir() {
			ret = append(ret, name) // defer the error handling to the scanner
		} else {
			ret, e = walkDir(ret, name)
			if e != nil {
				return names, e
			}
		}
	}
	return ret, nil
}

// readNames函数：文件内容读入
func readNames(names []string) ([]string, error) {
	// 没有输入文件
	if len(inputFile) == 0 {
		return names, nil
	}

	var scanner *bufio.Scanner
	if inputFile != "-" {
		in, err := os.Open(inputFile)
		if err != nil {
			return nil, err
		}

		defer in.Close()
		scanner = bufio.NewScanner(in)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for scanner.Scan() {
		names = append(names, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

// getFileNames函数：根据recurse来确定是否使用递归遍历符合要求的go文件
func getFileNames() ([]string, error) {
	var names []string

	names = append(names, flags.Args()...)
	names, err := readNames(names)
	if err != nil {
		return nil, err
	}

	if recurse {
		names, err = recurseNames(names)
		if err != nil {
			return nil, err
		}
	}

	return names, nil
}

// 主函数
func main() {
	if err := flags.Parse(os.Args[1:]); err == flag.ErrHelp {
		return
	}

	if printVersion {
		fmt.Printf("gotags version %s\n", Version)
		return
	}

	if listLangs {
		fmt.Println("Go")
		return
	}

	files, err := getFileNames()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get specified files\n\n")
		flags.Usage()
		os.Exit(1)
	}

	if len(files) == 0 && len(inputFile) == 0 {
		fmt.Fprintf(os.Stderr, "no file specified\n\n")
		flags.Usage()
		os.Exit(1)
	}

	var basedir string
	if relative {
		basedir, err = filepath.Abs(filepath.Dir(outputFile))
		if err != nil {
			if !silent {
				fmt.Fprintf(os.Stderr, "could not determine absolute path: %s\n", err)
			}
			os.Exit(1)
		}
	}

	// 解析fields，支持的语言，对应全局变量fields
	fieldSet, err := parseFields(fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n\n", err)
		flags.Usage()
		os.Exit(1)
	}

	// 解析额外的Symbols，对应全局变量 extraSymbols
	symbolSet, err := parseExtraSymbols(extraSymbols)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n\n", err)
		flags.Usage()
		os.Exit(1)
	}

	// 解析Tags
	tags := []Tag{}
	for _, file := range files {
		ts, err := Parse(file, relative, basedir, symbolSet)
		if err != nil {
			if !silent {
				fmt.Fprintf(os.Stderr, "parse error: %s\n\n", err)
			}
			continue
		}
		tags = append(tags, ts...)
	}

	// 输出部分
	output := createMetaTags()
	for _, tag := range tags {
		if fieldSet.Includes(Language) {
			tag.Fields[Language] = "Go"
		}
		output = append(output, tag.String())
	}

	if sortOutput {
		sort.Sort(sort.StringSlice(output))
	}

	var out io.Writer
	if len(outputFile) == 0 || outputFile == "-" {
		// For compatibility with older gotags versions, also write to stdout
		// when outputFile is not specified.
		out = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create output file: %s\n", err)
			os.Exit(1)
		}
		out = file
		defer file.Close()
	}

	for _, s := range output {
		fmt.Fprintln(out, s)
	}
}

// createMetaTags函数： 生成源数据
func createMetaTags() []string {
	var sorted int
	if sortOutput {
		sorted = 1
	}
	return []string{
		"!_TAG_FILE_FORMAT\t2",
		fmt.Sprintf("!_TAG_FILE_SORTED\t%d\t/0=unsorted, 1=sorted/", sorted),
		fmt.Sprintf("!_TAG_PROGRAM_AUTHOR\t%s\t/%s/", AuthorName, AuthorEmail),
		fmt.Sprintf("!_TAG_PROGRAM_NAME\t%s", Name),
		fmt.Sprintf("!_TAG_PROGRAM_URL\t%s", URL),
		fmt.Sprintf("!_TAG_PROGRAM_VERSION\t%s\t/%s/", Version, runtime.Version()),
	}
}
