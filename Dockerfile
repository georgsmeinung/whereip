# Dockerfile References: https://docs.docker.com/engine/reference/builder/

############################
# STEP 1 build executable binary
############################
# Start from the latest golang base image
FROM golang:alpine AS builder

# Configuration Environment variables
ENV GO111MODULE=on

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR /app
COPY . .

# Fetch dependencies.# Using go get.
RUN go get -d -v

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Expose port 8080 to the outside world
EXPOSE 3000

# Configuration Environment variables
ENV WHIP_UNKCURR=XX
ENV WHIP_UNAVCUR=0
ENV WHIP_BASECUR=USD
ENV WHIP_BASELAT=-34.603333
ENV WHIP_BASELNG=-58.381667
ENV WHIP_INFOURL=https://api.ip2country.info/ip?%v
ENV WHIP_CINFURL=https://restcountries.eu/rest/v2/alpha/%v
ENV WHIP_CUEXKEY=fadee53564a2f8e3e95d7d8cc47d4c64
ENV WHIP_CUEXURL=http://data.fixer.io/api/latest?access_key=%v&symbols=%v,%v
ENV WHIP_CCODPAT=countryCode
ENV WHIP_CNAMPAT=name
ENV WHIP_CRATPAT=rates.%v 
ENV WHIP_CUCDPAT=currencies.0.code
ENV WHIP_LANGPAT=languages.#.name
ENV WHIP_TZONPAT=timezones
ENV WHIP_BLATPAT=latlng.0
ENV WHIP_BLNGPAT=latlng.1

# Install required CA certificates to establish Secured Communication (SSL/TLS)
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
# COPY ./mycert.crt /usr/local/share/ca-certificates/mycert.crt
RUN update-ca-certificates

# Command to run the executable
CMD ["/app/main"]