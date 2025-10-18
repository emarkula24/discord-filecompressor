FROM golang:1.25


RUN apt-get update && \
    apt-get install -y ffmpeg && \
    rm -rf /var/lib/apt/lists/*

    
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_DIR=
# RUN go build -v -o /usr/local/bin/app ./${SERVICE_DIR}
WORKDIR /usr/src/app/${SERVICE_DIR}

# CMD ["app"]
CMD ["go", "run", "main.go"]