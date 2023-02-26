package main

import (
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// Function for creating Digital ocean client using the godo package
func createDOClient(token string) *godo.Client {
	oauthClient := oauth2.NewClient(oauth2.NoContext, &tokenSource{
		AccessToken: token,
	})
	client := godo.NewClient(oauthClient)
	return client
}

func createDroplet(godoClient *godo.Client, spec Spec, c *gin.Context) (*godo.Droplet, error) {

	sshKey, _, err := godoClient.Keys.Create(c, &godo.KeyCreateRequest{
		Name:      fmt.Sprintf("%s-ssh-key", spec.Name),
		PublicKey: string(spec.Config.PublicKey),
	})
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return droplet, nil
}
