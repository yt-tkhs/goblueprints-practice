package main

import "testing"

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar
	client := new(client)

	url, err := authAvatar.GetAvatarURL(client)

	if err != ErrNoAvatarURL {
		t.Error("Should return ErrNoAvatarURL if url is empty.")
	}

	testUrl := "http://url-to-avatar.com"
	client.userData = map[string]interface{} {
		"avatar_url": testUrl,
	}

	url, err = authAvatar.GetAvatarURL(client)
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
	client := new(client)
	client.userData = map[string]interface{} {
		"email": "MyEmailAddress@example.com",
	}

	url, err := gravatarAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("GravatarAvatar.GetAvatarURL should not return error.")
	}

	if url != "//www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346" {
		t.Errorf("GravatarAvatar.GetAvatarURL returns wrong url '%s'.", url)
	}
}