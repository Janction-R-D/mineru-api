package pkg

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ExtractTaskResponse struct {
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
	Data ExtractTaskResponseData `json:"data"`
}

type ExtractTaskResponseData struct {
	TaskID string `json:"task_id"`
}

type Task struct {
	TaskId     string    `json:"task_id"`
	Url        string    `json:"url"`
	Status     string    `json:"status"`
	Msg        string    `json:"msg"`
	CreateTime time.Time `json:"create_time"`
	Path       string    `json:"path"`
	FileName   string    `json:"file_name"`
}

var taskMap struct {
	tasks map[string]Task
	mutex sync.RWMutex
}

func ExecExtractTask(req ExtractTaskRequest) (resp ExtractTaskResponse) {
	taskID := uuid.New().String()
	resp.Code = 0
	resp.Msg = "success"
	resp.Data.TaskID = taskID
	t := Task{
		TaskId:     taskID,
		Url:        req.Url,
		Status:     TaskStatusPending,
		CreateTime: time.Now(),
	}
	addTask(t)
	go run(t)
	return resp
}

type GetExtractTaskDetailResponse struct {
	Code int                              `json:"code"`
	Msg  string                           `json:"msg"`
	Data GetExtractTaskDetailResponseData `json:"data"`
}

type GetExtractTaskDetailResponseData struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

func GetExtractTaskDetail(id string) (resp GetExtractTaskDetailResponse, mdPath string, name string) {
	t, ok := getTask(id)
	resp.Data.TaskID = t.TaskId
	if !ok {
		resp.Code = 404
		resp.Msg = "task not found"
		return resp, mdPath, name
	}
	resp.Data.Status = t.Status
	if t.Status == TaskStatusFailed {
		resp.Code = 500
		resp.Msg = t.Msg
		return resp, mdPath, name
	}

	if t.Status == TaskStatusSuccess {
		mdPath = fmt.Sprintf("%v/%v.md", t.Path, GetFileNameWithoutExt(t.FileName))
		resp.Code = 200
		resp.Msg = t.Msg
		return resp, mdPath, GetFileNameWithoutExt(t.FileName) + ".md"
	}
	resp.Code = 500
	resp.Msg = t.Msg
	return resp, mdPath, name
}
