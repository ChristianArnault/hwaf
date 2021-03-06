package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func hwaf_make_cmd_self_update() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_self_update,
		UsageLine: "update [options]",
		Short:     "update hwaf itself",
		Long: `
update updates hwaf internal files and the hwaf binary itself.

ex:
 $ hwaf self update
`,
		Flag: *flag.NewFlagSet("hwaf-self-update", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")

	return cmd
}

func hwaf_run_cmd_self_update(cmd *commander.Command, args []string) {
	var err error
	n := "hwaf-self-" + cmd.Name()

	switch len(args) {
	case 0:
		// ok
	default:
		err = fmt.Errorf("%s: does NOT take any argument", n)
		handle_err(err)
	}

	verbose := cmd.Flag.Lookup("v").Value.Get().(bool)

	if verbose {
		fmt.Printf("%s...\n", n)
	}

	old, err := exec.LookPath(os.Args[0])
	handle_err(err)

	if os.Getenv("GOPATH") != "" {
		// use go get...
		pwd := ""
		pwd, err = os.Getwd()
		handle_err(err)
		gopaths := strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
		gopath := ""
		hwafpkg := filepath.Join("github.com", "hwaf", "hwaf")
		for _, v := range gopaths {
			if path_exists(filepath.Join(v, "src", hwafpkg)) {
				gopath = v
				break
			}
		}
		if gopath == "" {
			// hwaf package not installed...
			gopath = gopaths[0]
			gosrc := filepath.Join(gopath, "src")
			if !path_exists(gosrc) {
				err = os.MkdirAll(gosrc, 0700)
				handle_err(err)
			}
			err = os.Chdir(gosrc)
			handle_err(err)
			// first try r/w repository
			git := exec.Command(
				"git", "clone", "git@github.com:hwaf/hwaf",
				"github.com/hwaf/hwaf",
			)

			if verbose {
				git.Stdout = os.Stdout
				git.Stderr = os.Stderr
			}

			if git.Run() != nil {
				git := exec.Command(
					"git", "clone",
					"git://github.com/hwaf/hwaf",
					"github.com/hwaf/hwaf",
				)
				if verbose {
					git.Stdout = os.Stdout
					git.Stderr = os.Stderr
				}
				err = git.Run()
				handle_err(err)
			}
			err = os.Chdir(pwd)
			handle_err(err)
		}
		gosrc := filepath.Join(gopath, "src", hwafpkg)
		err = os.Chdir(gosrc)
		handle_err(err)

		// fetch...
		git := exec.Command("git", "fetch", "--all")
		if verbose {
			git.Stdout = os.Stdout
			git.Stderr = os.Stderr
		}
		err = git.Run()
		handle_err(err)

		// update...
		git = exec.Command("git", "checkout", "master")
		if verbose {
			git.Stdout = os.Stdout
			git.Stderr = os.Stderr
		}
		err = git.Run()
		handle_err(err)
		git = exec.Command("git", "pull", "origin", "master")
		if verbose {
			git.Stdout = os.Stdout
			git.Stderr = os.Stderr
		}
		err = git.Run()
		handle_err(err)

		// make sure we have all deps
		goget := exec.Command("go", "get", "-d", ".")
		if verbose {
			goget.Stdout = os.Stdout
			goget.Stderr = os.Stderr
		}
		err = goget.Run()
		handle_err(err)

		// rebuild
		goget = exec.Command("go", "build", ".")
		if verbose {
			goget.Stdout = os.Stdout
			goget.Stderr = os.Stderr
		}
		err = goget.Run()
		handle_err(err)

		// self init
		bin := filepath.Join(gosrc, "hwaf")
		hwaf := exec.Command(bin, "self", "init", fmt.Sprintf("-v=%v", verbose))
		hwaf.Stderr = os.Stderr
		hwaf.Stdout = os.Stdout
		err = hwaf.Run()
		handle_err(err)

		// replace current binary
		mv := exec.Command("/bin/mv", "-f", bin, old)
		mv.Stderr = os.Stderr
		mv.Stdout = os.Stdout
		err = mv.Run()
		handle_err(err)

		if verbose {
			fmt.Printf("%s... [ok]\n", n)
		}
		return
	}

	url := fmt.Sprintf(
		"http://cern.ch/hwaf/downloads/bin/hwaf-%s-%s",
		runtime.GOOS, runtime.GOARCH,
	)
	tmp, err := ioutil.TempFile("", "hwaf-self-update-")
	handle_err(err)
	defer tmp.Close()

	// make it executable
	err = tmp.Chmod(0777)
	handle_err(err)

	// download new file
	resp, err := http.Get(url)
	handle_err(err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("could not d/l [%s] (reason: %q)\n", url, resp.Status)
		handle_err(err)
	}

	nbytes, err := io.Copy(tmp, resp.Body)
	handle_err(err)

	if nbytes <= 0 {
		err = fmt.Errorf("could not copy hwaf from [%s]", url)
		handle_err(err)
	}

	err = tmp.Sync()
	handle_err(err)

	err = tmp.Close()
	handle_err(err)

	// self-init
	hwaf := exec.Command(
		tmp.Name(),
		"self", "init", fmt.Sprintf("-v=%v", verbose),
	)
	hwaf.Stderr = os.Stderr
	hwaf.Stdout = os.Stdout
	err = hwaf.Run()
	handle_err(err)

	// replace current binary
	mv := exec.Command("/bin/mv", "-f", tmp.Name(), old)
	mv.Stderr = os.Stderr
	mv.Stdout = os.Stdout
	err = mv.Run()
	handle_err(err)

	if verbose {
		fmt.Printf("%s: [%s] updated\n", n, old)
		fmt.Printf("%s... [ok]\n", n)
	}
}

// EOF
