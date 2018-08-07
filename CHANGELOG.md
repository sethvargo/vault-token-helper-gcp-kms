# CHANGELOG

## v0.2.0

- Don't base64 decode values when writing to disk - this saves the extra call
  and makes the file contents more portable instead of binary.
- Add Dockerfile and publish image

## v0.1.0

- Initial release
