package main

import (
	"errors"
	"crypto/md5"
	"io"
	"strings"
	"fmt"
)

var ErrNoAvatarURL = errors.New("chat: can not fetch your avatar url.")

type Avatar interface {

	// Return avatar url
	// If an error has occurred, return ErrNoAvatarURL
	GetAvatarURL(c *client) (string, error)
}

type AuthAvatar struct {}
type GravatarAvatar struct {}

var UseAuthAvatar AuthAvatar
var UseGravatarAvatar GravatarAvatar

func (AuthAvatar) GetAvatarURL(c *client) (string, error) {
	if url, ok := c.userData["avatar_url"]; ok {
		if strUrl, ok := url.(string); ok {
			return strUrl, nil
		}
	}
	return "", ErrNoAvatarURL
}

func (GravatarAvatar) GetAvatarURL(c *client) (string, error) {
	if email, ok := c.userData["email"]; ok {
		if emailStr, ok := email.(string); ok {
			m := md5.New()
			io.WriteString(m, strings.ToLower(emailStr))
			return fmt.Sprintf("//www.gravatar.com/avatar/%x", m.Sum(nil)), nil

		}
	}
	return "", ErrNoAvatarURL
}