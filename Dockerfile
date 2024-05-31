FROM golang:alpine as builder-1
WORKDIR /app/catchAllTheZips/cmd
COPY catchAllTheZips /app/catchAllTheZips/
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o catz


FROM golang:alpine as builder-2
WORKDIR /app/whatsTheTemperature/cmd
COPY whatsTheTemperature /app/whatsTheTemperature/
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wtt

FROM scratch
WORKDIR /app
COPY --from=builder-1 /app/catchAllTheZips/cmd/catz .
COPY --from=builder-2 /app/whatsTheTemperature/cmd/wtt .
RUN cd ../../
#CMD ["./catz", "&", "./wtt"]