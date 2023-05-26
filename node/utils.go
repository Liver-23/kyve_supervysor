package node

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func startNode() (*os.Process, error) {
	// TODO: Check if process.id is still running
	// TODO: Move filled address book, expose seeds

	app := "osmosisd"
	arg1 := "start"
	cmdPath, err := exec.LookPath(app)
	if err != nil {
		return nil, err
	}

	// TODO: Add exposed seeds from cmd input
	cmd := exec.Command(cmdPath, arg1)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	processIDChan := make(chan int)

	go func() {
		err := cmd.Start()
		if err != nil {
			fmt.Println(err)
			processIDChan <- -1
			return
		}

		processIDChan <- cmd.Process.Pid

		// Wait for process end
		err = cmd.Wait()
		if err != nil {
			fmt.Println(err)
			processIDChan <- -1
		}
	}()

	processID := <-processIDChan

	if processID == -1 {
		return nil, fmt.Errorf("Couldn't start running the node.")
	}

	process, err := os.FindProcess(processID)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func startGhostNode() {

	// TODO: move addressbook, change laddr port, expose no seeds

	app := "osmosisd"
	arg1 := "start"
	arg2 := "--p2p.seeds"
	arg3 := "_"
	arg4 := "--p2p.laddr"
	arg5 := "tcp://0.0.0.0:26658"

	cmd := exec.Command(app, arg1, arg2, arg3, arg4, arg5)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Print the output
	fmt.Println(string(stdout))
}

func shutdownNode() {
	// TODO: Expect process.id and shutdown the process
	process, err := os.FindProcess(ProcessId)
	if err != nil {
		fmt.Println("Error finding process.")
	}

	// Terminate the process
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Printf("Error terminating process: %s\n", err)
		return
	}

	fmt.Println("Process terminated successfully.")
}
