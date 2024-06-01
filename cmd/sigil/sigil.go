package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/UNO-SOFT/sigil"
	_ "github.com/UNO-SOFT/sigil/builtin"
)

var Version string

var (
	filename   = flag.String("f", "", "use template file instead of STDIN")
	inline     = flag.String("i", "", "use inline template string instead of STDIN")
	posix      = flag.Bool("p", false, "preprocess with POSIX variable expansion")
	version    = flag.Bool("v", false, "prints version")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to this file")
	verbose    = flag.Bool("verbose", false, "verbose logging")
)

func template() ([]byte, string, error) {
	if *filename != "" {
		data, err := os.ReadFile(*filename)
		if err != nil {
			return []byte{}, "", err
		}
		sigil.PushPath(filepath.Dir(*filename))
		return data, filepath.Base(*filename), nil
	}
	if *inline != "" {
		return []byte(*inline), "<inline>", nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return []byte{}, "", err
	}
	return data, "<stdin>", nil
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}
	if *verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *posix {
		sigil.PosixPreprocess = true
	}
	if os.Getenv("SIGIL_PATH") != "" {
		sigil.TemplatePath = strings.Split(os.Getenv("SIGIL_PATH"), ":")
	}
	vars := make(map[string]interface{})
	for _, arg := range flag.Args() {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}
	tmpl, name, err := template()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	render, err := sigil.Execute(tmpl, vars, name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(render.Bytes())
}
