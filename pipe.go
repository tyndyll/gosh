package gosh

import (
	"bytes"
	"io"
	"log"
	"os/exec"
)

/* RunPipe takes a buffer to catch output and any number of exec.Cmd objects which it will run in order, exiting at any
 point if the Cmd does not return 0

 Example

 func main() {
	var b bytes.Buffer
	if err := Execute(&b,
		exec.Command("ls", "/Users/tyndyll/Downloads"),
		exec.Command("grep", "as"),
		exec.Command("sort", "-r"),
	); err != nil {
		log.Fatalln(err)
	}
	io.Copy(os.Stdout, &b)
 }
*/
func Execute(output_buffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var error_buffer bytes.Buffer
	pipe_stack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdin_pipe, stdout_pipe := io.Pipe()
		stack[i].Stdout = stdout_pipe
		stack[i].Stderr = &error_buffer
		stack[i+1].Stdin = stdin_pipe
		pipe_stack[i] = stdout_pipe
	}
	stack[i].Stdout = output_buffer
	stack[i].Stderr = &error_buffer

	if err := call(stack, pipe_stack); err != nil {
		log.Fatalln(string(error_buffer.Bytes()), err)
	}
	return err
}

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}
