normal:
	go build -o gotube main.go

clean:
	rm gotube

install:
	mkdir -p ~/.cache/gotube/thumbnails
	cp gotube ~/.local/bin/
	cp mpv/gotube.lua ~/.local/bin/

full: normal install clean
