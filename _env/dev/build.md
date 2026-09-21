# Commands to build images

# Golang (platform) image - dev
```bash
docker buildx build ./_env/dev -f ./_env/dev/platform/Dockerfile -t akha_platform.dev && \
  docker tag akha_platform.dev akha_platform.dev
```