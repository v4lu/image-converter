# Image Conversion Service

This is a simple web service that converts images to various formats (AVIF, WebP, JPEG, PNG) using ImageMagick. The service is built with Go and can be run as a Docker container.

## Features

- Convert images to AVIF, WebP, JPEG, or PNG formats
- Automatically optimize images for size and quality
- Simple HTTP API for easy integration
- (On `convert-upload/aws` branch) Upload converted images to AWS S3 and return the URL

## Prerequisites

- Docker
- curl (for testing)
- (On `convert-upload/aws` branch) AWS account and credentials

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/v4lu/image-conversion-service.git
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

- Supported input formats: Any format supported by ImageMagick (common formats like JPEG, PNG, WebP, AVIF, HEIC, etc.)

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

## AWS S3 Upload Functionality

On the `convert-upload/aws` branch, this service includes functionality to upload the converted image to AWS S3 and return the URL of the uploaded image.

To use this feature:

1. Switch to the `convert-upload/aws` branch:
   ```
   git checkout convert-upload/aws
   ```

2. Set up your AWS credentials and S3 bucket information in your environment or Docker configuration.

3. Build and run the Docker container as described above.

4. The API will now return an S3 URL instead of the image file. For example:
   ```
   curl -X POST -F "image=@/path/to/your/image.jpg" "http://localhost:8080/convert?format=avif"
   ```
   This will return a URL like: `https://your-bucket-name.s3.your-region.amazonaws.com/uuid-filename.avif`

Note: Make sure to properly configure your AWS credentials and S3 bucket permissions for this functionality to work.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.