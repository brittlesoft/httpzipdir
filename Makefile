OUTDIR=build/

build: *.go
	go build -o ${OUTDIR}

run:
	go run


test:
	go test
