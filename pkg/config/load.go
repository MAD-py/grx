package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func Load(filePath string) (Servers, error) {
	data, err := deserialize(filePath)
	if err != nil {
		return nil, err
	}

	var servers Servers
	if serversData, ok := data["servers"]; ok {
		servers, err = loadServers(serversData)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("servers section not found")
	}
	return servers, nil
}

func loadServers(serversData any) (Servers, error) {
	if serversData, ok := serversData.([]any); ok {
		servers := make([]any, 0, len(serversData))
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

func loadServer(serverData any, index int) (any, error) {
	if serverData, ok := serverData.(map[string]any); ok {
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

		forward, loadBalancer, err := loadServerForward(serverData, name)
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
			LoadBalancer:      loadBalancer,
			Forward:           forward,
			UseForwarded:      useForwarded,
			TimeoutPerRequest: timeout,
		}, nil
	}
	return nil, fmt.Errorf("wrong server %d configuration", index)
}

func loadServerName(serverData map[string]any, index int) (string, error) {
	if name, ok := serverData["name"]; ok {
		if name, ok := name.(string); ok {
			return name, nil
		}
		return "", fmt.Errorf("name of server %d must be a string", index)
	}
	return fmt.Sprintf("server %d", index), nil
}

func loadServerListen(serverData map[string]any, name string) (string, error) {
	if listen, ok := serverData["listen"]; ok {
		if listen, ok := listen.(string); ok {
			return listen, nil
		}
		return "", fmt.Errorf("listen of %s must be a string", name)
	}
	return "", fmt.Errorf("%s must have a listen", name)
}

func loadServerServe(serverData map[string]any, name string) (string, bool, error) {
	if serve, ok := serverData["serve"]; ok {
		if serve, ok := serve.(string); ok {
			return serve, true, nil
		}
		return "", false, fmt.Errorf("serve of %s must be a string", name)
	}
	return "", false, nil
}

func loadServerForward(serverData map[string]any, name string) ([]*Forward, LoadBalancer, error) {
	if forward, ok := serverData["forward"]; ok {
		if addr, ok := forward.(string); ok {
			return []*Forward{{Addr: addr}}, Base, nil
		}
		if forwards, ok := forward.([]any); ok {
			return loadServerLoadBalancer(forwards, name)
		}
		return nil, non, fmt.Errorf("forward of %s must be a string or array", name)
	}
	return nil, non, fmt.Errorf("%s must have a forward or serve", name)
}

func loadServerLoadBalancer(serverData []any, name string) ([]*Forward, LoadBalancer, error) {
	if _, ok := serverData[0].(string); ok {
		forwards := make([]*Forward, len(serverData))
		for i, forward := range serverData {
			if addr, ok := forward.(string); ok {
				forwards[i] = &Forward{Addr: addr}
			} else {
				return nil, non, fmt.Errorf("forward %s must be all of the same type", name)
			}
		}
		return forwards, RoundRobin, nil
	} else if _, ok := serverData[0].(map[string]any); ok {
		forwards := make([]*Forward, len(serverData))
		for i, v := range serverData {
			if f, ok := v.(map[string]any); ok {
				forward := &Forward{}
				if addr, ok := f["addres"].(string); ok {
					forward.Addr = addr
				} else {
					return nil, non, fmt.Errorf("the address of forward %s %d must be string", name, i)
				}
				if weight, ok := f["weight"].(int); ok {
					forward.Weight = uint8(weight)
				} else {
					return nil, non, fmt.Errorf("the Weight of forward %s %d must be integer", name, i)
				}
				forwards[i] = forward
			} else {
				return nil, non, fmt.Errorf("forward %s must be all of the same type", name)
			}
		}
		return forwards, WeightedRoundRobin, nil
	}
	return nil, non, fmt.Errorf("forward of %s must be a string array or dict array", name)
}

func loadServerHeader(serverData map[string]any, name string) (bool, string, error) {
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
		if header, ok := header.(map[string]any); ok {
			if forwarded, ok := header["forwarded"]; ok {
				if forwarded, ok := forwarded.(map[string]any); ok {
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

func loadServerConn(serverData map[string]any, name string) (time.Duration, int, error) {
	var timeout time.Duration = 40
	var maxConnections int = 1000

	if conn, ok := serverData["connection"]; ok {
		if conn, ok := conn.(map[string]any); ok {
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

func deserialize(filePath string) (map[string]any, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data := make(map[string]any)
	err = yaml.Unmarshal(fileData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
