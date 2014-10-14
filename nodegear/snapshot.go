package nodegear

import (
	"../models"
	"../config"
	"../connection"
	"bytes"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"io"
	"fmt"
	"path/filepath"
	"bufio"
	"mime/multipart"
	"time"
)

func (p *Instance) ApplySnapshot(process *models.AppProcess) {
	if config.Configuration.Storage.Enabled == false {
		return
	}

	client := http.Client{}
	request, err := http.NewRequest("GET", config.Configuration.Storage.Server + "/snapshots/" + process.DataSnapshot.Hex() + ".diff", bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", config.Configuration.Storage.Auth)

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println("Get Snapshot:", response.Status)
	
	if response.StatusCode != 200 {
		return
	}
	
	snapshot_file, err := os.Create("/tmp/snapshot_" + process.DataSnapshot.Hex() + ".diff")
	if err != nil {
		panic(err)
	}
	
	defer func() {
		if err := snapshot_file.Close(); err != nil {
			panic(err)
		}
	}()

	writer := bufio.NewWriter(snapshot_file)
	reader := bufio.NewScanner(response.Body)

	newline := []byte("\n")

	file_bytes := 0
	for reader.Scan() {
		line := append(reader.Bytes(), newline[0])

		file_bytes += len(line)
		if _, err := writer.Write(line); err != nil {
			panic(err)
		}
	}

	fmt.Println("Received snapshot size:", file_bytes, "bytes")

	if err := reader.Err(); err != nil {
		panic(err)
	}
	if err := writer.Flush(); err != nil {
		panic(err)
	}
}

func (p *Instance) SaveSnapshot(id bson.ObjectId, path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	num_bytes := stat.Size()
	fmt.Println("Snapshot is", stat.Size(), "bytes")
	if num_bytes == 0 {
		// Nothing to save
		return
	}

	c := connection.MongoC(models.AppProcessDataSnapshotC)
	snapshot := models.AppProcessDataSnapshot{
		ID: id,
		Created: time.Now(),
		App: p.App_id,
		OriginProcess: p.Process_id,
		OriginServer: Server.ID,
		ContentSize: int(num_bytes),
	}
	if err := c.Insert(&snapshot); err != nil {
		panic(err)
	}

	err = connection.MongoC(models.AppProcessC).UpdateId(p.Process_id, &bson.M{
		"$set": &bson.M{
			"dataSnapshot": snapshot.ID,
		},
	})
	if err != nil {
		panic(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(id.Hex() + ".diff", filepath.Base(path))
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	request, err := http.NewRequest("POST", config.Configuration.Storage.Server + "/snapshots/", body)
	if err != nil {
		panic(err)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", config.Configuration.Storage.Auth)

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.Status)
	
	if response.StatusCode != 200 {
		return
	}
}
