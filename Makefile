all: clean binary

clean:
	rm -f alicloud-controller-manager

binary:
	GOOS=linux go build -o alicloud-controller-manager .

.PHONY: clean binary
