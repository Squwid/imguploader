# Image Uploader

Image Uploader is a simple Go program that allows a folder to act as an instant
upload portal to a cloud storage bucket, and prints the URL of the uploaded
image, and deletes the image.

The general use for this is to quickly upload screenshots when writing
documentation or sharing images with friends, without re-inventing an entire
screen capture tool.

The tool is designed to run on startup and watch a single directory. When the 
command is run for the first tine, it will scan the directory and upload
images that do not already exist and delete them.

## Prerequisites

Before running this program, make sure you have the following:

- Go installed on your machine
- Google Cloud Storage account and project set up
- Environment variables set:
	- `IMGUPLOADER_WATCH_PATH`: The path to the directory to watch for new files
	- `IMGUPLOADER_BUCKET`: The name of the Google Cloud Storage bucket to upload files to
	- `IMGUPLOADER_URL`: The URL of the cloud storage bucket

## Installation

1. Clone the repository:
2. Build the executable:

```sh
go build -o imguploader
```

## Usage

1. Set the required environment variables:

```sh
export IMGUPLOADER_WATCH_PATH=/path/to/watch
export IMGUPLOADER_BUCKET=your-bucket-name
export IMGUPLOADER_URL=https://public-bucket-url # (optional)
```

2. Run the program:

```sh
./imguploader
```

3. The program will start watching the specified directory for new PNG and JPEG
files. When a new file is created, it will be uploaded to the specified cloud
storage bucket. The local file will be removed after successful upload.
