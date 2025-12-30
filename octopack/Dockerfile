# syntax=docker/dockerfile:1

FROM alpine:latest
WORKDIR /app
RUN ["sh", "-c", "echo 'print(\"Hello World\")' > /app/app.py"]
ENV PORT=8080 PYTHONUNBUFFERED=1
EXPOSE 8080/tcp
CMD ["python3", "app.py"]
