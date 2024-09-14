# Image Conversion Service

This is a simple web service that converts images to various formats (AVIF, WebP, JPEG, PNG) using ImageMagick. The service is built with Go and can be run as a Docker container.

## Features

- Convert images to AVIF, WebP, JPEG, or PNG formats
- Automatically optimize images for size and quality
- Simple HTTP API for easy integration

## Prerequisites

- Docker
- curl (for testing)

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/image-conversion-service.git
   cd image-conversion-service
   ```

2. Build the Docker image:
   ```
   docker build -t image-converter-server .
   ```

3. Run the Docker container:
   ```
   docker run -p 8080:8080 image-converter-server
   ```

The server will start and listen on port 8080.

## Usage

To convert an image, send a POST request to the `/convert` endpoint with the image file in the request body. You can specify the desired output format using the `format` query parameter.

### Supported Formats

- AVIF (default)
- WebP
- JPEG
- PNG

### API Endpoint

```
POST /convert?format=<desired_format>
```

### Examples

Here are some examples using curl to convert images:

1. Convert to AVIF (default):
   ```
   curl -X POST -F "image=@/path/to/your/image.jpg" http://localhost:8080/convert --output converted_image.avif
   ```

2. Convert to WebP:
   ```
   curl -X POST -F "image=@/path/to/your/image.jpg" "http://localhost:8080/convert?format=webp" --output converted_image.webp
   ```

3. Convert to JPEG:
   ```
   curl -X POST -F "image=@/path/to/your/image.png" "http://localhost:8080/convert?format=jpg" --output converted_image.jpg
   ```

4. Convert to PNG:
   ```
   curl -X POST -F "image=@/path/to/your/image.webp" "http://localhost:8080/convert?format=png" --output converted_image.png
   ```

## Error Handling

The service will return appropriate HTTP status codes and error messages if something goes wrong:

- 400 Bad Request: If the image file is missing or the request is malformed
- 405 Method Not Allowed: If a non-POST request is sent to the /convert endpoint
- 500 Internal Server Error: If there's an error during the conversion process

## Limitations

- Maximum file size: 10MB
- Supported input formats: Any format supported by ImageMagick (common formats like JPEG, PNG, WebP, AVIF, HEIC, etc.)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.