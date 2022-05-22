package tube_sock

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var ipMask string = "127.0.0.0/8"

func Start(appName string, lport int, ports []int) (err error) {
	pid := os.Getpid()

	//load tubular
	var msg string
	if msg, err = runSudoCmd("tubectl load"); err != nil {
		log.Printf("failed load tubectl: err: %v, msg: %s", err, msg)
		return
	}

	//first bind all configed ports
	for _, p := range ports {
		cmd := fmt.Sprintf("tubectl bind %s tcp %s %d", appName, ipMask, p)
		if msg, err = runSudoCmd(cmd); err != nil {
			log.Printf("failed to run cmd: %s\nerr: %v, msg: %s", cmd, err, msg)
		} else {
			log.Printf("succesfully bind: %s", cmd)
		}
	}

	//then register this server's listening socket
	cmd := fmt.Sprintf("tubectl register-pid %d %s tcp 127.0.0.1 %d", pid, appName, lport)
	if msg, err = runSudoCmd(cmd); err != nil {
		log.Printf("failed to register server socket: label: %s, port: %d, err: %v, msg: %s", appName, lport, err, msg)
	} else {
		log.Printf("successfully register: %s", cmd)
	}

	//log status
	msg, err = runSudoCmd("tubectl status")
	log.Printf("tubectl status: err: %v\n%s", err, msg)

	return
}

func Stop(appName string, ports []int) (err error) {
	//unbind ports
	var msg string
	for _, p := range ports {
		cmd := fmt.Sprintf("tubectl unbind %s tcp %s %d", appName, ipMask, p)
		if msg, err = runSudoCmd(cmd); err != nil {
			log.Printf("failed to run cmd: %s\nerr: %v, msg: %s", cmd, err, msg)
		} else {
			log.Printf("successfully unbind: %s", cmd)
		}
	}

	//unregister server socket
	if msg, err = runSudoCmd(fmt.Sprintf("tubectl unregister %s ipv4 tcp", appName)); err != nil {
		log.Printf("failed to unregister server socket: label: %s, err: %v, msg: %s", appName, err, msg)
	} else {
		log.Printf("successfully unregister: %s ipv4 tcp", appName)
	}

	//log status
	msg, err = runSudoCmd("tubectl status")
	log.Printf("tubectl status: err: %v\n%s", err, msg)

	return
}

func runSudoCmd(cmdLine string) (res string, err error) {
	dupPathEnv := fmt.Sprintf("PATH=%s", os.Getenv("PATH"))
	args := []string{dupPathEnv, "ASSUME_NO_MOVING_GC_UNSAFE_RISK_IT_WITH=go1.18", "/bin/sh", "-c"}
	args = append(args, cmdLine)
	cmd := exec.Command("sudo", args...)
	outerr, err := cmd.CombinedOutput()
	res = string(outerr)
	return
}
