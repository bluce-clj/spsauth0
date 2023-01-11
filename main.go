/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import "github.com/bluce-clj/spsauth0/cmd"

func main() {
	cmd.Execute()
}



// Do a tenant dump and store it in local cache. use that to search for things
	// refresh dump daily or force get new dump with a command


// Pre-load tenant info
// Make secret info hide when entering or viewing ***t

// Client search flags on what field on client object to search on
	// not sure how to display this in the gocui

// update client list to display clients for specific tenant - done
	// optional --all flag to show all clients

// client token --various flags to generate auth token requests in curl/python/etc
	//	https://learning.postman.com/docs/sending-requests/generate-code-snippets/
	//	https://stackoverflow.com/questions/33068128/whats-the-optimal-way-to-execute-a-nodejs-script-from-golang-that-returns-a-st

// User tokens
	// Need to preload gotjwt test and prod applications for this to work - maybe

// Remove Token from client

// Client add/update have client default audience = use default audience provided by the given tenant
	// make this editable so you could make a client audience for the management api instead

// Tenant update allow update APIs - how to handle this. Should API be there own cmd? probably not
// If you set a default client on a tenant you shoudl verifity that the client is set up for your tenant - it already might be

// Client search flags on what field on client object to search on
	// not sure how to display this in the gocui
// Extract tenantconfig/ clientConfig get into utils package

// Do a tenant dump and store it in local cache. use that to search for things
	// refresh dump daily or force get new dump with a command

// add client update

// add state query parameter to wsa, native and spa auth flow