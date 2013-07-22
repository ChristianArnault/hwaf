package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestCernRootCint(t *testing.T) {
	{
		if err := exec.Command("which", "root-config").Run(); err != nil {
			t.Skip("skipping test. (needs CERN-ROOT (missing root-config))")

		}
		if err := exec.Command("which", "rootcint").Run(); err != nil {
			t.Skip("skipping test. (needs CERN-ROOT) (missing rootcint)")

		}
	}
	workdir, err := ioutil.TempDir("", "hwaf-test-")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.RemoveAll(workdir)

	err = os.Chdir(workdir)
	if err != nil {
		t.Fatalf(err.Error())
	}

	hwaf, err := newlogger("hwaf.log")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer hwaf.Close()

	for _, cmd := range [][]string{
		{"hwaf", "init", "-q=0", "."},
		{"hwaf", "setup", "-q=0"},
		{"hwaf", "pkg", "co", "-q=0", "git://github.com/hwaf/hwaf-tests-pkg-settings", "pkg-settings"},
		{"hwaf", "pkg", "co", "-q=0", "git://github.com/hwaf/hwaf-tests-rootcint-pkg1", "pkg1"},
		{"hwaf", "pkg", "co", "-q=0", "git://github.com/hwaf/hwaf-tests-rootcint-pkg2", "pkg2"},
		{"hwaf", "configure"},
		{"hwaf"},
		{"hwaf", "run", "root", "-l", "-b", "src/pkg2/share/run.C"},
	} {
		err := hwaf.Run(cmd[0], cmd[1:]...)
		if err != nil {
			hwaf.Display()
			t.Fatalf("cmd %v failed: %v", cmd, err)
		}
	}
}

// EOF
