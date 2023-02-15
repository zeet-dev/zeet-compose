package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/gin-gonic/gin"
)

func PutServer(c *gin.Context) {
	var spec Spec
	err := c.BindJSON(&spec)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	godoClient := createDOClient(spec.Target.Token)

	// Get the droplet
	droplet, _, err := godoClient.Droplets.Get(c, spec.ID)
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

		sshKey, _, err := godoClient.Keys.Create(c, &godo.KeyCreateRequest{
			Name:      fmt.Sprintf("%s-ssh-key", spec.Name),
			PublicKey: string(spec.Config.PublicKey),
		})
		if err != nil {
			log.Fatal(err)
		}

		// Create droplet
		droplet, _, err := godoClient.Droplets.Create(c, &godo.DropletCreateRequest{
			Name:   spec.Name,
			Region: spec.Target.Region,
			Size:   spec.Target.Size,
			Image: godo.DropletCreateImage{
				Slug: "ubuntu-20-04-x64",
			},
			SSHKeys: []godo.DropletCreateSSHKey{
				{
					Fingerprint: sshKey.Fingerprint,
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, Status{
			ServerStatus: "creating",
		})
	}

	// SSH into droplet and execute commands
	// Create a new SSH client
	sshClient, err := createSSHClient(privateKey, droplet.Networks.V4[0].IPAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Install Docker Engine
	fmt.Println("Installing Docker Engine...")
	err = installDockerEngine(sshClient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Copying source code to the remote server...")
	scpClient, err := copyFolderToRemote(sshClient, "./nginx-golang", "/root/nginx-golang")
	defer scpClient.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(sshClient, "cd nginx-golang && docker compose up -d")
	if err != nil {
		log.Fatal(err)
	}

	// Update droplet status
	droplet.IP = "127.0.0.1" // Replace with actual IP
	droplet.Creating = false

	c.JSON(http.StatusOK, Status{
		IPAddress:    droplet.IP,
		ServerStatus: "healthy",
		Compose:      "up",
	})
}
