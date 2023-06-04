# **GRX - Reverse Proxy**

GRX is a reverse proxy built in Go, designed with the purpose of understanding the functioning of a reverse proxy and its components. The architecture of GRX is inspired by [NGINX](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/), with its focus on multiple processes managed by a master process.

## Key Features

* Uses a YAML file as the base for configuration.
* Allows configuration of all services.
* Defines the format of the header that the proxy should use to communicate the origin of the request (forwarded or x-forwarded).
* Offers options to choose the load balancing methodology per service.
* Provides basic configurations for connections between services and the proxy.

## Testing GRX

If you want to test GRX to learn, you can do so by following these simple steps:

1. Clone the repository to your local machine.

```bash
gh repo clone MAD-py/grx
```

2. Create a file called grx.yml in the root of the repository. In this file, you can perform all the necessary configurations. Below is an example of how its structure should be:

```yaml
servers:
  - name: Backend 1
    listen: 127.0.0.1:8000
    forward: 127.0.0.1:8001 # No load balancer
    connection:
      timeout: 40 # seconds
      concurrent: 1000
    header:
      forwarded: # enum: forwarded or x-forwarded
        id: toABfqD1egNrS
  - name: Backend 2
    listen: 127.0.0.1:8002
    forward: # Load balancer Round Robin
      - 127.0.0.1:8003
      - 127.0.0.1:8004
    connection:
      timeout: 40 # seconds
      concurrent: 1000
    header: x-forwarded # enum: forwarded or x-forwarded
  - name: Backend 3
    listen: 127.0.0.1:8005
    forward: # Load balancer Weight Round Robin
      - addres: 127.0.0.1:8006
        weight: 2
      - addres: 127.0.0.1:8007
        weight: 4
    connection:
      timeout: 40 # seconds
      concurrent: 1000
    header: x-forwarded # enum: forwarded or x-forwarded
  - name: Files # serve static files
    listen: 127.0.0.1:8008
    serve: /home/user/website
    connection:
      concurrent: 1000
```

3. From the repository's root, execute the following command to start the proxy:

```bash
go run ./cmd/main.go
```

If you want to run GRX with a configuration file in a different path or with a different name, you can do so by executing the same command in the following way:

```bash
go run ./cmd/main.go --file /custom/path/myfile.yml
```

Replace /custom/path/myfile.yaml with the path and name of the configuration file you want to use. This will allow GRX to load and use the specific configuration file you have indicated.

Remember that the configuration file must be in YAML format and contain the necessary configurations for services, headers, load balancing, and other relevant options.

## **License**

This project is distributed under the **MIT** license. Feel free to use and modify it according to your needs.