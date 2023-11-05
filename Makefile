smallchat: main.go
	go build -o smallchat main.go

.PHONY: clean
clean:
	rm -f smallchat