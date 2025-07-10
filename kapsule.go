package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	_ "github.com/PlakarKorp/kapsul/connectors/ptar"
	_ "github.com/PlakarKorp/kapsul/connectors/sftp"
	_ "github.com/PlakarKorp/kapsul/connectors/stdio"

	fs "github.com/PlakarKorp/integration-fs"
	"github.com/PlakarKorp/kloset/caching"
	"github.com/PlakarKorp/kloset/location"
	"github.com/PlakarKorp/kloset/logging"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/kloset/snapshot/exporter"
	"github.com/PlakarKorp/kloset/snapshot/importer"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/cookies"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/subcommands/archive"
	"github.com/PlakarKorp/plakar/subcommands/cat"
	"github.com/PlakarKorp/plakar/subcommands/check"
	"github.com/PlakarKorp/plakar/subcommands/diff"
	"github.com/PlakarKorp/plakar/subcommands/digest"
	"github.com/PlakarKorp/plakar/subcommands/help"
	"github.com/PlakarKorp/plakar/subcommands/locate"
	"github.com/PlakarKorp/plakar/subcommands/ls"
	"github.com/PlakarKorp/plakar/subcommands/ptar"
	"github.com/PlakarKorp/plakar/subcommands/restore"
	"github.com/PlakarKorp/plakar/subcommands/server"
	"github.com/PlakarKorp/plakar/subcommands/ui"
)

func init() {
	importer.Register("fs", location.FLAG_LOCALFS, fs.NewFSImporter)
	exporter.Register("fs", location.FLAG_LOCALFS, fs.NewFSExporter)
}

func main() {
	var kapsulPath string
	var ncores int
	flag.StringVar(&kapsulPath, "f", "", "Path to the kapsul")
	flag.IntVar(&ncores, "c", 0, "Number of cores to use (default: all available cores -1)")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	numcpu := runtime.NumCPU()
	if ncores < 0 || ncores > numcpu {
		fmt.Fprintf(os.Stderr, "Invalid number of cores: %d. Must be between 0 and %d.\n", ncores, numcpu)
		return
	}
	if ncores == 0 {
		ncores = numcpu - 1
		if ncores < 1 {
			ncores = 1
		}
	}
	runtime.GOMAXPROCS(ncores)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		return
	}

	// how do I create a temporary directory?
	tmp, err := os.MkdirTemp("", "kapsul")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tmp)

	ctx := appcontext.NewAppContext()
	ctx.CWD = cwd
	ctx.MaxConcurrency = 42
	ctx.SetCookies(cookies.NewManager(tmp))

	ctx.SetLogger(logging.NewLogger(os.Stdout, os.Stderr))
	ctx.SetCache(caching.NewManager(tmp))

	if flag.Arg(0) == "create" {
		repo, err := repository.Inexistent(ctx.GetInner(), map[string]string{
			"location": kapsulPath,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating kapsul: %v\n", err)
			return
		}

		subc := &ptar.Ptar{}
		args := append([]string{"-o", kapsulPath}, flag.Args()[1:]...)
		if err := subc.Parse(ctx, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing ptar command: %v\n", err)
			os.Exit(1)
		} else if _, err := subc.Execute(ctx, repo); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing ptar command: %v\n", err)
			os.Exit(1)
		}
		return
	}

	repo, err := openKapsule(ctx, kapsulPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening kapsul: %v\n", err)
		return
	}
	defer repo.Close()

	var subcommands = map[string]subcommands.Subcommand{
		"archive": &archive.Archive{},
		"cat":     &cat.Cat{},
		"check":   &check.Check{},
		// clone
		"diff":   &diff.Diff{},
		"digest": &digest.Digest{},
		"help":   &help.Help{},
		// info
		"locate": &locate.Locate{},
		"ls":     &ls.Ls{},
		// mount
		"restore": &restore.Restore{},
		"server":  &server.Server{},
		// sync
		"ui": &ui.Ui{},
	}
	if subc, ok := subcommands[flag.Arg(0)]; !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", flag.Arg(0))
		flag.Usage()
		os.Exit(1)
	} else {
		if err := subc.Parse(ctx, flag.Args()[1:]); err != nil {
			os.Exit(1)
		} else if _, err := subc.Execute(ctx, repo); err != nil {
			os.Exit(1)
		}
	}
}
