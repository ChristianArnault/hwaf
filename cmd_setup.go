package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbinet/go-commander"
	"github.com/sbinet/go-flag"
	gocfg "github.com/sbinet/go-config/config"
)

func hwaf_make_cmd_setup() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_setup,
		UsageLine: "setup [options] <workarea>",
		Short:     "setup an existing workarea",
		Long: `
setup sets up an existing workarea.

ex:
 $ hwaf setup
 $ hwaf setup .
 $ hwaf setup my-work-area
 $ hwaf setup -projects=/opt/sw/mana/mana-core/20121207 my-work-area
 $ hwaf setup -cfg=${HWAF_CFG}/usr.cfg my-work-area
`,
		Flag: *flag.NewFlagSet("hwaf-setup", flag.ExitOnError),
	}
	cmd.Flag.String("projects", "/opt/sw/mana", "List of paths to projects to setup against")
	cmd.Flag.String("cfg", "", "Path to a configuration file")
	cmd.Flag.Bool("q", false, "only print error and warning messages, all other output will be suppressed")

	return cmd
}

func hwaf_run_cmd_setup(cmd *commander.Command, args []string) {
	var err error
	n := "hwaf-" + cmd.Name()
	dirname := "."
	switch len(args) {
	case 0:
		dirname = "."
	case 1:
		dirname = args[0]
	default:
		err = fmt.Errorf("%s: you need to give a directory name", n)
		handle_err(err)
	}

	dirname = os.ExpandEnv(dirname)
	dirname = filepath.Clean(dirname)

	quiet := cmd.Flag.Lookup("q").Value.Get().(bool)
	cfg_fname:= cmd.Flag.Lookup("cfg").Value.Get().(string)

	projdirs := []string{}
	const pathsep = string(os.PathListSeparator)
	for _, v := range strings.Split(cmd.Flag.Lookup("projects").Value.Get().(string), pathsep) {
		if v != "" {
			v = os.ExpandEnv(v)
			v = filepath.Clean(v)
			projdirs = append(projdirs, v)
		}
	}

	if !quiet {
		fmt.Printf("%s: setup workarea [%s]...\n", n, dirname)
		fmt.Printf("%s: projects=%v\n", n, projdirs)
		if cfg_fname != "" {
			fmt.Printf("%s: cfg-file=%s\n", n, cfg_fname)
		}			
	}

	for _, projdir := range projdirs {
		if !path_exists(projdir) {
			err = fmt.Errorf("no such directory: [%s]", projdir)
			handle_err(err)
		}

		pinfo := filepath.Join(projdir, "project.info")
		if !path_exists(pinfo) {
			err = fmt.Errorf("no such file: [%s]", pinfo)
			handle_err(err)
		}
	}

	pwd, err := os.Getwd()
	handle_err(err)
	defer os.Chdir(pwd)

	err = os.Chdir(dirname)
	handle_err(err)

	if !quiet {
		fmt.Printf("%s: create local config...\n", n)
	}

	lcfg_fname := filepath.Join(dirname, ".hwaf", "local.cfg")
	if path_exists(lcfg_fname) {
		err = os.Remove(lcfg_fname)
		handle_err(err)
	}

	lcfg := gocfg.NewDefault()
	section := "hwaf-cfg"
	if !lcfg.AddSection(section) {
		err = fmt.Errorf("%s: could not create section [%s] in file [%s]", 
			n, section, lcfg_fname)
		handle_err(err)
	}
	
	for k, v := range map[string]string{
		"projects": strings.Join(projdirs, pathsep),
		"cmtpkgs": "pkg",
	} {
		if !lcfg.AddOption(section, k, v) {
			err := fmt.Errorf("%s: could not add option [%s] to section [%s]", 
				n, k, section,
				)
			handle_err(err)
		}
	}

	err = lcfg.WriteFile(lcfg_fname, 0600, "")
	handle_err(err)
	
	if !quiet {
		fmt.Printf("%s: setup workarea [%s]... [ok]\n", n, dirname)
	}
}

// EOF