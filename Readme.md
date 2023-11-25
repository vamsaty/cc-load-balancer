# Write Your Own Load Balancer

## Description
The challenge was to use build your own Application Load Balancer. Supports -
1. Distribute load to >= 2 servers
2. Health checks the servers
3. Handle server going offline
4. Handle server coming back online

## Usage
Steps to build the binary -
```
go build -o cc-load-balancer .
```
#### start a load balancer instance
```
./cc-load-balancer -lb -lb-endpoints ":8080,:8081,:8082" -port 80
```
#### start a backend server instance
```
./cc-load-balancer -port 8080
```

## Flags
| Flag | Description                                                                     | Default |
|------|---------------------------------------------------------------------------------| --- |
| -lb  | Indicates a loadbalancer instance, if false indicates a backend server instance | false |
| -lb-endpoints | Comma separated list of endpoints to load balance                               | "" |
| -port | Port to run the server on (this can be load balancer or backend server)         | 80 |
