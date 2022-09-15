#### How to build rathena image

##### TL;DR do not build rathena image use ready one https://hub.docker.com/r/rathena/rathena-prere

Unfortunately I was too lazy to check if rathena already have docker image and built my own using instructions for ubuntu. Copy these files to rathena source directory and run `nerdctl build -t my-rathena .` or `docker build -t my-rathena .`

Anyway I hope image can be easly replaced with rathena/rathena image because binaries reside under `/rathena/` directory so deployment and all other things must not change drastically