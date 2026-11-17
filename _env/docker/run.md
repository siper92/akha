# run config

# build docker
docker build -t main_site-preview .

# run docker
docker run -d --name main_site-preview -p 8080:80 \
-v "$(pwd)/__arch/v5.site_gen/site_gen_2/deploy/static:/var/www/sites/main:ro" main_site-preview
