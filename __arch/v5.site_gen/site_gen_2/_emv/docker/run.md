# run

run from the `site_gen_2` directory

## build containers

### dev - go, tailwindcss
```
docker build -t main_site-dev _emv/docker/dev
```

### site-html - nginx
```
docker build -t main_site-preview _emv/docker/site-html
```

## run

### css
```
docker run --rm -v "$(pwd):/src" -v main_site-go:/go main_site-dev \
  tailwindcss -i web/styles/main.css -o web/static/css/style.css --minify
```

### static site
```
docker run --rm -v "$(pwd):/src" -v main_site-go:/go main_site-dev \
  go run . --out deploy/static
```

### preview on port 8080
```
docker run -d --name main_site-preview -p 8080:80 \
  -v "$(pwd)/deploy/static:/var/www/sites/main:ro" main_site-preview
```

### stop preview
```
docker rm -f main_site-preview
```
