build:
	go build -o onboarding main.go

run:
	./onboarding create simple.yml

all: build run
