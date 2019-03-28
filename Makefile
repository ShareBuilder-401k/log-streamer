ZIP_FILE := infrastructure/lambda/lambda.zip
OUT_FILE := main

clean: 
	rm -rf ${OUT_FILE} && \
	rm -rf ${ZIP_FILE}
	
build:
	GOARCH=amd64 \
	GOOS=linux \
	CGO_ENABLED=0 \
	go build -o ${OUT_FILE} main.go

set_version:
	echo ${BUILD_VERSION} > version

zip:
	zip ${ZIP_FILE} main version

package: clean build set_version zip