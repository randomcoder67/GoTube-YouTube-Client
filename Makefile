normal:
	go mod tidy
	go build -o gotube main.go

clean:
	rm gotube

install:
	mkdir -p ~/.cache/gotube/thumbnails
	cp gotube ~/.local/bin/
	cp mpv/gotube.lua ~/.local/bin/

uninstall:
	rm ~/.local/bin/gotube
	rm ~/.local/bin/gotube.lua

full: normal install clean
