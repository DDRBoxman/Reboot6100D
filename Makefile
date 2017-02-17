all: docker

docker: main
	docker build -t="ddrboxman/reboot6100d" .

push:
	docker push ddrboxman/reboot6100d

main:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

clean:
	rm main
