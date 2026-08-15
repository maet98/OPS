# This repository is to download all the image of a anime website and create a pdf file

## Setup

```bash
cp .env.example .env
```

## How to run it

This will download only this chaper

```bash
go run main.go --chapter 90
```

This will download all chapters from 90 to 99

```bash
go run main.go --from 90 --to 99
```

