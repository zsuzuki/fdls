package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type options struct {
	root     string
	pathMode string
	showHash bool
	showDate bool
	showSize bool
	maxDepth int
}

type entry struct {
	path    string
	size    int64
	modTime string
	hash    string
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	var opts options

	fs := flag.NewFlagSet("fdls", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.pathMode, "path", "rel", "path output mode: rel or abs")
	fs.BoolVar(&opts.showHash, "sha256", false, "show SHA256 hash")
	fs.BoolVar(&opts.showDate, "date", false, "show modified date")
	fs.BoolVar(&opts.showSize, "size", false, "show file size in bytes")
	fs.IntVar(&opts.maxDepth, "depth", -1, "search depth: -1 is infinite, 0 is current directory only")

	if err := fs.Parse(args); err != nil {
		return usageError(err)
	}
	if fs.NArg() != 1 {
		return usageError(errors.New("directory must be specified"))
	}

	opts.root = fs.Arg(0)
	return listFiles(opts, out)
}

func usageError(err error) error {
	return fmt.Errorf("%w\nusage: fdls [options] <directory>\n\noptions:\n  -path rel|abs\n  -sha256\n  -date\n  -size\n  -depth N (-1=infinite, 0=no recursion)", err)
}

func listFiles(opts options, out io.Writer) error {
	if opts.pathMode != "rel" && opts.pathMode != "abs" {
		return fmt.Errorf("invalid -path value %q: use rel or abs", opts.pathMode)
	}
	if opts.maxDepth < -1 {
		return fmt.Errorf("invalid -depth value %d: use -1 or greater", opts.maxDepth)
	}

	rootAbs, err := filepath.Abs(opts.root)
	if err != nil {
		return err
	}

	info, err := os.Stat(rootAbs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", opts.root)
	}

	entries := make([]entry, 0)
	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == rootAbs {
			return nil
		}

		depth, err := depthFromRoot(rootAbs, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if opts.maxDepth >= 0 && depth > opts.maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if opts.maxDepth >= 0 && depth > opts.maxDepth {
			return nil
		}

		e, err := buildEntry(opts, rootAbs, path, d)
		if err != nil {
			return err
		}
		entries = append(entries, e)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].path < entries[j].path
	})

	for _, e := range entries {
		cols := []string{e.path}
		if opts.showSize {
			cols = append(cols, fmt.Sprintf("%d", e.size))
		}
		if opts.showDate {
			cols = append(cols, e.modTime)
		}
		if opts.showHash {
			cols = append(cols, e.hash)
		}
		fmt.Fprintln(out, strings.Join(cols, "\t"))
	}

	return nil
}

func depthFromRoot(rootAbs, path string) (int, error) {
	rel, err := filepath.Rel(rootAbs, path)
	if err != nil {
		return 0, err
	}
	return len(strings.Split(rel, string(os.PathSeparator))) - 1, nil
}

func buildEntry(opts options, rootAbs, path string, d fs.DirEntry) (entry, error) {
	info, err := d.Info()
	if err != nil {
		return entry{}, err
	}

	displayPath := path
	if opts.pathMode == "rel" {
		displayPath, err = filepath.Rel(rootAbs, path)
		if err != nil {
			return entry{}, err
		}
	}

	e := entry{
		path:    displayPath,
		size:    info.Size(),
		modTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
	}

	if opts.showHash {
		hash, err := sha256File(path)
		if err != nil {
			return entry{}, err
		}
		e.hash = hash
	}

	return e, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
