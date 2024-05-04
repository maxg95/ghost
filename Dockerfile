FROM golang:1.20
-
WORKDIR /app

RUN apt-get update && apt-get install -y \
    libx11-dev \
    libasound2-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libgl1-mesa-dev \
    libxxf86vm-dev

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app .

EXPOSE 8080

CMD ["./app"]
