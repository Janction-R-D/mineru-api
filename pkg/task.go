package pkg

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"time"
)

const TaskStatusPending = "pending"
const TaskStatusRunning = "running"
const TaskStatusSuccess = "success"
const TaskStatusFailed = "failed"

var VlmSglangClientUrl = ""

func addTask(t Task) {
	taskMap.mutex.Lock()
	defer taskMap.mutex.Unlock()
	taskMap.tasks[t.TaskId] = t
}

func delTask(taskID string) {
	taskMap.mutex.Lock()
	defer taskMap.mutex.Unlock()
	if _, ok := taskMap.tasks[taskID]; !ok {
		return
	}
	delete(taskMap.tasks, taskID)
}

func getTask(taskID string) (Task, bool) {
	taskMap.mutex.Lock()
	defer taskMap.mutex.Unlock()
	t, ok := taskMap.tasks[taskID]
	return t, ok
}

func setTaskStatus(taskID string, status, msg string) {
	taskMap.mutex.Lock()
	defer taskMap.mutex.Unlock()
	t, ok := taskMap.tasks[taskID]
	if !ok {
		return
	}
	t.Status = status
	t.Msg = msg
	taskMap.tasks[t.TaskId] = t
}

const defaultPath = "/opt/parse"

func Init(vlmSglangClientUrl string) {
	_ = os.RemoveAll(defaultPath)
	VlmSglangClientUrl = vlmSglangClientUrl

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	var check = func() {
		for _, v := range taskMap.tasks {
			if v.Status == TaskStatusSuccess || v.Status == TaskStatusFailed {
				if time.Since(v.CreateTime) > 20*time.Minute {
					_ = os.Remove(v.Path)
					delTask(v.TaskId)
				}
			}
		}
	}

	for {
		select {
		case <-ticker.C:
			check()
		}
	}
}

func run(t Task) {
	u, err := url.Parse(t.Url)
	if err != nil {
		return
	}
	t.FileName = path.Base(u.Path)
	t.Status = TaskStatusRunning
	addTask(t)

	outputPath := fmt.Sprintf("%v/%v", defaultPath, t.TaskId)
	os.MkdirAll(outputPath, 0777)

	downloadCmd := fmt.Sprintf("wget -p %s %s", outputPath, t.Url)
	msg, err := ExecuteCommand(downloadCmd)
	if err != nil {
		t.Status = TaskStatusFailed
		t.Msg = msg
		return
	}

	parseCmd := fmt.Sprintf("mineru -p %s/%s -o %v -b vlm-sglang-client -u %s", outputPath, t.FileName, outputPath, VlmSglangClientUrl)
	msg, err = ExecuteCommand(parseCmd)
	if err != nil {
		t.Status = TaskStatusFailed
		t.Msg = msg
		addTask(t)
		return
	}

	mdPath := fmt.Sprintf("%v/%v.md", t.Path, GetFileNameWithoutExt(t.FileName))
	t.Status = TaskStatusSuccess
	if !FileExists(mdPath) {
		t.Status = TaskStatusFailed
	}
	t.Msg = msg
	t.Path = outputPath
	addTask(t)
	return
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if err != nil {
		return stdout, fmt.Errorf("exec cmd: %v,stderr : %s err %v", command, stderr, err.Error())
	}

	if stderr != "" {
		return stderr, fmt.Errorf("exec cmd: %v,stderr : %s ", command, stderr)
	}

	return stdout, nil
}
