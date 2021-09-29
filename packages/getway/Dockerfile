FROM golang:1.17-alpine

WORKDIR /app

COPY utils ./utils/

COPY getway/go.mod ./getway/
COPY getway/go.sum ./getway/
RUN cd getway && go mod download

COPY getway/src/*.go ./getway/

RUN cd getway && go build -o /getway .

COPY getway/config.yml ./
COPY getway/ports.yml ./

EXPOSE 9000

ENV PORT=9000

CMD [ "/getway" ]