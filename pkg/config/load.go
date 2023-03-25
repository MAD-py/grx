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

func loadServers(serversData interface{}) ([]interface{}, error) {
	if serversData, ok := serversData.([]interface{}); ok {
		servers := make([]interface{}, 0, len(serversData))
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

func loadServer(serverData interface{}, index int) (interface{}, error) {
	if serverData, ok := serverData.(map[string]interface{}); ok {
		name, err := loadServerName(serverData, index)
		if err != nil {
			return nil, err
		}
		listen, err := loadServerListen(serverData, name)
		if err != nil {
			return nil, err
		}

		timeout, maxConnection, err := loadServerConn(serverData, name)
		if err != nil {
			return nil, err
		}

		serve, ok, err := loadServerServe(serverData, name)
		if err != nil {
			return nil, err
		}

		if ok {
			return &StaticServer{
				Server: Server{
					Name:           name,
					ListenAddr:     listen,
					MaxConnections: maxConnection,
				},
				PathPrefix: serve,
			}, nil
		}

		forward, err := loadServerForward(serverData, name)
		if err != nil {
			return nil, err
		}

		useForwarded, id, err := loadServerHeader(serverData, name)
		if err != nil {
			return nil, err
		}

		return &ForwardServer{
			Server: Server{
				Name:           name,
				ListenAddr:     listen,
				MaxConnections: maxConnection,
			},
			ID:                id,
			PatternAddr:       forward,
			UseForwarded:      useForwarded,
			TimeoutPerRequest: timeout,
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

func loadServerServe(serverData map[string]interface{}, name string) (string, bool, error) {
	if serve, ok := serverData["serve"]; ok {
		if serve, ok := serve.(string); ok {
			return serve, true, nil
		}
		return "", false, fmt.Errorf("serve of %s must be a string", name)
	}
	return "", false, nil
}

func loadServerForward(serverData map[string]interface{}, name string) (string, error) {
	if forward, ok := serverData["forward"]; ok {
		if forward, ok := forward.(string); ok {
			return forward, nil
		}
		return "", fmt.Errorf("forward of %s must be a string", name)
	}
	return "", fmt.Errorf("%s must have a forward or serve", name)
}

func loadServerHeader(serverData map[string]interface{}, name string) (bool, string, error) {
	if header, ok := serverData["header"]; ok {
		if header, ok := header.(string); ok {
			if header == "forwarded" {
				return true, "", nil
			}
			if header == "x-forwarded" {
				return false, "", nil
			}
			return false, "", fmt.Errorf(
				"%s server its header value must be forwarded or x-forwarded", name,
			)
		}
		if header, ok := header.(map[string]interface{}); ok {
			if forwarded, ok := header["forwarded"]; ok {
				if forwarded, ok := forwarded.(map[string]interface{}); ok {
					if id, ok := forwarded["id"]; ok {
						if id, ok := id.(string); ok {
							return true, id, nil
						}
						return false, "", fmt.Errorf(
							"id field of the forwarded header of %s server must be a string", name,
						)
					}
					return false, "", fmt.Errorf(
						"forwarded header of the %s server must have the field id", name,
					)
				}
				return false, "", fmt.Errorf(
					"forwarded header of the %s server must have the field id", name,
				)
			}
			return false, "", fmt.Errorf(
				"%s server its header value must be forwarded or x-forwarded", name,
			)
		}
	}
	return true, "", nil
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
