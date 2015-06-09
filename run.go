package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

var cmdRun = &Command{
	Exec:        runRun,
	UsageLine:   "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
	Description: "Run a command in a new server",
	Help:        "Run a command in a new server.",
	Examples: `
    $ scw run ubuntu-trusty
    $ scw run --name=mydocker docker docker run moul/nyancat:armhf
    $ scw run --bootscript=3.2.34 --env="boot=live rescue_image=http://j.mp/scaleway-ubuntu-trusty-tarball" 50GB bash
`,
}

func init() {
	cmdRun.Flag.StringVar(&runCreateName, []string{"-name"}, "", "Assign a name")
	cmdRun.Flag.StringVar(&runCreateBootscript, []string{"-bootscript"}, "", "Assign a bootscript")
	cmdRun.Flag.StringVar(&runCreateEnv, []string{"e", "-env"}, "", "Provide metadata tags passed to initrd (i.e., boot=resue INITRD_DEBUG=1)")
	cmdRun.Flag.StringVar(&runCreateVolume, []string{"v", "-volume"}, "", "Attach additional volume (i.e., 50G)")
	cmdRun.Flag.BoolVar(&runHelpFlag, []string{"h", "-help"}, false, "Print usage")
	// FIXME: handle start --timeout
}

// Flags
var runCreateName string       // --name flag
var runCreateBootscript string // --bootscript flag
var runCreateEnv string        // -e, --env flag
var runCreateVolume string     // -v, --volume flag
var runHelpFlag bool           // -h, --help flag

func runRun(cmd *Command, args []string) {
	if runHelpFlag {
		cmd.PrintUsage()
	}
	if len(args) < 1 {
		cmd.PrintShortUsage()
	}

	//image := args[0]

	// create IMAGE
	log.Debugf("Creating a new server")
	serverID, err := createServer(cmd.API, args[0], runCreateName, runCreateBootscript, runCreateEnv, runCreateVolume)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	log.Debugf("Created server: %s", serverID)

	// start SERVER
	log.Debugf("Starting server")
	err = startServer(cmd.API, serverID, false)
	if err != nil {
		log.Fatalf("Failed to start server %s: %v", serverID, err)
	}
	log.Debugf("Server is booting")

	// waiting for server to be ready
	log.Debugf("Waiting for server to be ready")
	server, err := WaitForServerReady(cmd.API, serverID)
	if err != nil {
		log.Fatalf("Cannot get access to server %s: %v", serverID, err)
	}
	log.Debugf("Server is ready: %s", server.PublicAddress.IP)

	// exec -w SERVER COMMAND ARGS...
	log.Debugf("Executing command")
	if len(args) < 2 {
		err = sshExec(server.PublicAddress.IP, []string{"if [ -x /bin/bash ]; then /bin/bash; else /bin/sh; fi"}, false)
	} else {
		err = sshExec(server.PublicAddress.IP, args[1:], false)
	}
	if err != nil {
		log.Debugf("Command execution failed: %v", err)
		os.Exit(1)
	}
	log.Debugf("Command successfuly executed")
}
