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

func getNetfilterPersistentPath() (string, error) {
	path, err := exec.LookPath("netfilter-persistent")
	if err != nil {
		return "", err
	}

	return path, nil
}

func getIPTablesPath() (string, error) {
	path, err := exec.LookPath("iptables")
	if err != nil {
		return "", err
	}

	return path, nil
}

func execute(path, args string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", "sudo "+path+" "+args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func isRuleExists(r Rule) (bool, error) {
	path, err := getIPTablesPath()
	if err != nil {
		return false, err
	}

	args := fmt.Sprintf("-t nat -C PREROUTING -i enp0s8 -p TCP -d %s --dport %s -j DNAT --to-destination %s:%s", r.SourceIP, r.SourcePort, r.DestinationIP, r.DestinationPort)
	_, err = execute(path, args)
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
		path, err := getIPTablesPath()
		if err != nil {
			return err
		}

		args := fmt.Sprintf("-t nat -I PREROUTING -i enp0s8 -p TCP -d %s --dport %s -j DNAT --to-destination %s:%s", r.SourceIP, r.SourcePort, r.DestinationIP, r.DestinationPort)
		result, err := execute(path, args)
		if err != nil {
			return err
		}

		if len(result) > 0 {
			return errors.New(result)
		}
	}

	return nil
}

func Save() error {
	path, err := getNetfilterPersistentPath()
	if err != nil {
		return err
	}

	_, err = execute(path, "save")
	if err != nil {
		return err
	}

	return nil
}
