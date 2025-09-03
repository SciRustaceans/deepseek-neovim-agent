package nvim

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type RPCClient struct {
	conn net.Conn
}

func NewRPCClient(socketPath string) (*RPCClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neovim: %v", err)
	}

	return &RPCClient{conn: conn}, nil
}

func (c *RPCClient) Call(method string, args ...interface{}) (interface{}, error) {
	request := []interface{}{0, method, args}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	jsonData = append(jsonData, '\n')

	if _, err := c.conn.Write(jsonData); err != nil {
		return nil, err
	}

	// Set read timeout
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buffer := make([]byte, 4096)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	var response []interface{}
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		return nil, err
	}

	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}

	// Check for error (response[2] would contain error if present)
	if len(response) > 2 && response[2] != nil {
		return nil, fmt.Errorf("RPC error: %v", response[2])
	}

	return response[1], nil
}

func (c *RPCClient) Close() error {
	return c.conn.Close()
}

func (c *RPCClient) GetCurrentBufferContent() (string, error) {
	result, err := c.Call("nvim_get_current_buf")
	if err != nil {
		return "", err
	}

	bufID, ok := result.(float64)
	if !ok {
		return "", fmt.Errorf("invalid buffer ID")
	}

	content, err := c.Call("nvim_buf_get_lines", int(bufID), 0, -1, true)
	if err != nil {
		return "", err
	}

	lines, ok := content.([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	var code string
	for _, line := range lines {
		if str, ok := line.(string); ok {
			code += str + "\n"
		}
	}

	return code, nil
}

func (c *RPCClient) ReplaceBufferContent(newContent string) error {
	result, err := c.Call("nvim_get_current_buf")
	if err != nil {
		return err
	}

	bufID, ok := result.(float64)
	if !ok {
		return fmt.Errorf("invalid buffer ID")
	}

	// Split content into lines
	var lines []string
	start := 0
	for i, char := range newContent {
		if char == '\n' {
			lines = append(lines, newContent[start:i])
			start = i + 1
		}
	}
	if start < len(newContent) {
		lines = append(lines, newContent[start:])
	}

	_, err = c.Call("nvim_buf_set_lines", int(bufID), 0, -1, true, lines)
	return err
}

func (c *RPCClient) GetFiletype() (string, error) {
	result, err := c.Call("nvim_buf_get_option", 0, "filetype")
	if err != nil {
		return "", err
	}

	filetype, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("invalid filetype")
	}

	return filetype, nil
}
