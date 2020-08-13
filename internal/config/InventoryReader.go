// Inventory reader parses the YAML file and returns Zone. into a pb external zone. 

package main

import(

	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"context"
	
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

func (x *EXternalZone) Reset(){

	yamlFile, err := ioutil.ReadFile("C:\Users\Waheguru\go\src\github.com\Jim3Things\CloudChamber\configs\Inventory.yaml")
	if err != nil {
		log.PrintF("yamlFile.Get err  #%v", err)
	}
	err = yaml.marshal(yamlFile, x)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return x
}