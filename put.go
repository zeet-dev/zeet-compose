package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/digitalocean/godo"
	"github.com/gin-gonic/gin"
)

func PutServer(c *gin.Context) {
	var spec Spec
	err := c.BindJSON(&spec)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	doClient := createDOClient(spec.Target.Token)

	// Get the droplet
	tag := "zeet-" + spec.ID
	droplets, _, err := doClient.Droplets.ListByTag(c, tag, &godo.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	if len(droplets) == 0 {
		// Create droplet and return creating. Once done creating, do we need to immediately start provisioning the docker compose thing?
	} else {

	}

	// droplet, _, err := doClient.Droplets.Get(c, spec.ID)
	// dropletExists := true
	// if err != nil {
	// 	// Check if the error is a not found error
	// 	log.Println(err)
	// 	if _, ok := err.(*godo.ErrorResponse); ok {
	// 		dropletExists = false
	// 	} else {
	// 		log.Fatal(err)
	// 	}
	// }

	if dropletExists {
		if droplet.Status != "active" {
			c.JSON(http.StatusOK, Status{
				ServerStatus: "creating",
			})
		}
	} else {
		droplet, err = createDroplet(doClient, spec, c)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, Status{
			ServerStatus: "creating",
		})
	}

	// Droplet exists and is active
	// SSH into droplet and execute commands
	// Create a new SSH client
	sshClient, err := createSSHClient([]byte(spec.Config.PrivateKey), droplet.Networks.V4[0].IPAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Installing Docker Engine...")
	err = installDockerEngine(sshClient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Cloning Github repository...")
	err = pullGithubRepo(sshClient, spec.Source.PublicGithub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Run docker compose up")
	// Get repo name from public github url
	u, err := url.Parse(spec.Source.PublicGithub)
	if err != nil {
		log.Fatal(err)
	}
	repoName := path.Base(u.Path)
	err = runCommand(sshClient, fmt.Sprintf("cd ~/%s && docker compose up -d", repoName))
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, Status{
		IPAddress:    droplet.Networks.V4[0].IPAddress,
		ServerStatus: "healthy",
		Compose:      "up",
	})
}
