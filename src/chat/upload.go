package main

import (
	"net/http"
	"io"
	"io/ioutil"
	"path/filepath"
)

func uploaderHandler(w http.ResponseWriter, req *http.Request) {
	userId := req.FormValue("user_id")
	file, header, err := req.FormFile("avatarFile")

	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		io.WriteString(w, "ReadAll:" + err.Error())
		return
	}

	fileName := filepath.Join("avatars", userId + filepath.Ext(header.Filename))
	err = ioutil.WriteFile(fileName, data, 0777)
	if err != nil {
		io.WriteString(w, "WriteFile:" + err.Error())
		return
	}

	io.WriteString(w, "Success")
}
