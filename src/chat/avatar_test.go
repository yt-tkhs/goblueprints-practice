package main

import (
	"testing"
	"path/filepath"
	"io/ioutil"
	"os"
	"github.com/stretchr/gomniauth/test"
)

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar

	testUser := &test.TestUser{}
	testUser.On("AvatarURL").Return("", ErrNoAvatarURL)

	testChatUser := &chatUser{User: testUser}

	url, err := authAvatar.GetAvatarURL(testChatUser)
	if err != ErrNoAvatarURL {
		t.Error("Should return ErrNoAvatarURL if url is empty.")
	}

	testUrl := "http://url-to-avatar/"
	testUser = &test.TestUser{}
	testChatUser.User = testUser
	testUser.On("AvatarURL").Return(testUrl, nil)
	url, err = authAvatar.GetAvatarURL(testChatUser)
	if err != nil {
		t.Error("Should not return an error if url is not empty.")
	} else {
		if url != testUrl {
			t.Error("AuthAvatar.GetAvatarURL should return correct url.")
		}
	}
}

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar GravatarAvatar
	user := &chatUser{uniqueID: "abc"}

	url, err := gravatarAvatar.GetAvatarURL(user)
	if err != nil {
		t.Error("GravatarAvatar.GetAvatarURL should not return error.")
	}

	if url != "//www.gravatar.com/avatar/abc" {
		t.Errorf("GravatarAvatar.GetAvatarURL returns wrong url '%s'.", url)
	}
}

func TestFileSystemAvatar(t *testing.T) {
	fileName := filepath.Join("avatars", "abc.jpg")
	ioutil.WriteFile(fileName, []byte{}, 0777)
	defer func() { os.Remove(fileName) }()

	var fileSystemAvatar FileSystemAvatar
	user := &chatUser{uniqueID: "abc"}

	url, err := fileSystemAvatar.GetAvatarURL(user)

	if err != nil {
		t.Error("FileSystemAvatar.GetAvatarURL should not return error.")
	}

	if url != "/avatars/abc.jpg" {
		t.Errorf("FileSystemAvatar.GetAvatarURL returns wrong url '%s'", url)
	}
}