FROM znly/protoc:0.3.0

WORKDIR /app
COPY ./tools/protoc/entrypoint.sh entrypoint.sh
RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
