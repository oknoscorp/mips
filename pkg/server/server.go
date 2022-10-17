package server

import (
	"io/ioutil"
	"os"

	"github.com/oknos-ba/mips/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// configurationFile is path to YAML configuration file
const configurationFile = "configuration.yaml"

// directories holds a list of required directories.
// This list will be used to check if all folders
// exist with right permissions.
var directories = []string{"configuration", "postback_queue", "firmware_files"}

// files hold a list of required files
var files = []string{configurationFile}

type Configuration struct {
	FarmID      string   `yaml:"farm_id"`
	WorkersList []string `yaml:"workers_list"`
	Postbacks   []string `yaml:"postbacks"`
}

type Server struct {
	Configuration Configuration
}

// New will initialize server component.
func New() *Server {
	new := &Server{}
	new.createDirectories()
	new.createFiles()
	new.readConfig()
	return new
}

// createDirectories will create all necesary folders for data storing
// and caching.
func (server *Server) createDirectories() {
	for _, folder := range directories {
		if _, err := os.Stat(helpers.Path(folder)); os.IsNotExist(err) {
			err := os.Mkdir(helpers.Path(folder), os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// createFiles will create all required files
func (server *Server) createFiles() {
	for _, file := range files {
		if _, err := os.Stat(helpers.Path(file)); os.IsNotExist(err) {
			_, err := os.Create(helpers.Path(file))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// readConfig opens a .yaml configuration file and decodes it.
func (server *Server) readConfig() Configuration {
	yfile, err := ioutil.ReadFile(helpers.Path(configurationFile))
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yfile, &server.Configuration)
	if err != nil {
		log.Fatal(err)
	}

	return server.Configuration
}

// ConfigLists extracts the list of URL's that contain worker IP's.
func (server *Server) ConfigLists() []string {
	list := []string{}
	for _, link := range server.Configuration.WorkersList {
		list = append(list, link)
	}
	return list
}

// ConfigPostbacks extracts the list of URL's where postback requests will be
// broadcasted.
func (server *Server) ConfigPostbacks() []string {
	list := []string{}
	for _, link := range server.Configuration.Postbacks {
		list = append(list, link)
	}
	return list
}
