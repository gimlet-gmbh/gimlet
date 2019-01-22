package gmbh

/*
 * gmbh.go
 * Abe Dick
 * Nov 2018
 */

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gimlet-gmbh/gimlet/gprint"
	"github.com/gimlet-gmbh/gimlet/gproto"
	yaml "gopkg.in/yaml.v2"
)

func handleDataRequest(req gproto.Request) (*gproto.Responder, error) {

	var request Request
	request = requestFromProto(req)
	responder := Responder{}

	handler, ok := g.registeredFunctions[request.Method]
	if !ok {
		responder.HadError = true
		responder.ErrorString = "Could not locate method in registered process map"
	} else {
		handler(request, &responder)
	}

	return responder.toProto(), nil
}

func createLog() {

}

func parseYamlConfig(relativePath string) (*Gimlet, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	var conf Gimlet
	yamlFile, err := ioutil.ReadFile(path + "/" + relativePath)
	if err != nil {
		gprint.Err(path+relativePath, 0)
		return nil, errors.New("could not find yaml file")
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, errors.New("could not unmarshal config")
	}
	return &conf, nil
}
