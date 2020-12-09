build:
	go build -o onboarding main.go

run:
	cp onboarding ~/.jfrog/plugins/
	jfrog onboarding create simple.yml
	
all: build run
