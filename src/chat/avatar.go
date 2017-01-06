package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"
)

var ErrNoAvatarURL = errors.New("chat: can not fetch your avatar url.")

type Avatar interface {

	// Return avatar url
	// If an error has occurred, return ErrNoAvatarURL
	GetAvatarURL(u ChatUser) (string, error)
}

type TryAvatars []Avatar

type AuthAvatar struct {}
type GravatarAvatar struct {}
type FileSystemAvatar struct {}

var UseAuthAvatar AuthAvatar
var UseGravatarAvatar GravatarAvatar
var UseFileSystemAvatar FileSystemAvatar

func (avatars TryAvatars) GetAvatarURL(u ChatUser) (string, error) {
	for _, avatar := range avatars {
		if url, err := avatar.GetAvatarURL(u); err == nil {
			return url, nil
		}
	}
	return "", ErrNoAvatarURL
}

func (AuthAvatar) GetAvatarURL(u ChatUser) (string, error) {
	url := u.AvatarURL()
	if url != "" {
		return url, nil
	}
	return "", ErrNoAvatarURL
}

func (GravatarAvatar) GetAvatarURL(u ChatUser) (string, error) {
	return "//www.gravatar.com/avatar/" + u.UniqueID(), nil
}

func (FileSystemAvatar) GetAvatarURL(u ChatUser) (string, error) {
	if files, err := ioutil.ReadDir("avatars"); err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if match, _ := filepath.Match(u.UniqueID() + "*", file.Name()); match {
				return "/avatars/" + file.Name(), nil
			}
		}
	}
	return "", ErrNoAvatarURL
}