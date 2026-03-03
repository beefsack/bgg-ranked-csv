FROM golang:1.26

ENV GIT_CREDENTIAL_USERNAME=""
ENV GIT_CREDENTIAL_PASSWORD=""
ENV GIT_USER_NAME=""
ENV GIT_USER_EMAIL=""
ENV BGG_USERNAME=""
ENV BGG_PASSWORD=""

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /usr/local/bin/bgg-ranked-csv

COPY bgg-ranked-csv.sh /usr/local/bin/bgg-ranked-csv.sh

CMD [ "bgg-ranked-csv.sh" ]
