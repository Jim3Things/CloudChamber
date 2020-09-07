package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

/*
PUT api/Users/admin?op=login  <-- No password or credentials at this point
--- Log cookies
GET api/Racks
--- Add login cookie
--- Display output
GET api/Racks/{id} <-- Use the first link returned above, and for all other rack designators below
--- Display output
GET api/Racks/{id}/Blades
GET api/Racks/{id}/Blades/{bladeId} <-- Use the first link returned from the previous call
PUT api/Users/admin?op=logout
*/

// Command line: M1S -config=<global config file>
func main() {
	iow := exporters.NewExporter(exporters.NewIOWForwarder())
	exporters.Init(iow)

	if err := iow.Open(os.Stdout); err != nil {
		log.Fatalf("failed to open trace file: %v", err)
	}

	cfgPath := flag.String("config", ".", "path to the configuration file")
	showConfig := flag.Bool("showConfig", false, "display the current configuration settings")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	if *showConfig {
		fmt.Println(cfg)
		os.Exit(0)
	}

	endpoint := fmt.Sprintf("http://%s:%d", cfg.WebServer.FE.Hostname, cfg.WebServer.FE.Port)
	baseAddress := fmt.Sprintf("%s/api", endpoint)
	client := &http.Client{}

	// 1: try to login
	target := fmt.Sprintf("%s/users/admin?op=login", baseAddress)
	resp, err := put(client, target, nil, strings.NewReader(cfg.WebServer.SystemAccountPassword))
	if err != nil {
		panic(err)
	}

	dumpResponse(resp, err)

	// 1a: get list of known users
	target = fmt.Sprintf("%s/users", baseAddress)
	resp, err = get(client, target, resp.Cookies(), nil)
	if err != nil {
		panic(err)
	}

	dumpResponse(resp, err)

	// 2: get the list of racks
	target = fmt.Sprintf("%s/racks", baseAddress)
	resp, err = get(client, target, resp.Cookies(), nil)
	if err != nil {
		panic(err)
	}

	list := &pb.ExternalZoneSummary{}
	err = getJSONBody(resp, list)
	if err != nil {
		panic(err)
	}

	for name, item := range list.Racks {
		fmt.Printf("Found rack %q: %q\n", name, item.Uri)
	}

	// 3: get one rack's detail info
	for _, summary := range list.Racks {
		target = fmt.Sprintf("%s%s", endpoint, summary.Uri)
		break
	}

	resp, err = get(client, target, resp.Cookies(), nil)
	if err != nil {
		panic(err)
	}

	dumpResponse(resp, err)

	// 4: get one rack's list of blades
	target = fmt.Sprintf("%s/racks/rack1/blades", baseAddress)
	resp, err = get(client, target, resp.Cookies(), nil)
	if err != nil {
		panic(err)
	}

	body := dumpResponse(resp, err)
	lines := strings.Split(body, "\n")

	// 5: get one rack's first blade details
	target = fmt.Sprintf("%s%s", endpoint, lines[1])
	resp, err = get(client, target, resp.Cookies(), nil)
	if err != nil {
		panic(err)
	}

	dumpResponse(resp, err)

	// last: and now logout
	target = fmt.Sprintf("%s/users/admin?op=logout", baseAddress)
	resp, err = put(client, target, resp.Cookies(), nil)
	if err != nil {
		panic(err)
	}

	dumpResponse(resp, err)
}

func get(client *http.Client, uri string, cookies []*http.Cookie, body io.Reader) (*http.Response, error) {
	fmt.Printf("GET to %s\n", uri)
	req, err := http.NewRequest("GET", uri, body)
	if err != nil {
		return nil, err
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func put(client *http.Client, uri string, cookies []*http.Cookie, body io.Reader) (*http.Response, error) {
	fmt.Printf("PUT to %s\n", uri)
	req, err := http.NewRequest("PUT", uri, body)

	if err != nil {
		return nil, err
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func dumpResponse(resp *http.Response, err error) string {
	fmt.Printf("resp: %v\nerr: %v\n", resp, err)

	for _, cookie := range resp.Cookies() {
		fmt.Println(cookie.Name)
	}

	body, err := getBody(resp)
	if err != nil {
		panic(err)
	}

	s := string(body)

	fmt.Println(s)
	return s
}

func getBody(resp *http.Response) ([]byte, error) {
	defer func() { _ = resp.Body.Close() }()
	return ioutil.ReadAll(resp.Body)
}

// Get the body of a response, unmarshaled into the supplied message structure
func getJSONBody(resp *http.Response, v proto.Message) error {
	defer func() { _ = resp.Body.Close() }()
	return jsonpb.Unmarshal(resp.Body, v)
}
