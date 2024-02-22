build-doc:
	go install github.com/instill-ai/component/tools/compogen@latest
gen-doc:
	go generate ./...
