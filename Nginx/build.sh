#podman build -f ./Dockerfile -t ghcr.io/memmaker/nginx:latest .
podman build --cap-add all -f ./Dockerfile -t ghcr.io/memmaker/nginx:latest .
