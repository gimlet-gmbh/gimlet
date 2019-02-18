package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gmbh-micro/notify"
)

// basePath attempts to get the absolute path to the directory in which the config file is specified
func basePath(configPath string) string {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		notify.LnRedF("error=%v", err.Error())
		return ""
	}
	return filepath.Dir(abs)
}

// scanForServices scans for directories (or symbolic links to directories)
// that containa gmbh config file and returns an array of absolute paths
// to any found directories that contain the config file
// TODO: Need to verify that we are getting the correct yaml file
// if there are several yaml files and if there are no yaml
func scanForServices(baseDir string) ([]string, error) {
	servicePaths := []string{}

	baseDirFiles, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return servicePaths, errors.New("could not scan base directory: " + err.Error())
	}

	for _, file := range baseDirFiles {

		// eval symbolic links
		fpath := filepath.Join(baseDir, file.Name())
		potentialSymbolic, err := filepath.EvalSymlinks(fpath)
		if err != nil {
			notify.LnRedF(err.Error(), 0)
			continue
		}

		// If it wasn't a symbolic path check if it was a dir, skip if not
		if fpath == potentialSymbolic {
			if !file.IsDir() {
				continue
			}
		}

		// Try and open the symbolic link path and check for dir, skip if not
		newFile, err := os.Stat(potentialSymbolic)
		if err != nil {
			notify.LnRedF(err.Error())
			continue
		}

		if !newFile.IsDir() {
			continue
		}

		// Looking through potential gmbH service directory
		serviceFiles, err := ioutil.ReadDir(filepath.Join(baseDir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		for _, sfile := range serviceFiles {
			match, err := regexp.MatchString(".toml", sfile.Name())
			if err == nil && match {
				servicePaths = append(servicePaths, filepath.Join(baseDir, file.Name(), sfile.Name()))
			}
		}
	}

	return servicePaths, nil
}

// launch service fork and exec's using gmbh remote with config path set to the known config path
func launchService(validConfigPaths []string, servicesDir string, coreAddress string) {
	if len(validConfigPaths) == 0 {
		return
	}

	args := []string{"--remote"}
	for _, a := range validConfigPaths {
		args = append(args, "--config="+a)
	}
	if *l.verboseAll {
		args = append(args, "--verbose")
	}

	cmd := exec.Command("gmbhProcm", args...)
	cmd.Dir = servicesDir

	workingEnv := []string{
		"GMBHCORE=" + coreAddress,
		"SERVICEMODE=managed",
	}

	logName := genLogFileName(validConfigPaths)

	if *l.verboseAll {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {

		f, err := getLogFile("logs", logName+"-remote.log")
		if err == nil {
			notify.LnYellowF("%s", filepath.Join(notify.Getpwd(), "logs", logName+"-remote.log"))
			cmd.Stdout = f
			cmd.Stderr = f
		} else {
			notify.LnBRedF("could not create log file: " + logName)
		}
	}

	workingEnv = append(workingEnv, "LOGFILENAME="+logName+"-data.log")
	cmd.Env = append(os.Environ(), workingEnv...)

	err := cmd.Start()
	if err != nil {
		notify.LnCyanF("could not start remote")
	}

}

// genLogFileName generates a filename by discarding everything except the last directory name,
// joins them together with a _
func genLogFileName(configPaths []string) string {
	dirName := []string{}
	for _, c := range configPaths {
		dirs := strings.Split(filepath.Dir(c), string(filepath.Separator))
		dirName = append(dirName, dirs[len(dirs)-1])
	}
	return strings.Join(dirName, "_")
}
