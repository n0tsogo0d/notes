# Notes
![preview.jpg](/preview.jpg)
**Notes** is a simple self-hosted markdown editor.

## Features
- **Attachments** upload through pasting/dropping files into the editor
- **Dark Mode** can be enabled/disabled with CTRL+SHIFT+D
- **Reader Mode** can be toggled with CTRL+D
- **Automatic file creation** through the URL. If you want to create a new note,
simply enter it in the URL, e.g. `http(s)://<host>/<your_path_plus_name>.md` and
it will be created for you.

## Usage
This project is meant to be run in a Docker container (but you can build the single binary yourself with `go build -o notes main.go`)

First build the container
`docker build -t <imagename> .`

Then simply run it
`docker run -d -p 8000:8000 -v /some/local/path:/data <imagename>`

That's it.


## License
I don't care, do whatever you want with it.