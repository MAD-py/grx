package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func Load(filePath string) (*Config, error) {
	data, err := deserialize(filePath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if serversData, ok := data["servers"]; ok {
		servers, err := loadServers(serversData)
		if err != nil {
			return nil, err
		}
		config.Servers = servers
	} else {
		return nil, errors.New("servers section not found")
	}
	return config, nil
}

func loadServers(serversData interface{}) ([]*Server, error) {
	if serversData, ok := serversData.([]interface{}); ok {
		servers := make([]*Server, 0, len(serversData))
		for i, serverData := range serversData {
			server, err := loadServer(serverData, i)
			if err != nil {
				return nil, err
			}
			servers = append(servers, server)
		}
		return servers, nil
	}
	return nil, errors.New("no servers configured")
}

func loadServer(serverData interface{}, index int) (*Server, error) {
	if serverData, ok := serverData.(map[string]interface{}); ok {
		name, err := loadServerName(serverData, index)
		if err != nil {
			return nil, err
		}
		listen, err := loadServerListen(serverData, name)
		if err != nil {
			return nil, err
		}
		forward, err := loadServerForward(serverData, name)
		if err != nil {
			return nil, err
		}
		timeout, maxConnection, err := loadServerConn(serverData, name)
		if err != nil {
			return nil, err
		}

		return &Server{
			Name:              name,
			ListenAddr:        listen,
			PatternAddr:       forward,
			TimeoutPerRequest: timeout,
			MaxConnections:    maxConnection,
		}, nil
	}
	return nil, fmt.Errorf("wrong server %d configuration", index)
}

func loadServerName(serverData map[string]interface{}, index int) (string, error) {
	if name, ok := serverData["name"]; ok {
		if name, ok := name.(string); ok {
			return name, nil
		}
		return "", fmt.Errorf("name of server %d must be a string", index)
	}
	return fmt.Sprintf("server %d", index), nil
}

func loadServerListen(serverData map[string]interface{}, name string) (string, error) {
	if listen, ok := serverData["listen"]; ok {
		if listen, ok := listen.(string); ok {
			return listen, nil
		}
		return "", fmt.Errorf("listen of %s must be a string", name)
	}
	return "", fmt.Errorf("%s must have a listen", name)
}

func loadServerForward(serverData map[string]interface{}, name string) (string, error) {
	if forward, ok := serverData["forward"]; ok {
		if forward, ok := forward.(string); ok {
			return forward, nil
		}
		return "", fmt.Errorf("forward of %s must be a string", name)
	}
	return "", fmt.Errorf("%s must have a forward", name)
}

func loadServerConn(serverData map[string]interface{}, name string) (time.Duration, int, error) {
	var timeout time.Duration = 0
	var maxConnections int = 1000

	if conn, ok := serverData["connection"]; ok {
		if conn, ok := conn.(map[string]interface{}); ok {
			if t, ok := conn["timeout"]; ok {
				if t, ok := t.(int); ok {
					timeout = time.Duration(t)
				} else {
					return 0, 0, fmt.Errorf("timeout of %s must be an int", name)
				}
			}

			if mc, ok := conn["concurrent"]; ok {
				if mc, ok := mc.(int); ok {
					maxConnections = mc
				} else {
					return 0, 0, fmt.Errorf("concurrent of %s must be an int", name)
				}
			}
		} else {
			return 0, 0, fmt.Errorf("wrong %s configuration", name)
		}
	}
	return timeout, maxConnections, nil
}

func deserialize(filePath string) (map[string]interface{}, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	err = yaml.Unmarshal(fileData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
