//
// Replicate objects between available blob-servers.
//

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/sundogrd/sdos/libconfig"
)

//
// Get the objects on the given server
//
func Objects(server string) []string {
	type list_strings []string
	var tmp list_strings

	//
	// Make the request to get the list of objects.
	//
	response, err := http.Get(server + "/blobs")
	if err != nil {
		log.Fatal(err)
	}

	//
	// Read the (JSON) response-body.
	//
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	//
	// Decode into an array of strings, and return it.
	//
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Fatal(err)
	}
	return tmp
}

//
// Does the given server contain the given object?
//
func HasObject(server string, object string) bool {

	response, err := http.Head(server + "/blob/" + object)
	if err != nil {
		fmt.Printf("Error Fetching: %s/blob/%s\n", server, object)
		return false
	} else {
		if response.StatusCode == 200 {
			fmt.Printf("\tObject %s is present on %s\n", object, server)
			return true
		} else {
			fmt.Printf("\tObject %s is missing on %s\n", object, server)
			return false
		}
	}
}

//
// Mirror the given object.
//
func MirrorObject(src string, dst string, obj string, options replicateCmd) bool {

	if options.verbose {
		fmt.Printf("\t\tMirroring %s from %s to %s\n", obj, src, dst)
	}

	//
	// Prepare to download the object.
	//
	src_url := fmt.Sprintf("%s%s%s", src, "/blob/", obj)
	fmt.Printf("\tFetching :%s\n", src_url)

	response, err := http.Get(src_url)

	//
	// If there was an error we're done.
	//
	if err != nil {
		fmt.Printf("Error fetching %s from %s%s%s\n",
			obj, src, "/blob/", obj)
		return false
	}

	//
	// Prepare to POST the body we've downloaded to
	// the mirror-location
	//
	dst_url := fmt.Sprintf("%s%s%s", dst, "/blob/", obj)
	fmt.Printf("\tUploading :%s\n", dst_url)

	//
	// Build up a new request.
	//
	child, _ := http.NewRequest("POST", dst_url, response.Body)

	//
	// Copy any X-Header which was present
	// in our download to the mirror.
	//
	for header, value := range response.Header {
		if strings.HasPrefix(header, "X-") {
			child.Header.Set(header, value[0])
		}
	}

	//
	// Send the request.
	//
	client := &http.Client{}
	r, err := client.Do(child)

	//
	// If there was no error we're good.
	//
	if err != nil {
		fmt.Printf("Error sending to %s - %s\n", dst_url, r.Body)
		return false
	}

	return true
}

//
// Sync the members of the set we're given.
//
func SyncGroup(servers []libconfig.BlobServer, options replicateCmd) {
	//
	// If we're being verbose show the members
	//
	if options.verbose {
		for _, s := range servers {
			fmt.Printf("\tGroup member: %s\n", s.Location)
		}
	}

	//
	// For each server - download the content-list here
	//
	//   key is server-name
	//   val is array of strings
	//
	objects := make(map[string][]string)

	//
	//  Store the list of objects each server hosts in the
	// hash, keyed upon the server-location/name.
	//
	for _, s := range servers {
		objects[s.Location] = Objects(s.Location)
	}

	//
	// Right we have a list of servers.
	//
	// For each server we also have the list of objects
	// that they contain.
	//
	for _, server := range servers {

		//
		// The objects on this server
		//
		var obs = objects[server.Location]

		//
		// For each object.
		//
		for _, i := range obs {

			//
			//  Mirror the object to every server that is not itself
			//
			for _, mirror := range servers {

				//
				// Ensure that src != dst.
				//
				if mirror.Location != server.Location {

					// If the object is missing.
					if !HasObject(mirror.Location, i) {
						MirrorObject(server.Location, mirror.Location, i, options)
					}
				}

			}
		}
	}
}

//
// This is where the main-work happens.
//
func replicate(options replicateCmd) {

	//
	// If we received blob-servers on the command-line use them too.
	//
	// NOTE: blob-servers added on the command-line are placed in the
	// "default" group.
	//
	if options.blob != "" {
		servers := strings.Split(options.blob, ",")
		for _, entry := range servers {
			libconfig.AddServer("default", entry)
		}
	} else {

		//
		//  Initialize the servers from our config file(s).
		//
		libconfig.InitServers()
	}

	//
	// Show the blob-servers.
	//
	if options.verbose {
		fmt.Printf("\t% 10s - %s\n", "group", "server")
		for _, entry := range libconfig.Servers() {
			fmt.Printf("\t% 10s - %s\n", entry.Group, entry.Location)
		}
	}

	//
	// Get a list of groups.
	//
	for _, entry := range libconfig.Groups() {

		if options.verbose {
			fmt.Printf("Syncing group: %s\n", entry)
		}

		//
		// For each group, get the members, and sync them.
		//
		SyncGroup(libconfig.GroupMembers(entry), options)
	}
}
