// We will be writing a simple HTTP server using the gin-gonic framework.
// It will respond to PUT, GET, and DELETE requests. The server will be completely stateless.
// Here is some sudo code for the PUT request:
// PUT {
// 1. transform input to a droplet locator
// if droplet exist {
//  if (droplet.creating) {
//     return status {
//        creating
//     }
//  }
//  connect to it -> next step
// } else {
//  create the droplet
//  return status {
//     creating
//  }
// }
//
// // next step (make sure we assume things could already exists
// ssh into droplet
// > // install git install docker, blah blah
// > git pull
// > docker compose up
//
// return status {
//    droplet.ip
//    healthy
// }
// }
//
// Here is some sudo code for the GET request:
// GET {
//
// if droplet exist {
//  if (droplet.creating) {
//     return status {
//        creating
//     }
//  }
//  connect to it -> next step
// } else {
//   return status {
//     doesnt eixt
//   }
// }
//
// ssh into droplet
// > git status
// > docker compose ls
//
// return status {
//    droplet.ip
//    healthy
// }
//
// DELETE {
// if droplet exist nuke it
// }
//
// Each endpoint will receive 2 inputs: a spec object and a status. The spec object will look like this:
// spec: {
//   id
//   name
//   source {
//     public-github
//     compose-file
//   }
//   target {
//      cloud: digitalocean
//      token
//      region?
//      size?
//      image?
//   }
//   config {
//      private-key
//   }
// }
// and the status represents the current state of the service, which is managed externally since this server is stateless.
// status: {
//    source: update-to-date
//    server: creating/healthy
//    ip-addres
//    exposed-ports
//    compose: up/down
// }
// Now generate the server code according to the above specs:

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Spec struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Source struct {
		PublicGithub string `json:"public-github"`
		ComposeFile  string `json:"compose-file"`
	} `json:"source"`
	Target struct {
		Cloud  string `json:"cloud"`
		Token  string `json:"token"`
		Region string `json:"region"`
		Size   string `json:"size"`
		Image  string `json:"image"`
	} `json:"target"`
	Config struct {
		PrivateKey string `json:"private-key"`
		PublicKey  string `json:"public-key"`
	} `json:"config"`
}

type Status struct {
	Source       string `json:"source"`
	ServerStatus string `json:"server"`
	IPAddress    string `json:"ip-address"`
	ExposedPorts string `json:"exposed-ports"`
	Compose      string `json:"compose"`
}

type Droplet struct {
	ID       string
	IP       string
	Creating bool
}

// Function for creating Digital ocean client using the godo package
func createDOClient(token string) *godo.Client {
	oauthClient := oauth2.NewClient(oauth2.NoContext, &tokenSource{
		AccessToken: token,
	})
	client := godo.NewClient(oauthClient)
	return client
}

func main() {
	router := gin.Default()

	// Initialize a map to store droplets
	droplets := make(map[string]*Droplet)

	// Handle PUT requests
	router.PUT("/", PutServer)

	// Handle GET requests
	router.GET("/", func(c *gin.Context) {
		var spec Spec
		err := c.BindJSON(&spec)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Transform input to a droplet locator
		dropletID := fmt.Sprintf("%s-%s-%s-%s", spec.Target.Cloud, spec.Target.Region, spec.Target.Size, spec.ID)

		// Check if droplet exists
		droplet, exists := droplets[dropletID]
		if exists {
			if droplet.Creating {
				c.JSON(http.StatusOK, Status{
					ServerStatus: "creating",
				})
			}

			// Connect to existing droplet
			// ...

		} else {
			c.JSON(http.StatusOK, Status{
				ServerStatus: "doesn't exist",
			})
			return
		}

		// SSH into droplet and execute commands
		// ...

		c.JSON(http.StatusOK, Status{
			IPAddress:    droplet.IP,
			ServerStatus: "healthy",
			ExposedPorts: "8080",
			Compose:      "up",
		})
	})

	// Handle DELETE requests
	router.DELETE("/", func(c *gin.Context) {
		var spec Spec
		err := c.BindJSON(&spec)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Transform input to a droplet locator
		dropletID := fmt.Sprintf("%s-%s-%s-%s", spec.Target.Cloud, spec.Target.Region, spec.Target.Size, spec.ID)

		// Check if droplet exists
		_, exists := droplets[dropletID]
		if exists {
			// Delete droplet
			delete(droplets, dropletID)

			c.JSON(http.StatusOK, gin.H{
				"message": "droplet deleted",
			})
		} else {
			c.JSON(http.StatusOK, Status{
				ServerStatus: "doesn't exist",
			})
		}
	})
	// Start the server
	log.Fatal(router.Run(":8080"))
}
