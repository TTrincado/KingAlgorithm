FROM golang:1.18

RUN apt-get update && \
    apt-get install -y iputils-ping && \
    rm -rf /var/lib/apt/lists/*  # Limpiar caché para optimizar la imagen

WORKDIR /app

COPY . .

CMD ["go", "run", "src/main.go"]
