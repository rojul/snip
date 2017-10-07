# Snip

## Setup

``` bash
# build the frontend
git clone https://github.com/rojul/snip-web.git
cd snip-web
make docker-build
cd ..

git clone https://github.com/rojul/snip.git
cd snip
# build and start the backend
make

# build only the Ash image
make runner-build image-build-ash
# or build all images
make runner-build image-build
```
