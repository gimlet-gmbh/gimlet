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
func launchService(servicePath, validConfigPath, coreAddress string, verbose bool) {
	args := []string{"--remote", "--config=" + validConfigPath}

	if verbose {
		args = append(args, "--verbose")
	}

	cmd := exec.Command("gmbhProcm", args...)

	workingEnv := []string{
		"GMBHCORE=" + coreAddress,
		"SERVICEMODE=managed",
		"CONFIGPATH=" + validConfigPath,
	}

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirs := strings.Split(filepath.Dir(validConfigPath), string(filepath.Separator))
		f, err := getLogFile("gmbh", dirs[len(dirs)-1]+"-remote.log")
		if err == nil {
			cmd.Stdout = f
			cmd.Stderr = f
		}
		workingEnv = append(workingEnv, "LOGPATH="+filepath.Join(getpwd(), "gmbh"))
		workingEnv = append(workingEnv, "LOGNAME="+dirs[len(dirs)-1])
		notify.LnBYellowF("Log=%s", filepath.Join(getpwd(), "gmbh", dirs[len(dirs)-1]+"-remote.log"))
	}

	cmd.Env = append(os.Environ(), workingEnv...)

	err := cmd.Start()
	if err != nil {
		notify.LnCyanF("could not start remote")
	}

}
