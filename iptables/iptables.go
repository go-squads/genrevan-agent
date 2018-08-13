package iptables

import (
	"errors"
	"fmt"
	"os/exec"
)

type Rule struct {
	SourceIP        string
	SourcePort      string
	DestinationIP   string
	DestinationPort string
}

func getIPTablesPath() (string, error) {
	path, err := exec.LookPath("iptables")
	if err != nil {
		return "", err
	}

	return path, nil
}

func execute(args string) (string, error) {
	path, err := getIPTablesPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("/bin/sh", "-c", "sudo "+path+" "+args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func isRuleExists(r Rule) (bool, error) {
	args := fmt.Sprintf("-t nat -C PREROUTING -i enp0s8 -p TCP -d %s --dport %s -j DNAT --to-destination %s:%s -m comment --comment 'forward to the Nginx container'", r.SourceIP, r.SourcePort, r.DestinationIP, r.DestinationPort)
	_, err := execute(args)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func Insert(r Rule) error {
	exist, err := isRuleExists(r)
	if err != nil {
		return err
	}

	if !exist {
		args := fmt.Sprintf("-t nat -I PREROUTING -i enp0s8 -p TCP -d %s --dport %s -j DNAT --to-destination %s:%s -m comment --comment 'forward to the Nginx container'", r.SourceIP, r.SourcePort, r.DestinationIP, r.DestinationPort)
		result, err := execute(args)
		if err != nil {
			return err
		}

		if len(result) > 0 {
			return errors.New(result)
		}
	}

	return nil
}
