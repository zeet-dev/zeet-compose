package main

import (
	"log"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/gin-gonic/gin"
)

func GetServer(c *gin.Context) {
	var spec Spec
	err := c.BindJSON(&spec)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	doClient := createDOClient(spec.Target.Token)

	// ASSUMPTION: ID property on spec is the Droplet ID
	droplet, _, err := doClient.Droplets.Get(c, spec.ID)
	dropletExists := true
	if err != nil {
		// Check if the error is a not found error
		log.Println(err)
		if _, ok := err.(*godo.ErrorResponse); ok {
			dropletExists = false
		} else {
			log.Fatal(err)
		}
	}

	if dropletExists {
		if droplet.Status != "active" {
			c.JSON(http.StatusOK, Status{
				ServerStatus: "creating",
			})
		}
	} else {
		c.JSON(http.StatusOK, Status{
			ServerStatus: "non-existent",
		})
	}

	// Droplet exists and is active
	// SSH into droplet and check status of commands
	// c.JSON(http.StatusOK, Status{
	// 	IPAddress:    droplet.IP,
	// 	ServerStatus: "healthy",
	// 	ExposedPorts: "8080",
	// 	Compose:      "up",
	// })
}
