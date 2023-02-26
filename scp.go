package main

import (
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"
)

func copyFolderToRemote(sshClient *ssh.Client, source string, destination string) (*scp.Client, error) {

	scpClient, err := scp.NewClientFromExistingSSH(sshClient, &scp.ClientOption{})
	if err != nil {
		return nil, err
	}

	err = scpClient.CopyDirToRemote(source, destination, &scp.DirTransferOption{})
	if err != nil {
		return nil, err
	}

	return scpClient, nil
}
