package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

func main() {
	router := gin.Default()

	// Initialize a map to store droplets
	droplets := make(map[string]*Droplet)

	// Handle PUT requests
	router.PUT("/", PutServer)

	// Handle GET requests
	router.GET("/", GetServer)

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
