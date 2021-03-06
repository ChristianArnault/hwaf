package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/hwaf/hwaf/hwaflib"
)

func hwaf_make_cmd_waf_bdist_rpm() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_waf_bdist_rpm,
		UsageLine: "bdist-rpm [rpm-name]",
		Short:     "create a RPM from the local project/packages",
		Long: `
bdist-rpm creates a RPM from the local project/packages.

ex:
 $ hwaf bdist-rpm
 $ hwaf bdist-rpm -name=mana
 $ hwaf bdist-rpm -name=mana -version=20130101
`,
		Flag: *flag.NewFlagSet("hwaf-bdist-rpm", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	cmd.Flag.String("name", "", "name of the binary distribution (default: project name)")
	cmd.Flag.String("version", "", "version of the binary distribution (default: project version)")
	cmd.Flag.String("release", "1", "release version of the binary distribution (default: 1)")
	cmd.Flag.String("cmtcfg", "", "CMTCFG quadruplet for the binary distribution (default: project CMTCFG)")
	cmd.Flag.String("spec", "", "RPM SPEC file for the binary distribution")
	cmd.Flag.String("url", "", "URL for the RPM binary distribution")
	return cmd
}

func hwaf_run_cmd_waf_bdist_rpm(cmd *commander.Command, args []string) {
	var err error
	n := "hwaf-" + cmd.Name()

	switch len(args) {
	case 0:
	default:
		err = fmt.Errorf("%s: too many arguments (%s)", n, len(args))
		handle_err(err)
	}

	verbose := cmd.Flag.Lookup("v").Value.Get().(bool)

	bdist_name := cmd.Flag.Lookup("name").Value.Get().(string)
	bdist_vers := cmd.Flag.Lookup("version").Value.Get().(string)
	bdist_release := cmd.Flag.Lookup("release").Value.Get().(string)
	bdist_cmtcfg := cmd.Flag.Lookup("cmtcfg").Value.Get().(string)
	bdist_spec := cmd.Flag.Lookup("spec").Value.Get().(string)

	bdist_url := cmd.Flag.Lookup("url").Value.Get().(string)
	if bdist_url == "" {
		bdist_url = "http://cern.ch/mana-fwk"
	}

	type RpmInfo struct {
		Name      string // RPM package name
		Vers      string // RPM package version
		Release   string // RPM package release
		CmtCfg    string // RPM CMTCFG quadruplet
		BuildRoot string // RPM build directory
		Url       string // URL home page
	}

	workdir, err := g_ctx.Workarea()
	if err != nil {
		// not a git repo... assume we are at the root, then...
		workdir, err = os.Getwd()
	}
	handle_err(err)

	if bdist_name == "" {
		bdist_name = workdir
		bdist_name = filepath.Base(bdist_name)
	}
	if bdist_vers == "" {
		bdist_vers = time.Now().Format("20060102")
	}
	if bdist_cmtcfg == "" {
		// FIXME: get actual value from waf, somehow
		pinfo_name := filepath.Join(workdir, "__build__", "c4che", "_cache.py")
		if !path_exists(pinfo_name) {
			err = fmt.Errorf(
				"no such file [%s]. did you run \"hwaf configure\" ?",
				pinfo_name,
			)
			handle_err(err)
		}
		pinfo, err := hwaflib.NewProjectInfo(pinfo_name)
		handle_err(err)
		bdist_cmtcfg, err = pinfo.Get("CMTCFG")
		handle_err(err)
	}
	fname := bdist_name + "-" + bdist_vers + "-" + bdist_cmtcfg
	rpmbldroot, err := ioutil.TempDir("", "hwaf-rpm-buildroot-")
	handle_err(err)
	defer os.RemoveAll(rpmbldroot)
	for _, dir := range []string{
		"RPMS", "SRPMS", "BUILD", "SOURCES", "SPECS", "tmp",
	} {
		err = os.MkdirAll(filepath.Join(rpmbldroot, dir), 0700)
		handle_err(err)
	}

	specfile, err := os.Create(filepath.Join(rpmbldroot, "SPECS", bdist_name+".spec"))
	handle_err(err)

	rpminfos := RpmInfo{
		Name:      bdist_name,
		Vers:      bdist_vers,
		Release:   bdist_release,
		CmtCfg:    bdist_cmtcfg,
		BuildRoot: rpmbldroot,
		Url:       bdist_url,
	}

	// get tarball from 'hwaf bdist'...
	bdist_fname := strings.Replace(fname, ".rpm", "", 1) + ".tar.gz"
	if !path_exists(bdist_fname) {
		err = fmt.Errorf("no such file [%s]. did you run \"hwaf bdist\" ?", bdist_fname)
		handle_err(err)
	}
	bdist_fname, err = filepath.Abs(bdist_fname)
	handle_err(err)
	{
		// first, massage the tar ball to something rpmbuild expects...

		// ok, now we're done.
		dst, err := os.Create(filepath.Join(rpmbldroot, "SOURCES", filepath.Base(bdist_fname)))
		handle_err(err)
		src, err := os.Open(bdist_fname)
		handle_err(err)
		_, err = io.Copy(dst, src)
		handle_err(err)
	}

	if bdist_spec != "" {
		bdist_spec = os.ExpandEnv(bdist_spec)
		bdist_spec, err = filepath.Abs(bdist_spec)
		handle_err(err)

		if !path_exists(bdist_spec) {
			err = fmt.Errorf("no such file [%s]", bdist_spec)
			handle_err(err)
		}
		user_spec, err := os.Open(bdist_spec)
		handle_err(err)
		defer user_spec.Close()

		_, err = io.Copy(specfile, user_spec)
		handle_err(err)
	} else {
		bdist_spec = specfile.Name()

		var spec_tmpl *template.Template
		spec_tmpl, err = template.New("SPEC").Parse(`
%define __spec_install_post %{nil}
%define   debug_package %{nil}
%define __os_install_post %{_dbpath}/brp-compress
%define   cmtcfg {{.CmtCfg}}
%define _topdir {{.BuildRoot}}
%define _tmppath  %{_topdir}/tmp

Summary: hwaf generated RPM for {{.Name}}
Name: {{.Name}}
Version: {{.Vers}}
Release: {{.Release}}
License: Unknown
Group: Development/Tools
SOURCE0 : %{name}-%{version}-%{cmtcfg}.tar.gz
URL: {{.Url}}

BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root

%description
%{summary}

%prep
%setup -q

%build
# Empty section.

%install
rm -rf %{buildroot}
mkdir -p  %{buildroot}

# in builddir
cp -a * %{buildroot}


%clean
rm -rf %{buildroot}


%files
%defattr(-,root,root,-)
/*
`) // */ for emacs...
		handle_err(err)

		err = spec_tmpl.Execute(specfile, rpminfos)
		handle_err(err)
	}

	err = specfile.Sync()
	handle_err(err)
	err = specfile.Close()
	handle_err(err)

	if !strings.HasSuffix(fname, ".rpm") {
		fname = fname + ".rpm"
	}

	if verbose {
		fmt.Printf("%s: building RPM [%s]...\n", n, fname)
	}

	rpmbld, err := exec.LookPath("rpmbuild")
	handle_err(err)

	rpm := exec.Command(rpmbld,
		"-bb",
		filepath.Join("SPECS", rpminfos.Name+".spec"),
	)
	rpm.Dir = rpmbldroot
	if verbose {
		rpm.Stdin = os.Stdin
		rpm.Stdout = os.Stdout
		rpm.Stderr = os.Stderr
	}
	err = rpm.Run()
	handle_err(err)

	dst, err := os.Create(fname)
	handle_err(err)
	defer dst.Close()

	rpmarch := ""
	switch runtime.GOARCH {
	case "amd64":
		rpmarch = "x86_64"
	case "386":
		rpmarch = "i386"
	default:
		err = fmt.Errorf("unhandled GOARCH [%s]", runtime.GOARCH)
		handle_err(err)
	}
	srcname := fmt.Sprintf(
		"%s-%s-%s.%s.rpm",
		rpminfos.Name,
		rpminfos.Vers,
		rpminfos.Release,
		rpmarch)

	src, err := os.Open(filepath.Join(rpmbldroot, "RPMS", rpmarch, srcname))
	handle_err(err)
	defer src.Close()

	_, err = io.Copy(dst, src)
	handle_err(err)
	err = dst.Sync()
	handle_err(err)

	if verbose {
		fmt.Printf("%s: building RPM [%s]...[ok]\n", n, fname)
	}
}

// EOF
