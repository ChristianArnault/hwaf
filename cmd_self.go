package main

import (
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	//gocfg "github.com/gonuts/config"
)

func hwaf_make_cmd_self() *commander.Commander {
	cmd := &commander.Commander{
		Name:  "self",
		Short: "modify hwaf internal state",
		Commands: []*commander.Command{
			hwaf_make_cmd_self_init(),
			hwaf_make_cmd_self_bdist(),
			hwaf_make_cmd_self_bdist_upload(),
			hwaf_make_cmd_self_update(),
		},
		Flag: flag.NewFlagSet("hwaf-self", flag.ExitOnError),
	}
	return cmd
}

// EOF
