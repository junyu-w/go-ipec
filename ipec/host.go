package ipec

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/DrakeW/go-ipec/pb"
	"github.com/libp2p/go-libp2p-core/host"
	log "github.com/sirupsen/logrus"
)

// Node - represents a node in the network
type Node struct {
	host.Host
	ts *TaskService
}

// NewNode - create a new ndoe
func NewNode(ctx context.Context, h host.Host) *Node {
	n := &Node{Host: h}
	ts := NewTaskService(ctx, n)
	n.ts = ts
	return n
}

// HandleTaskRequest - Implements HandleTaskRequest of TaskPerformer
func (n *Node) HandleTaskRequest(req *pb.TaskRequest) (*pb.TaskResponse, error) {
	log.Infof("Start handling task - Task ID: %s", req.Task.TaskId)

	taskDir, err := setupTaskDir(req.Task)
	if err != nil {
		return nil, err
	}

	if err = executeTask(taskDir); err != nil {
		return nil, err
	}

	output, err := ioutil.ReadFile(filepath.Join(taskDir, "output"))
	if err != nil {
		return nil, err
	}

	return &pb.TaskResponse{
		Status:     pb.TaskResponse_DONE,
		TaskId:     req.Task.TaskId,
		Output:     output,
		FinishedAt: time.Now().Unix(),
		Performer: &pb.TaskPerformer{
			HostId: n.ID().Pretty(),
		},
	}, nil
}

// CreateTask - Implements CreateTask of TaskOwner
func (n *Node) CreateTask(function, input []byte, description string) *pb.Task { return nil }

// CreateTaskRequest - Implements CreateTaskRequest of TaskOwner
func (n *Node) CreateTaskRequest(*pb.Task) *pb.TaskRequest { return nil }

// Dispatch - Implements Dispatch of TaskOwner
func (n *Node) Dispatch(*pb.TaskRequest) (*pb.TaskResponse, error) { return nil, nil }

// HandleTaskResponse - Implements HandleTaskResponse of TaskOwner
func (n *Node) HandleTaskResponse(*pb.TaskResponse) error { return nil }

func setupTaskDir(task *pb.Task) (string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-", task.TaskId))
	if err != nil {
		return "", err
	}

	if err = ioutil.WriteFile(filepath.Join(dir, "func"), task.Function, os.FileMode(os.O_RDONLY|os.O_EXCL)); err != nil {
		return dir, err
	}
	if err = ioutil.WriteFile(filepath.Join(dir, "input"), task.Input, os.FileMode(os.O_RDONLY)); err != nil {
		return dir, err
	}
	if err = ioutil.WriteFile(filepath.Join(dir, "output"), []byte{}, os.FileMode(os.O_WRONLY)); err != nil {
		return dir, err
	}

	return dir, nil
}

func executeTask(taskDir string) error {
	inputFile, err := os.Open(filepath.Join(taskDir, "func"))
	if err != nil {
		return err
	}
	inputRdr := bufio.NewReader(inputFile)

	outputFile, err := os.Open(filepath.Join(taskDir, "output"))
	if err != nil {
		return err
	}
	outputWtr := bufio.NewWriter(outputFile)

	cmd := exec.Command(filepath.Join(taskDir, "func"))
	cmd.Stdin = inputRdr
	cmd.Stdout = outputWtr

	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}
