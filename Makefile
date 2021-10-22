OUTDIR=build/

build:
	go build -o ${OUTDIR}

run:
	go run


test:
	go test
