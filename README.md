# Chat Application

## Introduction
This is a simple Go gRPC project that implements a chat application where a client can send messages to other clients in real-time. In this application, the connection will automatically disconnect the client when the session expires or when the inactivity period exceeds 10 minutes.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [References](#references)

## Prerequisites 
Before running this project, make sure you have the following installed on your system:
1. Go: [Download](https://golang.org/doc/install)

* Run the Executable file and check the installation using the below command in CMD:

```
go version
```

2. Protocol Buffers (protoc): [Download ](https://developers.google.com/protocol-buffers/docs/downloads)

* Choose the protoc according to your Operating System.

* Unzip the protoc zip file in any directory and give the path to the bin in the System's
Environment Variables : 

`$(pwd)/bin`

* To check the installation of protoc Compiler run the following command in CMD :

```
protoc --version
```

3. gRPC Go: Install using `go get -u google.golang.org/grpc`

4. Install Docker Desktop on your system from https://www.docker.com/products/docker-desktop/.
## Usage
1. Clone the repository:
```bash
   git clone -b chat-application-gRPC-golang https://github.com/zversal-ecom/z-Skeleton.git

   cd z-Skeleton/backend/grpc-go-blueprint/chat app
```

## Command to initalize the Project

1. Now run the following command to import or download the required dependencies:

```bash
go mod tidy
```

## Run the Server

1. cd server

```bash
go run main.go 
```

We can pass the session time and port of the server at run time by this command :

```bash
go run main.go -port 7000 -session 2
```

2. create the executable file of the server.go file by this command :

```bash
go build -o bin/server.exe server/main.go
```

* cd bin

open the `CMD` run the following 

```bash
server.exe -port 7000 -session 2
```

## Run the Client 

1. cd client

```bash
go run main.go
```
or we can pass the address of the server by passing `-addr` flag on command line as shown in example bellow:

```bash
go run main.go -addr "localhost:7000"
```

2. create the executable file of the client.go file by this command :

```bash
go build -o bin/client.exe client/main.go
```

cd bin

open the `CMD` run the following 

```bash
client.exe -addr "localhost:7000"
```
Now, you will be asked to provide your name and email address. After submitting both of these fields, you will be registered for the chat app. You can then send messages in real-time to any registered client on the server.


## Docker Deployment

In **.env** file provide PORT as 7000

Sample env file

```bash
PORT=7000
```
default port = 7000

To create a docker image:

```
docker build -t <image-name> .
```

Here, _image-name_ is the name of the image to be created and this named image can be seen on the Docker Desktop under **Images** section.

To run the image:

```bash
docker run --name chat-server -p 7000:7000 -ti <image-name>
```

After that the Chat Server will run in a Container at port `7000`

After that go to the root Project Directory and open CMD and run the following command to run the Chat Client 

1. cd client

```bash
go run main.go
```
or we can pass the address of the server by passing `-addr` flag on command line as shown in example bellow:

```bash
go run main.go -addr "localhost:7000"
```
After that we will see this type of output: 

![Alt text](<Screenshot (35).png>)

## Deployment

### Infrastructure setup

1. **Create a github OIDC identity**

To create OIDC identity for github actions

- Create github action role by running script [runner script](z-skeleton-role.sh) given in the scripts folder. Provide the .env file with values in the same folder. Sample env
- we need to set the environments dev/prod as per our deplyments stages on [base.env](https://github.com/zversal-ecom/z-Skeleton/blob/main/scripts/base.env) file.
- After that we need to rename enviroments file name as per [dev.env/prod.env] also we can configure value as per our project.

```
  aws iam create-open-id-connect-provider --url "https://token.actions.githubusercontent.com" --thumbprint-list  "6938fd4d98bab03faadb97b34396831e3780aea1" "1c58a3a8518e8759bf075b76b750d4f2df264fcd" --client-id-list "sts.amazonaws.com"
```

dev

```
PROJECT_NAME=z-Skeleton
REGION=ap-south-1
OIDC_PROVIDER_NAME=token.actions.githubusercontent.com
ACCOUNT_NUMBER=***
Environment=dev
```

prod

```
set PROJECT_NAME=z-Skeleton
set REGION=ap-south-1
set OIDC_PROVIDER_NAME=token.actions.githubusercontent.com
set ACCOUNT_NUMBER=811375541953
set Environment=prod
```

This is the [permission_file](https://github.com/zversal-ecom/z-Skeleton/blob/main/scripts/z-skeleton-gh-role.yaml) used to OIDC identity for github actions.
This is the permission file used to create the user stack
[-------------------Yet to be added---------------------]

2.  **Workflow** (reference: https://github.com/zversal-ecom/z-Skeleton/blob/main/.github/workflows/grpc-go-blueprint.yml)

- Add access key, secret key , external id, and role in the workflow file (get these credentials from aws stack).
- Mention working directory.
- Set env variables.

  - Define following Variables:

        - REGION
        - ECR_REPOSITORY ## ecr repositary name
        - ECR_CLUSTER ## ecs cluster name
        - PROJECT_NAME
        - ACCOUNT_ID
        - VPC_SUBNET_A
        - VPC_SUBNET_B
        - VPC_SUBNET_C
        - SECURITY_GROUP_A
        - SECURITY_GROUP_B
        - ECR_CONTAINER ##ecs task defination
        - CONTAINER_NAME ##ecs container name

- Install dependencies using: npm i.
- Install serverless framework using: npm i serverless nx.
- firstly we need to create ecr repositary on this steps [Login to Amazon ECR] than build docker image using (bash_script)[https://github.com/zversal-ecom/z-Skeleton/blob/main/backend/grpc-go-blueprint/deploy_ecr_script.sh] .
- After that using nx command checking lint, formatter, test cases, deploy ecr fargate.
- On the other steps we need to run ECS TASk by putting (VPC_SUBNET_A) env value for the the tasks.

3.  **Serverless** (reference:
    - https://github.com/zversal-ecom/z-Skeleton/blob/main/backend/grpc-go-blueprint/serverless.yml)

- Set name.

  - Define services name
  - set region
  - set stage
  - set ECR_CONTAINER
  - set ECR_CLUSTER
  - set ECS_CONTAINER_IMAGE_URI
  - set plugins

- aws ecr fargate configuration so we need to task defination for containers, cluster for run container, for networking we need to [vpc, subnets, security group] also we need add ecs execution role for ecs fargte.
