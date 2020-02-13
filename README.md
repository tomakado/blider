# Blider
Tool for scheduled background wallpaper changing in KDE Plasma and GNOME. Now uses pictures from simpledesktops.com

## Installation

Blider requires Go 1.12+ to build.

```shell script
git clone https://github.com/ildarkarymoff/blider
cd blider
go build
```

## Usage

```shell script
./blider
```

or

```shell script
./blider --config=<config_path>
```

```shell script
./blider -h  
Usage of ./blider:
  -config string
    	path to JSON file with configuration (default "$HOME/.blider/config.json")

```

## Configuration

Blider can be configured with passed JSON config file. By default it's located in `$HOME/.blider/config.json`, but you can pass file in other location by specifying `config` argument.

```json
{
  "local_storage_path": "/home/<username>/.blider/images", 
  "local_storage_limit": 100,
  "db_path": "/home/<username>/.blider/blider.sqlite",
  "max_fetch_pages": 10
}
```

## Project status

Blider now is alpha and contains some ugly pieces of code. Also code is not properly covered by unit tests.

You can track roadmap and current state of project [on Trello board](https://trello.com/b/fshibgJo/core).