package storage

import (
	"bash-commands/internal/cashe"
	"bash-commands/internal/storage/postgresql"
	"bash-commands/server/command_result"
	"bash-commands/server/get_all_commands"
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
)

type Storage struct {
	db          *postgresql.Postgres
	runningProc *cashe.Cache
}

func New(db string) *Storage {
	st := Storage{}
	st.db = postgresql.NewPostgresRepository(db)
	st.runningProc = cashe.NewCache()
	return &st
}

func (s Storage) AddAndRun(command string) (int64, *exec.Cmd, error) {
	id, err := s.db.InsertCommand(command)
	if err != nil {
		return 0, nil, err
	}
	cmd := exec.Command("bash", "-c", command)
	err = cmd.Err
	if err != nil {
		return 0, nil, err
	}
	return id, cmd, nil
}
func (s Storage) Start(cmd *exec.Cmd) (int, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	err = cmd.Start()
	if err != nil {
		return 0, err
	}

	s.runningProc.AddOne(cmd.Process.Pid, stdout, stderr)
	return cmd.Process.Pid, nil
}
func (s Storage) Exec(cmd *exec.Cmd, pid int, id int64) (*command_result.CommandResult, error) {
	stdout, stderr := s.runningProc.GetOne(pid)
	if stdout == nil || stderr == nil {
		return nil, errors.New("can't get command output")
	}
	var stdoutStr, stderrStr string

	go func(stdout io.Reader) {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutStr += line + "\n"
		}
	}(stdout)

	go func(stderr io.Reader) {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrStr += line + "\n"
		}
	}(stderr)

	if err := cmd.Wait(); err != nil {
		if stderrStr == "" {
			return nil, err
		}
	}
	s.runningProc.Stop(pid)
	err := s.db.AddResult(id, pid, stdoutStr, stderrStr)
	if err != nil {
		return nil, err
	}
	return &command_result.CommandResult{
		PID:           pid,
		Output:        stdoutStr,
		CommandErrors: stderrStr,
	}, nil
}

func (s Storage) FindAndRun(id int64) (*exec.Cmd, error) {
	command, err := s.db.SelectOne(id)
	if err != nil || command == "" {
		return nil, err
	}
	cmd := exec.Command("bash", "-c", command)
	err = cmd.Err
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (s Storage) Kill(pid int) error {
	if !s.runningProc.Check(pid) {
		return errors.New("pid not exists")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	err = process.Signal(os.Interrupt)
	if err != nil {
		return err
	}
	return nil
}

func (s Storage) Stop() error {
	return s.db.Stop()
}

func (s Storage) GetAll() (*get_all_commands.Response, error) {
	res, err := s.db.SelectAll()
	if err != nil {
		return nil, err
	}
	ans := get_all_commands.Response{}

	for _, commands := range *res {
		command := get_all_commands.Command{
			Id:      commands.Id,
			Command: commands.Command,
		}
		ans.Commands = append(ans.Commands, command)
	}
	return &ans, nil
}

func (s Storage) Delete(id int64) (bool, error) {
	return s.db.Delete(id)
}

func (s Storage) Get(id int64) (string, error) {
	return s.db.SelectOne(id)
}
